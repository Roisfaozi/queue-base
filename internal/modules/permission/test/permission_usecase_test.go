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
		name      string
		category  string
		userID    string
		roleName  string
		userErr   error
		user      *userEntity.User
		roleErr   error
		role      *roleEntity.Role
		removeErr error
		addErr    error
		wantErr   error
	}{
		{name: "Success", category: "positive", userID: "user123", roleName: "editor", user: &userEntity.User{ID: "user123"}, role: &roleEntity.Role{Name: "editor"}},
		{name: "UserNotFound", category: "negative", userID: "user123", roleName: "editor", userErr: gorm.ErrRecordNotFound, wantErr: exception.ErrNotFound},
		{name: "UserRepoError", category: "negative", userID: "user123", roleName: "editor", userErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "RoleNotFound", category: "negative", userID: "user123", roleName: "non_existent_role", user: &userEntity.User{ID: "user123"}, roleErr: gorm.ErrRecordNotFound, wantErr: exception.ErrBadRequest},
		{name: "RoleRepoError", category: "negative", userID: "user123", roleName: "editor", user: &userEntity.User{ID: "user123"}, roleErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "EnforcerRemoveError", category: "negative", userID: "user123", roleName: "editor", user: &userEntity.User{ID: "user123"}, role: &roleEntity.Role{Name: "editor"}, removeErr: errors.New("casbin error"), wantErr: exception.ErrInternalServer},
		{name: "EnforcerAddError", category: "negative", userID: "user123", roleName: "editor", user: &userEntity.User{ID: "user123"}, role: &roleEntity.Role{Name: "editor"}, addErr: errors.New("casbin error"), wantErr: errors.New("casbin error")},
		{name: "EmptyUserID", category: "negative", userID: "", roleName: "role", wantErr: errors.New("userID and role are required")},
		{name: "EmptyRole", category: "negative", userID: "user", roleName: "", wantErr: errors.New("userID and role are required")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			ctx := context.Background()

			if tt.userID != "" && tt.roleName != "" {
				deps.UserRepo.On("FindByID", mock.Anything, tt.userID).Return(tt.user, tt.userErr).Maybe()
				if tt.userErr == nil && tt.user != nil {
					deps.RoleRepo.On("FindByName", mock.Anything, tt.roleName).Return(tt.role, tt.roleErr).Maybe()
					if tt.roleErr == nil && tt.role != nil {
						if tt.removeErr != nil {
							deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything, mock.Anything).Return(false, tt.removeErr).Maybe()
						} else {
							deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil).Maybe()
							if tt.addErr != nil {
								deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, tt.addErr).Maybe()
							} else {
								deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil).Maybe()
							}
						}
					}
				}
			}

			err := uc.AssignRoleToUser(ctx, tt.userID, tt.roleName, "global")
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGrantPermissionToRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		role     string
		path     string
		method   string
		roleErr  error
		roleEnt  *roleEntity.Role
		addErr   error
		wantErr  error
	}{
		{name: "Success", category: "positive", role: "editor", path: "/api/v1/articles", method: "POST", roleEnt: &roleEntity.Role{Name: "editor"}},
		{name: "RoleNotFound", category: "negative", role: "non_existent_role", path: "/api/v1/articles", method: "POST", roleErr: gorm.ErrRecordNotFound, wantErr: exception.ErrBadRequest},
		{name: "RoleRepoError", category: "negative", role: "role", path: "/path", method: "POST", roleErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "EnforcerError", category: "negative", role: "role", path: "/path", method: "POST", roleEnt: &roleEntity.Role{Name: "role"}, addErr: errors.New("casbin error"), wantErr: errors.New("casbin error")},
		{name: "EmptyRole", category: "negative", role: "", path: "path", method: "GET", wantErr: errors.New("role, path, and method are required")},
		{name: "EmptyPath", category: "negative", role: "role", path: "", method: "GET", wantErr: errors.New("role, path, and method are required")},
		{name: "EmptyMethod", category: "negative", role: "role", path: "path", method: "", wantErr: errors.New("role, path, and method are required")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			ctx := context.Background()
			if tt.role != "" && tt.path != "" && tt.method != "" {
				deps.RoleRepo.On("FindByName", mock.Anything, tt.role).Return(tt.roleEnt, tt.roleErr).Maybe()
				if tt.roleErr == nil && tt.roleEnt != nil {
					if tt.addErr != nil {
						deps.Enforcer.On("AddPolicy", mock.Anything).Return(false, tt.addErr).Maybe()
					} else {
						deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil).Maybe()
					}
				}
			}
			err := uc.GrantPermissionToRole(ctx, tt.role, tt.path, tt.method, "global")
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRevokePermissionFromRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		role     string
		path     string
		method   string
		roleErr  error
		roleEnt  *roleEntity.Role
		removeOK bool
		remErr   error
		wantErr  error
	}{
		{name: "Success", category: "positive", role: "editor", path: "/api/v1/articles", method: "DELETE", roleEnt: &roleEntity.Role{Name: "editor"}, removeOK: true},
		{name: "RoleNotFound", category: "negative", role: "non_existent_role", path: "/api/v1/articles", method: "DELETE", roleErr: gorm.ErrRecordNotFound, wantErr: exception.ErrBadRequest},
		{name: "RoleRepoError", category: "negative", role: "role", path: "/path", method: "DELETE", roleErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "EnforcerError", category: "negative", role: "role", path: "/path", method: "DELETE", roleEnt: &roleEntity.Role{Name: "role"}, remErr: errors.New("casbin error"), wantErr: errors.New("casbin error")},
		{name: "PolicyNotFound", category: "negative", role: "role", path: "/path", method: "DELETE", roleEnt: &roleEntity.Role{Name: "role"}, removeOK: false, wantErr: errors.New("policy to revoke not found in specified domain")},
		{name: "EmptyRole", category: "negative", role: "", path: "path", method: "GET", wantErr: errors.New("role, path, and method are required")},
		{name: "EmptyPath", category: "negative", role: "role", path: "", method: "GET", wantErr: errors.New("role, path, and method are required")},
		{name: "EmptyMethod", category: "negative", role: "role", path: "path", method: "", wantErr: errors.New("role, path, and method are required")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			ctx := context.Background()
			if tt.role != "" && tt.path != "" && tt.method != "" {
				deps.RoleRepo.On("FindByName", mock.Anything, tt.role).Return(tt.roleEnt, tt.roleErr).Maybe()
				if tt.roleErr == nil && tt.roleEnt != nil {
					deps.Enforcer.On("RemovePolicy", mock.Anything).Return(tt.removeOK, tt.remErr).Maybe()
				}
			}
			err := uc.RevokePermissionFromRole(ctx, tt.role, tt.path, tt.method, "global")
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetAllPermissions_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	expectedPolicies := [][]string{{"role", "path", "GET"}}

	deps.Enforcer.On("GetPolicy").Return(expectedPolicies, nil)

	policies, err := uc.GetAllPermissions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicies, policies)
}

