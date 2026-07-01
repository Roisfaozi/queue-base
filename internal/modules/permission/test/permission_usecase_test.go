package test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	accessMocks "github.com/Roisfaozi/queue-base/internal/modules/access/test/mocks"
	auditMocks "github.com/Roisfaozi/queue-base/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/model"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/usecase"
	roleEntity "github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	roleMocks "github.com/Roisfaozi/queue-base/internal/modules/role/test/mocks"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type permissionTestDeps struct {
	Enforcer   *mocks.MockIEnforcer
	RoleRepo   *roleMocks.MockRoleRepository
	UserRepo   *userMocks.MockUserRepository
	AccessRepo *accessMocks.MockAccessRepository
	AuditUC    *auditMocks.MockAuditUseCase
}

func setupPermissionTest() (*permissionTestDeps, usecase.IPermissionUseCase) {
	deps := &permissionTestDeps{
		Enforcer:   new(mocks.MockIEnforcer),
		RoleRepo:   new(roleMocks.MockRoleRepository),
		UserRepo:   new(userMocks.MockUserRepository),
		AccessRepo: new(accessMocks.MockAccessRepository),
		AuditUC:    new(auditMocks.MockAuditUseCase),
	}

	// Default behavior for enforcer with context to return itself
	deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)

	log := logrus.New()
	log.SetOutput(io.Discard)
	uc := usecase.NewPermissionUseCase(deps.Enforcer, log, deps.RoleRepo, deps.UserRepo, deps.AccessRepo, deps.AuditUC)
	return deps, uc
}

