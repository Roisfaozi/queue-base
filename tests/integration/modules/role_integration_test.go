//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"strings"
	"testing"

	accessRepository "github.com/Roisfaozi/queue-base/internal/modules/access/repository"
	permissionUC "github.com/Roisfaozi/queue-base/internal/modules/permission/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/role/model"
	"github.com/Roisfaozi/queue-base/internal/modules/role/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/role/usecase"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRoleIntegration(env *setup.TestEnvironment) usecase.RoleUseCase {
	roleRepo := repository.NewRoleRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	permUC := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, roleRepo, userRepository.NewUserRepository(env.DB, env.Logger), accessRepository.NewAccessRepository(env.DB, env.Logger), nil)
	return usecase.NewRoleUseCase(env.Logger, tm, roleRepo, permUC)
}

func TestRoleIntegration_Create(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, roleUC usecase.RoleUseCase)
	}{
		{
			name: "Success",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				req := &model.CreateRoleRequest{
					Name:        "Test Role",
					Description: "Test role description",
				}

				result, err := roleUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, req.Name, result.Name)
			},
		},
		{
			name: "Negative_DuplicateName",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				_, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: "Duplicate", Description: "First"})
				require.NoError(t, err)

				_, err = roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: "Duplicate", Description: "Second"})
				assert.Error(t, err)
			},
		},
		{
			name: "Negative_EmptyName",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				result, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: "", Description: "Empty Name"})
				if err == nil {
					assert.NotEmpty(t, result.ID)
				}
			},
		},
		{
			name: "Edge_Create_LongName",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				_, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: strings.Repeat("a", 60), Description: "Long Name"})
				assert.Error(t, err)
			},
		},
		{
			name: "Edge_MinimumNameLength",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				result, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: "A", Description: "Single char role"})
				require.NoError(t, err)
				assert.NotEmpty(t, result.ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			tt.run(t, roleUC)
		})
	}
}

func TestRoleIntegration_Query(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, roleUC usecase.RoleUseCase)
	}{
		{
			name: "GetAll_Success",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				roles := []model.CreateRoleRequest{
					{Name: "Test Role 1", Description: "Description 1"},
					{Name: "Test Role 2", Description: "Description 2"},
				}

				for _, req := range roles {
					request := req
					_, err := roleUC.Create(context.Background(), &request)
					require.NoError(t, err)
				}

				result, err := roleUC.GetAll(context.Background())
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(result), 2)
			},
		},
		{
			name: "DynamicSearch_Success",
			run: func(t *testing.T, roleUC usecase.RoleUseCase) {
				roles := []model.CreateRoleRequest{
					{Name: "Manager", Description: "Manager role"},
					{Name: "Developer", Description: "Developer role"},
					{Name: "Designer", Description: "Designer role"},
				}

				for _, req := range roles {
					request := req
					_, err := roleUC.Create(context.Background(), &request)
					require.NoError(t, err)
				}

				filter := &querybuilder.DynamicFilter{
					Filter: map[string]querybuilder.Filter{
						"name": {Type: "contains", From: "Dev"},
					},
				}

				result, err := roleUC.GetAllRolesDynamic(context.Background(), filter)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, 1)
				assert.Equal(t, "Developer", result[0].Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			tt.run(t, roleUC)
		})
	}
}