func TestGetAllPermissions_Error(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.Enforcer.On("GetPolicy").Return(nil, errors.New("casbin error"))

	_, err := uc.GetAllPermissions(context.Background())
	assert.Error(t, err)
}

func TestGetPermissionsForRole_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	expectedPolicies := [][]string{{"admin", "path", "GET"}}

	deps.Enforcer.On("GetFilteredPolicy", 0, mock.Anything).Return(expectedPolicies, nil)

	policies, err := uc.GetPermissionsForRole(context.Background(), role)
	assert.NoError(t, err)
	assert.Equal(t, expectedPolicies, policies)
}

func TestGetPermissionsForRole_Error(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	deps.Enforcer.On("GetFilteredPolicy", 0, mock.Anything).Return(nil, errors.New("casbin error"))

	_, err := uc.GetPermissionsForRole(context.Background(), role)
	assert.Error(t, err)
}

func TestUpdatePermission_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"role", "old", "GET"}
	newP := []string{"role", "new", "GET"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(true, nil)

	updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
	assert.NoError(t, err)
	assert.True(t, updated)
}

func TestUpdatePermission_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	updated, err := uc.UpdatePermission(context.Background(), []string{}, []string{"a"})
	assert.Error(t, err)
	assert.False(t, updated)

	updated, err = uc.UpdatePermission(context.Background(), []string{"a"}, []string{})
	assert.Error(t, err)
	assert.False(t, updated)
}

func TestUpdatePermission_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"role", "old", "GET"}
	newP := []string{"role", "new", "GET"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(false, errors.New("casbin error"))

	updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
	assert.Error(t, err)
	assert.False(t, updated)
}

func TestUpdatePermission_PolicyNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	oldP := []string{"role", "old", "GET"}
	newP := []string{"role", "new", "GET"}

	deps.Enforcer.On("UpdatePolicy", oldP, newP).Return(false, nil)

	updated, err := uc.UpdatePermission(context.Background(), oldP, newP)
	assert.Error(t, err)
	assert.False(t, updated)
}

func TestDeleteRole_ReloadsPolicyOutsideTransaction(t *testing.T) {
	deps, uc := setupPermissionTest()
	roleName := "role:editor"

	deps.Enforcer.On("DeleteRole", roleName).Return(true, nil)
	deps.Enforcer.On("LoadPolicy").Return(nil)

	err := uc.DeleteRole(context.Background(), roleName)

	assert.NoError(t, err)
	deps.Enforcer.AssertCalled(t, "LoadPolicy")
}