func TestAssignRoleToUser(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(&userEntity.User{ID: "user123"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

				err := uc.AssignRoleToUser(ctx, "user123", "editor", "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(nil, gorm.ErrRecordNotFound)

				err := uc.AssignRoleToUser(ctx, "user123", "editor", "global")
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_UserRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(nil, errors.New("db error"))

				err := uc.AssignRoleToUser(ctx, "user123", "editor", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_RoleNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(&userEntity.User{ID: "user123"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "non_existent_role").Return(nil, gorm.ErrRecordNotFound)

				err := uc.AssignRoleToUser(ctx, "user123", "non_existent_role", "global")
				assert.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_RoleRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(&userEntity.User{ID: "user123"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(nil, errors.New("db error"))

				err := uc.AssignRoleToUser(ctx, "user123", "editor", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_EnforcerRemoveError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(&userEntity.User{ID: "user123"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(false, errors.New("casbin error"))

				err := uc.AssignRoleToUser(ctx, "user123", "editor", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_EnforcerAddError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.UserRepo.On("FindByID", mock.Anything, "user123").Return(&userEntity.User{ID: "user123"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

				err := uc.AssignRoleToUser(ctx, "user123", "editor", "global")
				assert.Error(t, err)
				assert.Equal(t, "casbin error", err.Error())
			},
		},
		{
			name:     "Negative_EmptyUserID",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.AssignRoleToUser(context.Background(), "", "role", "global")
				assert.Error(t, err)
				assert.Equal(t, "userID and role are required", err.Error())
			},
		},
		{
			name:     "Negative_EmptyRole",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.AssignRoleToUser(context.Background(), "user", "", "global")
				assert.Error(t, err)
				assert.Equal(t, "userID and role are required", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGrantPermissionToRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

				err := uc.GrantPermissionToRole(ctx, "editor", "/api/v1/articles", "POST", "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_RoleNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "non_existent_role").Return(nil, gorm.ErrRecordNotFound)

				err := uc.GrantPermissionToRole(ctx, "non_existent_role", "/api/v1/articles", "POST", "global")
				assert.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_RoleRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(nil, errors.New("db error"))

				err := uc.GrantPermissionToRole(ctx, "role", "/path", "POST", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, errors.New("casbin error"))

				err := uc.GrantPermissionToRole(ctx, "role", "/path", "POST", "global")
				assert.Error(t, err)
				assert.Equal(t, "casbin error", err.Error())
			},
		},
		{
			name:     "Negative_EmptyRole",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.GrantPermissionToRole(context.Background(), "", "path", "GET", "global")
				assert.Error(t, err)
				assert.Equal(t, "role, path, and method are required", err.Error())
			},
		},
		{
			name:     "Negative_EmptyPath",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.GrantPermissionToRole(context.Background(), "role", "", "GET", "global")
				assert.Error(t, err)
				assert.Equal(t, "role, path, and method are required", err.Error())
			},
		},
		{
			name:     "Negative_EmptyMethod",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.GrantPermissionToRole(context.Background(), "role", "path", "", "global")
				assert.Error(t, err)
				assert.Equal(t, "role, path, and method are required", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRevokePermissionFromRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.Enforcer.On("RemovePolicy", mock.Anything).Return(true, nil)

				err := uc.RevokePermissionFromRole(ctx, "editor", "/api/v1/articles", "DELETE", "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_RoleNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "non_existent_role").Return(nil, gorm.ErrRecordNotFound)

				err := uc.RevokePermissionFromRole(ctx, "non_existent_role", "/api/v1/articles", "DELETE", "global")
				assert.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_RoleRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(nil, errors.New("db error"))

				err := uc.RevokePermissionFromRole(ctx, "role", "/path", "DELETE", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)
				deps.Enforcer.On("RemovePolicy", mock.Anything).Return(false, errors.New("casbin error"))

				err := uc.RevokePermissionFromRole(ctx, "role", "/path", "DELETE", "global")
				assert.Error(t, err)
				assert.Equal(t, "casbin error", err.Error())
			},
		},
		{
			name:     "Negative_PolicyNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)
				deps.Enforcer.On("RemovePolicy", mock.Anything).Return(false, nil)

				err := uc.RevokePermissionFromRole(ctx, "role", "/path", "DELETE", "global")
				assert.Error(t, err)
				assert.Equal(t, "policy to revoke not found in specified domain", err.Error())
			},
		},
		{
			name:     "Negative_EmptyRole",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.RevokePermissionFromRole(context.Background(), "", "path", "GET", "global")
				assert.Error(t, err)
				assert.Equal(t, "role, path, and method are required", err.Error())
			},
		},
		{
			name:     "Negative_EmptyPath",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.RevokePermissionFromRole(context.Background(), "role", "", "GET", "global")
				assert.Error(t, err)
				assert.Equal(t, "role, path, and method are required", err.Error())
			},
		},
		{
			name:     "Negative_EmptyMethod",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.RevokePermissionFromRole(context.Background(), "role", "path", "", "global")
				assert.Error(t, err)
				assert.Equal(t, "role, path, and method are required", err.Error())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetAllPermissions(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				expectedPolicies := [][]string{{"role", "path", "GET"}}
				deps.Enforcer.On("GetPolicy").Return(expectedPolicies, nil).Once()

				policies, err := uc.GetAllPermissions(context.Background())
				assert.NoError(t, err)
				assert.Equal(t, expectedPolicies, policies)
			},
		},
		{
			name:     "Negative_Error",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.Enforcer.On("GetPolicy").Return(nil, errors.New("casbin error")).Once()

				policies, err := uc.GetAllPermissions(context.Background())
				assert.Error(t, err)
				assert.Nil(t, policies)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetPermissionsForRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				role := "admin"
				expectedPolicies := [][]string{{"admin", "path", "GET"}}
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{role}).Return(expectedPolicies, nil).Once()

				policies, err := uc.GetPermissionsForRole(context.Background(), role)
				assert.NoError(t, err)
				assert.Len(t, policies, 1)
			},
		},
		{
			name:     "Negative_Error",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				role := "admin"
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{role}).Return(nil, errors.New("casbin error")).Once()

				policies, err := uc.GetPermissionsForRole(context.Background(), role)
				assert.Error(t, err)
				assert.Nil(t, policies)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestUpdatePermission(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				oldP := []string{"role", "old", "GET"}
				newP := []string{"role", "new", "GET"}
				deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(true, nil).Once()

				updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
				assert.NoError(t, err)
				assert.True(t, updated)
			},
		},
		{
			name:     "Negative_EmptyOldInput",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				oldP := []string{}
				newP := []string{"a"}

				updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
				assert.Error(t, err)
				assert.False(t, updated)
			},
		},
		{
			name:     "Negative_EmptyNewInput",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				oldP := []string{"a"}
				newP := []string{}

				updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
				assert.Error(t, err)
				assert.False(t, updated)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				oldP := []string{"role", "old", "GET"}
				newP := []string{"role", "new", "GET"}
				deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(false, errors.New("casbin error")).Once()

				updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
				assert.Error(t, err)
				assert.False(t, updated)
			},
		},
		{
			name:     "Negative_PolicyNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				oldP := []string{"role", "old", "GET"}
				newP := []string{"role", "new", "GET"}
				deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(false, nil).Once()

				updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
				assert.Error(t, err)
				assert.False(t, updated)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestDeleteRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_ReloadsPolicyOutsideTransaction",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				roleName := "role:editor"

				deps.Enforcer.On("DeleteRole", roleName).Return(true, nil)
				deps.Enforcer.On("LoadPolicy").Return(nil)

				err := uc.DeleteRole(context.Background(), roleName)

				assert.NoError(t, err)
				deps.Enforcer.AssertCalled(t, "LoadPolicy")
			},
		},
		{
			name:     "Positive_DefersPolicyReloadInsideTransaction",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				roleName := "role:editor"

				deps.Enforcer.On("DeleteRole", roleName).Return(true, nil)

				db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
				assert.NoError(t, err)

				logger := logrus.New()
				logger.SetOutput(io.Discard)
				tm := tx.NewTransactionManager(db, logger)

				err = tm.WithinTransaction(context.Background(), func(txCtx context.Context) error {
					return uc.DeleteRole(txCtx, roleName)
				})

				assert.NoError(t, err)
				deps.Enforcer.AssertNotCalled(t, "LoadPolicy")
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
// BATCH PERMISSION CHECK TESTS
// ============================================================================

func TestPermissionUseCase_BatchCheckPermission(t *testing.T) {
	largeItems := make([]model.PermissionCheckItem, 100)
	for i := 0; i < 100; i++ {
		largeItems[i] = model.PermissionCheckItem{Resource: "/api/resource", Action: "GET"}
	}

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_SuccessAllAllowed",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				userID := "user-123"
				items := []model.PermissionCheckItem{{Resource: "/api/users", Action: "GET"}, {Resource: "/api/users", Action: "POST"}, {Resource: "/api/roles", Action: "GET"}}

				deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil)

				results, err := uc.BatchCheckPermission(ctx, userID, items)
				assert.NoError(t, err)
				assert.Len(t, results, 3)
				assert.True(t, results["/api/users:GET"])
				assert.True(t, results["/api/users:POST"])
				assert.True(t, results["/api/roles:GET"])
			},
		},
		{
			name:     "Positive_SuccessMixed",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				userID := "user-456"
				items := []model.PermissionCheckItem{{Resource: "/api/users", Action: "GET"}, {Resource: "/api/users", Action: "DELETE"}, {Resource: "/api/roles", Action: "POST"}, {Resource: "/api/admin", Action: "GET"}}

				deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool { return p[2] == "/api/users" && p[3] == "GET" })).Return(true, nil)
				deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool { return p[2] == "/api/users" && p[3] == "DELETE" })).Return(false, nil)
				deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool { return p[2] == "/api/roles" && p[3] == "POST" })).Return(true, nil)
				deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool { return p[2] == "/api/admin" && p[3] == "GET" })).Return(false, nil)
				deps.Enforcer.On("Enforce", mock.Anything).Return(false, nil)

				results, err := uc.BatchCheckPermission(ctx, userID, items)
				assert.NoError(t, err)
				assert.Len(t, results, 4)
				assert.True(t, results["/api/users:GET"])
				assert.False(t, results["/api/users:DELETE"])
				assert.True(t, results["/api/roles:POST"])
				assert.False(t, results["/api/admin:GET"])
			},
		},
		{
			name:     "Edge_EmptyItems",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				userID := "user-789"
				items := []model.PermissionCheckItem{}

				results, err := uc.BatchCheckPermission(ctx, userID, items)
				assert.NoError(t, err)
				assert.Len(t, results, 0)
				deps.Enforcer.AssertNotCalled(t, "Enforce")
			},
		},
		{
			name:     "Edge_LargeItemList",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				userID := "user-101"

				for i := 0; i < 100; i++ {
					deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Once()
				}

				results, err := uc.BatchCheckPermission(ctx, userID, largeItems)
				assert.NoError(t, err)
				assert.Len(t, results, 1) // Only 1 distinct item
				assert.True(t, results["/api/resource:GET"])
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				ctx := context.Background()
				userID := "user-202"
				items := []model.PermissionCheckItem{{Resource: "/api/users", Action: "GET"}, {Resource: "/api/roles", Action: "POST"}, {Resource: "/api/admin", Action: "GET"}}

				deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Once()
				deps.Enforcer.On("Enforce", mock.Anything).Return(false, errors.New("casbin database error")).Once()
				deps.Enforcer.On("Enforce", mock.Anything).Return(false, nil).Once()

				results, err := uc.BatchCheckPermission(ctx, userID, items)
				assert.NoError(t, err)
				assert.Len(t, results, 3)
				assert.True(t, results["/api/users:GET"])
				assert.False(t, results["/api/roles:POST"])
				assert.False(t, results["/api/admin:GET"])
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
// GUARDIAN PERMISSION TESTS
// ============================================================================

func TestRevokeRoleFromUser_Guardian(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				userID, roleName := "user123", "editor"

				deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)

				err := uc.RevokeRoleFromUser(context.Background(), userID, roleName, "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_EmptyUser",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.RevokeRoleFromUser(context.Background(), "", "role", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required")
			},
		},
		{
			name:     "Negative_EmptyRole",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()
				err := uc.RevokeRoleFromUser(context.Background(), "user", "", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "required")
			},
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, gorm.ErrRecordNotFound)

				err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_UserRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, errors.New("db fail"))

				err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_RoleNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, gorm.ErrRecordNotFound)

				err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
				assert.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_RoleRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, errors.New("db fail"))

				err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, errors.New("casbin fail"))

				err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_RoleNotAssigned",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, nil)

				err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "role was not assigned to user")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAddParentRole_Guardian(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "viewer").Return(&roleEntity.Role{Name: "viewer"}, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

				err := uc.AddParentRole(context.Background(), "editor", "viewer", "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_ChildRoleNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(nil, errors.New("not found"))

				err := uc.AddParentRole(context.Background(), "child", "parent", "global")
				assert.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_ParentRoleNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(&roleEntity.Role{Name: "child"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "parent").Return(nil, errors.New("not found"))

				err := uc.AddParentRole(context.Background(), "child", "parent", "global")
				assert.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_SelfInheritance",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)

				err := uc.AddParentRole(context.Background(), "role", "role", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "cannot inherit from itself")
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.RoleRepo.On("FindByName", mock.Anything, "editor").Return(&roleEntity.Role{Name: "editor"}, nil)
				deps.RoleRepo.On("FindByName", mock.Anything, "viewer").Return(&roleEntity.Role{Name: "viewer"}, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin fail"))

				err := uc.AddParentRole(context.Background(), "editor", "viewer", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "casbin fail")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRemoveParentRole_Guardian(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)

				err := uc.RemoveParentRole(context.Background(), "editor", "viewer", "global")
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, errors.New("casbin fail"))

				err := uc.RemoveParentRole(context.Background(), "editor", "viewer", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "casbin fail")
			},
		},
		{
			name:     "Negative_RelationshipNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, nil)

				err := uc.RemoveParentRole(context.Background(), "editor", "viewer", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "inheritance relationship not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetParentRoles_Guardian(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				parents := []string{"viewer"}
				deps.Enforcer.On("GetRolesForUser", mock.Anything, mock.Anything).Return(parents, nil)

				res, err := uc.GetParentRoles(context.Background(), "editor", "global")
				assert.NoError(t, err)
				assert.Equal(t, parents, res)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.Enforcer.On("GetRolesForUser", mock.Anything, mock.Anything).Return(nil, errors.New("casbin fail"))

				res, err := uc.GetParentRoles(context.Background(), "editor", "global")
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "casbin fail")
				assert.Nil(t, res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestPermissionUseCase_EdgeCasesAndVulnerabilities(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Edge_MaxStringLength",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				longString := strings.Repeat("a", 1000)

				deps.RoleRepo.On("FindByName", mock.Anything, longString).Return(nil, errors.New("record not found"))

				err := uc.GrantPermissionToRole(context.Background(), longString, longString, "GET", "global")
				assert.Error(t, err)
				assert.Equal(t, exception.ErrInternalServer, err)
			},
		},
		{
			name:     "Vulnerability_SQLInjectionInRole",
			category: "vulnerability",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				sqliRole := "admin' OR '1'='1"

				deps.RoleRepo.On("FindByName", mock.Anything, sqliRole).Return(nil, errors.New("record not found"))

				err := uc.GrantPermissionToRole(context.Background(), sqliRole, "/path", "GET", "global")
				assert.Error(t, err)
				assert.True(t, errors.Is(err, exception.ErrInternalServer) || errors.Is(err, exception.ErrBadRequest))
			},
		},
		{
			name:     "Negative_AssignRoleToUser_EmptyUser",
			category: "negative",
			run: func(t *testing.T) {
				_, uc := setupPermissionTest()

				err := uc.AssignRoleToUser(context.Background(), "", "role", "global")
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "empty"))
			},
		},
		{
			name:     "Edge_GrantPermissionToRole_SpecialChars",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				role := "admin"
				path := "/api/v1/resource/!@#$%^&*()"

				deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
				deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

				err := uc.GrantPermissionToRole(context.Background(), role, path, "GET", "global")
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
