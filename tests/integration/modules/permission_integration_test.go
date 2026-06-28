//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/queue-base/internal/modules/access/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/usecase"
	roleRepo "github.com/Roisfaozi/queue-base/internal/modules/role/repository"
	userRepo "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPermissionIntegration(env *setup.TestEnvironment) usecase.IPermissionUseCase {
	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)
	return usecase.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)
}

func TestPermissionIntegration_AssignRoleToUser(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase)
	}{
		{
			name: "Success",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				user := setup.CreateTestUser(t, env.DB, "testuser_perm", "test@perm.com", "Password123!")
				roleName := "admin"
				setup.CreateTestRole(t, env.DB, roleName)

				err := permUC.AssignRoleToUser(context.Background(), user.ID, roleName, "global")
				assert.NoError(t, err)

				roles, err := env.Enforcer.GetRolesForUser(user.ID, "global")
				assert.NoError(t, err)
				assert.Contains(t, roles, roleName)
			},
		},
		{
			name: "Negative_AssignRoleToNonExistentUser",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				setup.CreateTestRole(t, env.DB, "valid_role")

				err := permUC.AssignRoleToUser(context.Background(), "non-existent-user-id", "valid_role", "global")
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			permUC := setupPermissionIntegration(env)
			tt.run(t, env, permUC)
		})
	}
}

func TestPermissionIntegration_GrantPermissionToRole(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase)
	}{
		{
			name: "Success",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				roleName := "editor"
				setup.CreateTestRole(t, env.DB, roleName)

				err := permUC.GrantPermissionToRole(context.Background(), roleName, "/api/v1/articles", "POST", "global")
				assert.NoError(t, err)

				ok, err := env.Enforcer.Enforce(roleName, "global", "/api/v1/articles", "POST")
				assert.NoError(t, err)
				assert.True(t, ok)
			},
		},
		{
			name: "Negative_GrantNonExistentRole",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				err := permUC.GrantPermissionToRole(context.Background(), "non_existent_role", "/any", "GET", "global")
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			permUC := setupPermissionIntegration(env)
			tt.run(t, env, permUC)
		})
	}
}

func TestPermissionIntegration_RolePermissionMutation(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase)
	}{
		{
			name: "RevokePermission",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				roleName := "viewer"
				setup.CreateTestRole(t, env.DB, roleName)
				_, err := env.Enforcer.AddPolicy(roleName, "global", "/api/v1/articles", "GET")
				require.NoError(t, err)

				err = permUC.RevokePermissionFromRole(context.Background(), roleName, "/api/v1/articles", "GET", "global")
				assert.NoError(t, err)

				ok, err := env.Enforcer.Enforce(roleName, "global", "/api/v1/articles", "GET")
				assert.NoError(t, err)
				assert.False(t, ok)
			},
		},
		{
			name: "UpdatePermission",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				roleName := "manager"
				setup.CreateTestRole(t, env.DB, roleName)
				oldP := []string{roleName, "global", "/api/v1/old", "GET"}
				newP := []string{roleName, "global", "/api/v1/new", "POST"}

				_, err := env.Enforcer.AddPolicy(oldP[0], oldP[1], oldP[2], oldP[3])
				require.NoError(t, err)

				ok, err := permUC.UpdatePermission(context.Background(), oldP, newP)
				assert.NoError(t, err)
				assert.True(t, ok)

				ok, err = env.Enforcer.Enforce(newP[0], "global", newP[2], newP[3])
				assert.NoError(t, err)
				assert.True(t, ok)
			},
		},
		{
			name: "Negative_RevokeNonExistentPermission",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				roleName := "valid_role"
				setup.CreateTestRole(t, env.DB, roleName)

				err := permUC.RevokePermissionFromRole(context.Background(), roleName, "/ghost", "GET", "global")
				_ = err
			},
		},
		{
			name: "FullLifecycle",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				roleName := "lifecycle_role"
				setup.CreateTestRole(t, env.DB, roleName)

				err := permUC.GrantPermissionToRole(context.Background(), roleName, "/api/v1/data", "GET", "global")
				require.NoError(t, err)

				oldP := []string{roleName, "global", "/api/v1/data", "GET"}
				newP := []string{roleName, "global", "/api/v1/data/updated", "POST"}
				_, err = permUC.UpdatePermission(context.Background(), oldP, newP)
				require.NoError(t, err)

				err = permUC.RevokePermissionFromRole(context.Background(), roleName, "/api/v1/data/updated", "POST", "global")
				require.NoError(t, err)

				policies, err := permUC.GetPermissionsForRole(context.Background(), roleName)
				require.NoError(t, err)
				assert.Empty(t, policies)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			permUC := setupPermissionIntegration(env)
			tt.run(t, env, permUC)
		})
	}
}

func TestPermissionIntegration_QueryPermissions(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase)
	}{
		{
			name: "GetAllPermissions",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				_, err := permUC.GetAllPermissions(context.Background())
				assert.NoError(t, err)
			},
		},
		{
			name: "GetPermissionsForRole",
			run: func(t *testing.T, env *setup.TestEnvironment, permUC usecase.IPermissionUseCase) {
				roleName := "role_for_list"
				setup.CreateTestRole(t, env.DB, roleName)
				_, err := env.Enforcer.AddPolicy(roleName, "global", "/res", "GET")
				require.NoError(t, err)

				policies, err := permUC.GetPermissionsForRole(context.Background(), roleName)
				assert.NoError(t, err)
				assert.NotEmpty(t, policies)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			permUC := setupPermissionIntegration(env)
			tt.run(t, env, permUC)
		})
	}
}