func TestDeleteRole_DefersPolicyReloadInsideTransaction(t *testing.T) {
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
}

// ============================================================================
// BATCH PERMISSION CHECK TESTS
// ============================================================================

func TestPermissionUseCase_BatchCheckPermission_Success_AllAllowed(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-123"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/users", Action: "POST"},
		{Resource: "/api/roles", Action: "GET"},
	}

	// Mock Enforce - All allowed
	deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.True(t, results["/api/users:POST"])
	assert.True(t, results["/api/roles:GET"])
}

func TestPermissionUseCase_BatchCheckPermission_Success_Mixed(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-456"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},    // Allowed
		{Resource: "/api/users", Action: "DELETE"}, // Denied
		{Resource: "/api/roles", Action: "POST"},   // Allowed
		{Resource: "/api/admin", Action: "GET"},    // Denied
	}

	// Mock Enforce - Mixed results
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/users" && p[3] == "GET"
	})).Return(true, nil)
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/users" && p[3] == "DELETE"
	})).Return(false, nil)
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/roles" && p[3] == "POST"
	})).Return(true, nil)
	deps.Enforcer.On("Enforce", mock.MatchedBy(func(p []interface{}) bool {
		return p[2] == "/api/admin" && p[3] == "GET"
	})).Return(false, nil)
	deps.Enforcer.On("Enforce", mock.Anything).Return(false, nil)

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 4, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.False(t, results["/api/users:DELETE"])
	assert.True(t, results["/api/roles:POST"])
	assert.False(t, results["/api/admin:GET"])
}

func TestPermissionUseCase_BatchCheckPermission_EmptyItems(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-789"
	items := []model.PermissionCheckItem{} // Empty list

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Should return empty map
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 0, len(results))
	deps.Enforcer.AssertNotCalled(t, "Enforce")
}

func TestPermissionUseCase_BatchCheckPermission_LargeItemList(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-101"

	// Create 100 items
	items := make([]model.PermissionCheckItem, 100)
	for i := 0; i < 100; i++ {
		items[i] = model.PermissionCheckItem{
			Resource: "/api/resource",
			Action:   "GET",
		}
		// Mock each enforce call
		deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Once()
	}

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, results)
	// Note: All items have same resource:action, so only 1 key in map
	assert.Equal(t, 1, len(results))
	assert.True(t, results["/api/resource:GET"])
}

func TestPermissionUseCase_BatchCheckPermission_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	ctx := context.Background()

	userID := "user-202"
	items := []model.PermissionCheckItem{
		{Resource: "/api/users", Action: "GET"},
		{Resource: "/api/roles", Action: "POST"}, // This will error
		{Resource: "/api/admin", Action: "GET"},
	}

	// Mock Enforce - One with error
	deps.Enforcer.On("Enforce", mock.Anything).Return(true, nil).Once()
	deps.Enforcer.On("Enforce", mock.Anything).Return(false, errors.New("casbin database error")).Once()
	deps.Enforcer.On("Enforce", mock.Anything).Return(false, nil).Once()

	// Execute
	results, err := uc.BatchCheckPermission(ctx, userID, items)

	// Assert - Should not fail, but log error and continue
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, 3, len(results))
	assert.True(t, results["/api/users:GET"])
	assert.False(t, results["/api/roles:POST"]) // Error treated as false
	assert.False(t, results["/api/admin:GET"])
}

// ============================================================================
// GUARDIAN PERMISSION TESTS
// ============================================================================

func TestRevokeRoleFromUser_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	userID, roleName := "user123", "editor"

	deps.UserRepo.On("FindByID", mock.Anything, userID).Return(&userEntity.User{ID: userID}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, roleName).Return(&roleEntity.Role{Name: roleName}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)

	err := uc.RevokeRoleFromUser(context.Background(), userID, roleName, "global")
	assert.NoError(t, err)
}

func TestRevokeRoleFromUser_Guardian_EmptyInput(t *testing.T) {
	_, uc := setupPermissionTest()
	assert.Error(t, uc.RevokeRoleFromUser(context.Background(), "", "role", "global"))
	assert.Error(t, uc.RevokeRoleFromUser(context.Background(), "user", "", "global"))
}

