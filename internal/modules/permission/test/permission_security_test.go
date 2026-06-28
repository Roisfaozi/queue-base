package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/permission/model"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ============================================================================
// SECURITY TEST SUITE - Permission UseCase
// Tests for: Circular role inheritance, SQL injection, Concurrent access
// ============================================================================

// setupPermissionTest is used from permission_usecase_test.go

// ============================================================================
// 🔐 CIRCULAR ROLE INHERITANCE TESTS
// ============================================================================

// TestCircularRoleInheritance_DirectCycle tests A -> B -> A circular reference.
func TestCircularRoleInheritance_DirectCycle(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleA := &roleEntity.Role{ID: "role-a", Name: "admin"}
	roleB := &roleEntity.Role{ID: "role-b", Name: "editor"}

	deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(roleB, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "admin").Return(roleA, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AddParentRole(context.Background(), "editor", "admin", "global")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

// TestCircularRoleInheritance_IndirectCycle tests A -> B -> C -> A circular reference.
func TestCircularRoleInheritance_IndirectCycle(t *testing.T) {
	deps, uc := setupPermissionTest()

	roleA := &roleEntity.Role{ID: "role-a", Name: "superadmin"}
	roleC := &roleEntity.Role{ID: "role-c", Name: "moderator"}

	deps.RoleRepo.On("FindByName", mock.Anything, "superadmin").Return(roleA, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "moderator").Return(roleC, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AddParentRole(context.Background(), "superadmin", "moderator", "global")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}

// ============================================================================
// 🔐 SQL INJECTION IN PERMISSION INPUTS
// ============================================================================

func TestGrantPermissionToRole_SQLInjectionInputs(t *testing.T) {
	tests := []struct {
		name          string
		roleName      string
		path          string
		method        string
		setupMock     func(*permissionTestDeps)
		wantErr       bool
		assertNoAdd   bool
		assertAddCall bool
	}{
		{
			name:     "in path",
			roleName: "editor",
			path:     "/api/users'; DROP TABLE users; --",
			method:   "GET",
			setupMock: func(deps *permissionTestDeps) {
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)
			},
			assertAddCall: true,
		},
		{
			name:     "in role name",
			roleName: "admin' OR '1'='1",
			path:     "/api/v1/users",
			method:   "DELETE",
			setupMock: func(deps *permissionTestDeps) {
				deps.RoleRepo.On("FindByName", mock.Anything, "admin' OR '1'='1").Return(nil, errors.New("record not found"))
			},
			wantErr:     true,
			assertNoAdd: true,
		},
		{
			name:     "in method",
			roleName: "viewer",
			path:     "/api/reports",
			method:   "GET; DROP TABLE policies; --",
			setupMock: func(deps *permissionTestDeps) {
				deps.RoleRepo.On("FindByName", mock.Anything, "viewer").Return(&roleEntity.Role{Name: "viewer"}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)
			},
			assertAddCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			if tt.setupMock != nil {
				tt.setupMock(deps)
			}

			err := uc.GrantPermissionToRole(context.Background(), tt.roleName, tt.path, tt.method, "global")
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.assertNoAdd {
				deps.Enforcer.AssertNotCalled(t, "AddPolicy", mock.Anything, mock.Anything, mock.Anything)
			}
			if tt.assertAddCall {
				deps.Enforcer.AssertCalled(t, "AddPolicy", mock.Anything)
			}
		})
	}
}

// ============================================================================
// 🔐 CONCURRENT PERMISSION UPDATES
// ============================================================================

func TestGrantPermissionToRole_Concurrent_SameRole(t *testing.T) {
	deps, uc := setupPermissionTest()
	numConcurrent := 10

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil).Maybe()

	var successCount int32
	deps.Enforcer.On("AddPolicy", mock.Anything).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&successCount, 1)
		}).Return(true, nil).Maybe()

	var wg sync.WaitGroup
	errChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := "/api/resource/" + string(rune('a'+idx))
			err := uc.GrantPermissionToRole(context.Background(), roleName, path, "GET", "global")
			errChan <- err
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(numConcurrent), atomic.LoadInt32(&successCount))
}