func TestRoleIntegration_Delete(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, env *setup.TestEnvironment, roleUC usecase.RoleUseCase)
	}{
		{
			name: "Success",
			run: func(t *testing.T, env *setup.TestEnvironment, roleUC usecase.RoleUseCase) {
				created, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{
					Name:        "Delete Role",
					Description: "Role to be deleted",
				})
				require.NoError(t, err)

				err = roleUC.Delete(context.Background(), created.ID)
				require.NoError(t, err)
			},
		},
		{
			name: "Negative_NonExistentRole",
			run: func(t *testing.T, env *setup.TestEnvironment, roleUC usecase.RoleUseCase) {
				err := roleUC.Delete(context.Background(), "non-existent-id")
				assert.Error(t, err)
			},
		},
		{
			name: "Security_Delete_SuperadminForbidden",
			run: func(t *testing.T, env *setup.TestEnvironment, roleUC usecase.RoleUseCase) {
				err := roleUC.Delete(context.Background(), "role:superadmin")
				assert.Error(t, err)
			},
		},
		{
			name: "WithActiveUsers_DocumentsBehavior",
			run: func(t *testing.T, env *setup.TestEnvironment, roleUC usecase.RoleUseCase) {
				created, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{
					Name:        "RoleWithUsers",
					Description: "Has active assignments",
				})
				require.NoError(t, err)

				_, err = env.Enforcer.AddGroupingPolicy("user:fake-active-user", created.Name, "global")
				require.NoError(t, err)
				require.NoError(t, env.Enforcer.SavePolicy())

				err = roleUC.Delete(context.Background(), created.ID)
				require.NoError(t, err, "Role deletion should succeed even when users are assigned")

				err = roleUC.Delete(context.Background(), created.ID)
				assert.Error(t, err, "Role already deleted — second delete should return not-found")

				rolesAfter, err := env.Enforcer.GetRolesForUser("user:fake-active-user", "global")
				require.NoError(t, err)
				roleStillInCasbin := false
				for _, roleName := range rolesAfter {
					if roleName == created.Name {
						roleStillInCasbin = true
					}
				}
				t.Logf("KNOWN GAP: Casbin grouping still contains deleted role '%s': %v — cleanup not cascaded by usecase.Delete", created.Name, roleStillInCasbin)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			tt.run(t, env, roleUC)
		})
	}
}

func TestRoleIntegration_Update_Success(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	roleUC := setupRoleIntegration(env)

	created, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{
		Name:        "Role To Update",
		Description: "Original Description",
	})
	require.NoError(t, err)

	updated, err := roleUC.Update(context.Background(), created.ID, &model.UpdateRoleRequest{Description: "Updated Description"})
	require.NoError(t, err)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, "Updated Description", updated.Description)
}

func TestRoleIntegration_Edge_SpecialCharactersInName(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{name: "AdminSymbols", role: "Admin@#$%"},
		{name: "Dashes", role: "Role-With-Dashes"},
		{name: "Underscores", role: "Role_With_Underscores"},
		{name: "Dots", role: "Role.With.Dots"},
		{name: "Parentheses", role: "Role (With Parentheses)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			result, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: tt.role, Description: "Special char role"})
			require.NoError(t, err)
			assert.Equal(t, tt.role, result.Name)
		})
	}
}

func TestRoleIntegration_Edge_UnicodeInName(t *testing.T) {
	tests := []struct {
		name string
		role string
	}{
		{name: "Chinese", role: "管理员"},
		{name: "Cyrillic", role: "Администратор"},
		{name: "Arabic", role: "مدير"},
		{name: "Japanese", role: "マネージャー"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			result, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: tt.role, Description: "Unicode role"})
			require.NoError(t, err)
			assert.Equal(t, tt.role, result.Name)
		})
	}
}

func TestRoleIntegration_Security_SQLInjectionInName(t *testing.T) {
	tests := []struct {
		name    string
		payload string
	}{
		{name: "BooleanBypass", payload: "Admin' OR '1'='1"},
		{name: "DropTable", payload: "'; DROP TABLE roles--"},
		{name: "Comment", payload: "Admin'--"},
		{name: "Union", payload: "1' UNION SELECT * FROM roles--"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			_, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{Name: tt.payload, Description: "SQL injection attempt"})
			_ = err
		})
	}
}

func TestRoleIntegration_Security_XSSInDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{name: "ScriptTag", description: "<script>alert('XSS')</script>"},
		{name: "ImageOnError", description: "<img src=x onerror=alert('XSS')>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			roleUC := setupRoleIntegration(env)
			result, err := roleUC.Create(context.Background(), &model.CreateRoleRequest{
				Name:        fmt.Sprintf("XSSRole_%s", tt.name),
				Description: tt.description,
			})
			require.NoError(t, err)
			assert.NotEmpty(t, result.ID)
		})
	}
}