func TestRevokeRoleFromUser_Guardian_UserNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestRevokeRoleFromUser_Guardian_UserRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(nil, errors.New("db fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_RoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, gorm.ErrRecordNotFound)

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRevokeRoleFromUser_Guardian_RoleRepoError(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(nil, errors.New("db fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, errors.New("casbin fail"))

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.ErrorIs(t, err, exception.ErrInternalServer)
}

func TestRevokeRoleFromUser_Guardian_RoleNotAssigned(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.UserRepo.On("FindByID", mock.Anything, "u").Return(&userEntity.User{ID: "u"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "r").Return(&roleEntity.Role{Name: "r"}, nil)
	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, nil) // Removed = false

	err := uc.RevokeRoleFromUser(context.Background(), "u", "r", "global")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "role was not assigned to user")
}

func TestAddParentRole_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.RoleRepo.On("FindByName", mock.Anything, child).Return(&roleEntity.Role{Name: child}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, parent).Return(&roleEntity.Role{Name: parent}, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

	err := uc.AddParentRole(context.Background(), child, parent, "global")
	assert.NoError(t, err)
}

func TestAddParentRole_Guardian_ChildRoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(nil, errors.New("not found"))

	err := uc.AddParentRole(context.Background(), "child", "parent", "global")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestAddParentRole_Guardian_ParentRoleNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "child").Return(&roleEntity.Role{Name: "child"}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, "parent").Return(nil, errors.New("not found"))

	err := uc.AddParentRole(context.Background(), "child", "parent", "global")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestAddParentRole_Guardian_SelfInheritance(t *testing.T) {
	deps, uc := setupPermissionTest()
	deps.RoleRepo.On("FindByName", mock.Anything, "role").Return(&roleEntity.Role{Name: "role"}, nil)

	err := uc.AddParentRole(context.Background(), "role", "role", "global")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot inherit from itself")
}

func TestAddParentRole_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.RoleRepo.On("FindByName", mock.Anything, child).Return(&roleEntity.Role{Name: child}, nil)
	deps.RoleRepo.On("FindByName", mock.Anything, parent).Return(&roleEntity.Role{Name: parent}, nil)
	deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin fail"))

	err := uc.AddParentRole(context.Background(), child, parent, "global")
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestRemoveParentRole_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)

	err := uc.RemoveParentRole(context.Background(), child, parent, "global")
	assert.NoError(t, err)
}

func TestRemoveParentRole_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, errors.New("casbin fail"))

	err := uc.RemoveParentRole(context.Background(), child, parent, "global")
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestRemoveParentRole_Guardian_RelationshipNotFound(t *testing.T) {
	deps, uc := setupPermissionTest()
	child, parent := "editor", "viewer"

	deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(false, nil)

	err := uc.RemoveParentRole(context.Background(), child, parent, "global")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "inheritance relationship not found")
}

func TestGetParentRoles_Guardian_Success(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "editor"
	parents := []string{"viewer"}

	deps.Enforcer.On("GetRolesForUser", mock.Anything, mock.Anything).Return(parents, nil)

	res, err := uc.GetParentRoles(context.Background(), role, "global")
	assert.NoError(t, err)
	assert.Equal(t, parents, res)
}

func TestGetParentRoles_Guardian_EnforcerError(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "editor"

	deps.Enforcer.On("GetRolesForUser", mock.Anything, mock.Anything).Return(nil, errors.New("casbin fail"))

	_, err := uc.GetParentRoles(context.Background(), role, "global")
	assert.Error(t, err)
	assert.Equal(t, "casbin fail", err.Error())
}

func TestPermissionUseCase_Edge_MaxStringLength(t *testing.T) {
	deps, uc := setupPermissionTest()
	longString := strings.Repeat("a", 1000)

	deps.RoleRepo.On("FindByName", mock.Anything, longString).Return(nil, errors.New("record not found"))

	err := uc.GrantPermissionToRole(context.Background(), longString, longString, "GET", "global")
	assert.Error(t, err)
	assert.Equal(t, exception.ErrInternalServer, err)
}

func TestPermissionUseCase_Vulnerability_SQLInjectionInRole(t *testing.T) {
	deps, uc := setupPermissionTest()
	sqliRole := "admin' OR '1'='1"

	deps.RoleRepo.On("FindByName", mock.Anything, sqliRole).Return(nil, errors.New("record not found"))

	err := uc.GrantPermissionToRole(context.Background(), sqliRole, "/path", "GET", "global")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, exception.ErrInternalServer) || errors.Is(err, exception.ErrBadRequest))
}

func TestPermissionUseCase_Negative_AssignRoleToUser_EmptyUser(t *testing.T) {
	_, uc := setupPermissionTest()

	err := uc.AssignRoleToUser(context.Background(), "", "role", "global")
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "empty"))
}

func TestPermissionUseCase_Negative_GrantPermissionToRole_SpecialChars(t *testing.T) {
	deps, uc := setupPermissionTest()
	role := "admin"
	path := "/api/v1/resource/!@#$%^&*()"

	deps.RoleRepo.On("FindByName", mock.Anything, role).Return(&roleEntity.Role{Name: role}, nil)
	deps.Enforcer.On("AddPolicy", mock.Anything).Return(true, nil)

	err := uc.GrantPermissionToRole(context.Background(), role, path, "GET", "global")
	assert.NoError(t, err)

	deps.RoleRepo.AssertExpectations(t)
	deps.Enforcer.AssertExpectations(t)
}