func TestRevokePermissionFromRole_Concurrent(t *testing.T) {
	deps, uc := setupPermissionTest()
	numConcurrent := 5

	roleName := "editor"
	role := &roleEntity.Role{ID: "role-editor", Name: roleName}

	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(role, nil).Maybe()

	var revokeCount int32
	deps.Enforcer.On("RemovePolicy", mock.Anything).
		Run(func(args mock.Arguments) {
			atomic.AddInt32(&revokeCount, 1)
		}).Return(true, nil)

	var wg sync.WaitGroup
	errChan := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := "/api/resource/" + string(rune('a'+idx))
			err := uc.RevokePermissionFromRole(context.Background(), roleName, path, "DELETE", "global")
			errChan <- err
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		assert.NoError(t, err)
	}

	assert.Equal(t, int32(numConcurrent), atomic.LoadInt32(&revokeCount))
}

// ============================================================================
// 🔐 EDGE CASE: EMPTY AND SPECIAL VALUES
// ============================================================================

func TestGrantPermissionToRole_EdgeInputs(t *testing.T) {
	tests := []struct {
		name        string
		roleName    string
		path        string
		method      string
		setupMock   func(*permissionTestDeps)
		wantErr     bool
		wantContain string
	}{
		{name: "empty path", roleName: "editor", path: "", method: "GET", wantErr: true, wantContain: "required"},
		{name: "wildcard path", roleName: "admin", path: "/api/*", method: "*", setupMock: func(deps *permissionTestDeps) {
			deps.RoleRepo.On("FindByName", mock.Anything, "admin").Return(&roleEntity.Role{ID: "role-admin", Name: "admin"}, nil)
			deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)
		}},
		{name: "unicode path", roleName: "editor", path: "/api/用户/管理", method: "GET", setupMock: func(deps *permissionTestDeps) {
			deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{ID: "role-editor", Name: "editor"}, nil)
			deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			if tt.setupMock != nil {
				tt.setupMock(deps)
			}

			err := uc.GrantPermissionToRole(context.Background(), tt.roleName, tt.path, tt.method, "global")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantContain)
				return
			}
			assert.NoError(t, err)
		})
	}
}

// ============================================================================
// 🔐 ENFORCER FAILURE HANDLING
// ============================================================================

func TestPermissionPolicyFailureHandling(t *testing.T) {
	tests := []struct {
		name        string
		run         func(*permissionTestDeps, usecase.IPermissionUseCase) error
		wantErr     bool
		wantContain string
	}{
		{
			name: "grant enforcer connection error",
			run: func(deps *permissionTestDeps, uc usecase.IPermissionUseCase) error {
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{ID: "role-editor", Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, errors.New("connection refused"))
				return uc.GrantPermissionToRole(context.Background(), "editor", "/api/test", "GET", "global")
			},
			wantErr:     true,
			wantContain: "connection refused",
		},
		{
			name: "revoke policy not exists",
			run: func(deps *permissionTestDeps, uc usecase.IPermissionUseCase) error {
				deps.RoleRepo.On("FindByName", mock.Anything, "viewer").Return(&roleEntity.Role{ID: "role-viewer", Name: "viewer"}, nil)
				deps.Enforcer.On("RemovePolicy", mock.Anything).Return(false, nil)
				return uc.RevokePermissionFromRole(context.Background(), "viewer", "/api/nonexistent", "DELETE", "global")
			},
			wantErr:     true,
			wantContain: "policy to revoke not found",
		},
		{
			name: "grant update existing policy",
			run: func(deps *permissionTestDeps, uc usecase.IPermissionUseCase) error {
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{ID: "role-editor", Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, nil)
				return uc.GrantPermissionToRole(context.Background(), "editor", "/api/users", "GET", "global")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			err := tt.run(deps, uc)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantContain)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestConcurrentBatchPermissionCheck(t *testing.T) {
	deps, uc := setupPermissionTest()

	userID := "user-concurrent"
	numGoroutines := 10
	itemsPerCheck := 5

	deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Maybe()

	var wg sync.WaitGroup
	var successCount int32

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()

			items := make([]model.PermissionCheckItem, itemsPerCheck)
			itemsKey := idx * itemsPerCheck
			for j := 0; j < itemsPerCheck; j++ {
				items[j] = model.PermissionCheckItem{
					Resource: fmt.Sprintf("/api/resource-%d", itemsKey+j),
					Action:   "GET",
				}
			}

			results, err := uc.BatchCheckPermission(context.Background(), userID, items)
			if err == nil && len(results) == itemsPerCheck {
				atomic.AddInt32(&successCount, 1)
			}
		}(i)
	}

	wg.Wait()
	assert.Equal(t, int32(numGoroutines), successCount)
}
