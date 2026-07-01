package test

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/permission/model"
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

func TestCircularRoleInheritance(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "DirectCycle",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()

				roleA := &roleEntity.Role{ID: "role-a", Name: "admin"}
				roleB := &roleEntity.Role{ID: "role-b", Name: "editor"}

				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(roleB, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "admin").Return(roleA, nil)
				deps.Enforcer.On("AddGroupingPolicy", []interface{}{"editor", "admin", "global"}).Return(true, nil)

				err := uc.AddParentRole(context.Background(), "editor", "admin", "global")
				assert.NoError(t, err)

				deps.RoleRepo.AssertExpectations(t)
				deps.Enforcer.AssertExpectations(t)
			},
		},
		{
			name:     "IndirectCycle",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()

				roleA := &roleEntity.Role{ID: "role-a", Name: "superadmin"}
				roleC := &roleEntity.Role{ID: "role-c", Name: "moderator"}

				deps.RoleRepo.On("FindByName", mock.Anything, "superadmin").Return(roleA, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "moderator").Return(roleC, nil)
				deps.Enforcer.On("AddGroupingPolicy", []interface{}{"superadmin", "moderator", "global"}).Return(true, nil)

				err := uc.AddParentRole(context.Background(), "superadmin", "moderator", "global")
				assert.NoError(t, err)

				deps.RoleRepo.AssertExpectations(t)
				deps.Enforcer.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// ============================================================================
// 🔐 SQL INJECTION IN PERMISSION INPUTS
// ============================================================================

func TestGrantPermissionToRole_SQLInjectionInputs(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "in path",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"editor", "global", "/api/users'; DROP TABLE users; --", "GET"}).Return(true, nil)
				err := uc.GrantPermissionToRole(context.Background(), "editor", "/api/users'; DROP TABLE users; --", "GET", "global")
				assert.NoError(t, err)
				deps.Enforcer.AssertCalled(t, "AddPolicy", []interface{}{"editor", "global", "/api/users'; DROP TABLE users; --", "GET"})
			},
		},
		{
			name:     "in role name",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "admin' OR '1'='1").Return(nil, errors.New("record not found"))
				err := uc.GrantPermissionToRole(context.Background(), "admin' OR '1'='1", "/api/v1/users", "DELETE", "global")
				assert.Error(t, err)
				deps.Enforcer.AssertNotCalled(t, "AddPolicy", mock.Anything)
			},
		},
		{
			name:     "in method",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "viewer").Return(&roleEntity.Role{Name: "viewer"}, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"viewer", "global", "/api/reports", "GET; DROP TABLE policies; --"}).Return(true, nil)
				err := uc.GrantPermissionToRole(context.Background(), "viewer", "/api/reports", "GET; DROP TABLE policies; --", "global")
				assert.NoError(t, err)
				deps.Enforcer.AssertCalled(t, "AddPolicy", []interface{}{"viewer", "global", "/api/reports", "GET; DROP TABLE policies; --"})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// ============================================================================
// 🔐 CONCURRENT PERMISSION UPDATES
// ============================================================================

func TestPermissionConcurrentOperations(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "GrantPermissionConcurrentSameRole",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				var opCount int32

				roleName := "editor"
				deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{ID: "role-editor", Name: roleName}, nil).Maybe()
				deps.Enforcer.On("AddPolicy", mock.Anything).Run(func(args mock.Arguments) {
					atomic.AddInt32(&opCount, 1)
				}).Return(true, nil).Maybe()

				numGoroutines := 10
				var wg sync.WaitGroup
				errChan := make(chan error, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						path := "/api/resource/" + string(rune('a'+idx))
						errChan <- uc.GrantPermissionToRole(context.Background(), "editor", path, "GET", "global")
					}(i)
				}

				wg.Wait()
				close(errChan)
				for err := range errChan {
					assert.NoError(t, err)
				}

				assert.Equal(t, int32(10), atomic.LoadInt32(&opCount))
			},
		},
		{
			name:     "RevokePermissionConcurrent",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				var opCount int32

				roleName := "editor"
				deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{ID: "role-editor", Name: roleName}, nil).Maybe()
				deps.Enforcer.On("RemovePolicy", mock.Anything).Run(func(args mock.Arguments) {
					atomic.AddInt32(&opCount, 1)
				}).Return(true, nil)

				numGoroutines := 5
				var wg sync.WaitGroup
				errChan := make(chan error, numGoroutines)

				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						path := "/api/resource/" + string(rune('a'+idx))
						errChan <- uc.RevokePermissionFromRole(context.Background(), "editor", path, "DELETE", "global")
					}(i)
				}

				wg.Wait()
				close(errChan)
				for err := range errChan {
					assert.NoError(t, err)
				}

				assert.Equal(t, int32(5), atomic.LoadInt32(&opCount))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// ============================================================================
// 🔐 EDGE CASE: EMPTY AND SPECIAL VALUES
// ============================================================================

func TestGrantPermissionToRole_EdgeInputs(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "EmptyPath",
			category: "edge",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.GrantPermissionToRole(context.Background(), "editor", "", "GET", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required")
			},
		},
		{
			name:     "WildcardPath",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "admin").Return(&roleEntity.Role{ID: "role-admin", Name: "admin"}, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"admin", "global", "/api/*", "*"}).Return(true, nil)
				err := uc.GrantPermissionToRole(context.Background(), "admin", "/api/*", "*", "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "UnicodePath",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{ID: "role-editor", Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"editor", "global", "/api/用户/管理", "GET"}).Return(true, nil)
				err := uc.GrantPermissionToRole(context.Background(), "editor", "/api/用户/管理", "GET", "global")
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// ============================================================================
// 🔐 ENFORCER FAILURE HANDLING
// ============================================================================

func TestPermissionPolicyFailureHandling(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "GrantEnforcerConnectionError",
			category: "error",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{ID: "role-editor", Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"editor", "global", "/api/test", "GET"}).Return(false, errors.New("connection refused"))
				err := uc.GrantPermissionToRole(context.Background(), "editor", "/api/test", "GET", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "connection refused")
			},
		},
		{
			name:     "RevokePolicyNotExists",
			category: "error",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "viewer").Return(&roleEntity.Role{ID: "role-viewer", Name: "viewer"}, nil)
				deps.Enforcer.On("RemovePolicy", []interface{}{"viewer", "global", "/api/nonexistent", "DELETE"}).Return(false, nil)
				err := uc.RevokePermissionFromRole(context.Background(), "viewer", "/api/nonexistent", "DELETE", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "policy to revoke not found")
			},
		},
		{
			name:     "GrantUpdateExistingPolicy",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{ID: "role-editor", Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"editor", "global", "/api/users", "GET"}).Return(false, nil)
				err := uc.GrantPermissionToRole(context.Background(), "editor", "/api/users", "GET", "global")
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestConcurrentBatchPermissionCheck(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "ConcurrentBatchCheck",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Maybe()

				userID := "user-concurrent"
				numGoroutines := 10
				itemsPerCheck := 5
				var wg sync.WaitGroup
				var successCount int32
				for i := 0; i < numGoroutines; i++ {
					wg.Add(1)
					go func(idx int) {
						defer wg.Done()
						items := make([]model.PermissionCheckItem, itemsPerCheck)
						itemsKey := idx * itemsPerCheck
						for j := 0; j < itemsPerCheck; j++ {
							items[j] = model.PermissionCheckItem{Resource: fmt.Sprintf("/api/resource-%d", itemsKey+j), Action: "GET"}
						}
						results, err := uc.BatchCheckPermission(context.Background(), userID, items)
						if err == nil && len(results) == itemsPerCheck {
							atomic.AddInt32(&successCount, 1)
						}
					}(i)
				}

				wg.Wait()
				assert.Equal(t, int32(10), successCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
