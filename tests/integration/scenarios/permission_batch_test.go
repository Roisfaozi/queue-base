//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/queue-base/internal/modules/access/repository"
	permissionModel "github.com/Roisfaozi/queue-base/internal/modules/permission/model"
	permissionUC "github.com/Roisfaozi/queue-base/internal/modules/permission/usecase"
	roleModel "github.com/Roisfaozi/queue-base/internal/modules/role/model"
	roleRepo "github.com/Roisfaozi/queue-base/internal/modules/role/repository"
	roleUC "github.com/Roisfaozi/queue-base/internal/modules/role/usecase"
	userRepo "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_PermissionBatchCheck(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()
	setup.CleanupDatabase(t, env.DB)

	ctx := context.Background()
	tm := tx.NewTransactionManager(env.DB, env.Logger)

	rRepo := roleRepo.NewRoleRepository(env.DB, env.Logger)
	uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
	aRepo := accessRepo.NewAccessRepository(env.DB, env.Logger)
	permService := permissionUC.NewPermissionUseCase(env.Enforcer, env.Logger, rRepo, uRepo, aRepo, nil)
	roleService := roleUC.NewRoleUseCase(env.Logger, tm, rRepo, permService)

	roleName := "Editor"
	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: roleName})
	require.NoError(t, err)

	user := setup.CreateTestUser(t, env.DB, "editor_user", "editor@batch.com", "pass")
	err = permService.AssignRoleToUser(ctx, user.ID, roleName, "global")
	require.NoError(t, err)

	err = permService.GrantPermissionToRole(ctx, roleName, "/articles", "READ", "global")
	require.NoError(t, err)
	err = permService.GrantPermissionToRole(ctx, roleName, "/articles", "WRITE", "global")
	require.NoError(t, err)

	tests := []struct {
		name     string
		item     permissionModel.PermissionCheckItem
		expected bool
	}{
		{name: "ArticlesRead", item: permissionModel.PermissionCheckItem{Resource: "/articles", Action: "READ"}, expected: true},
		{name: "ArticlesWrite", item: permissionModel.PermissionCheckItem{Resource: "/articles", Action: "WRITE"}, expected: true},
		{name: "ArticlesDelete", item: permissionModel.PermissionCheckItem{Resource: "/articles", Action: "DELETE"}, expected: false},
		{name: "UsersRead", item: permissionModel.PermissionCheckItem{Resource: "/users", Action: "READ"}, expected: false},
	}

	items := make([]permissionModel.PermissionCheckItem, 0, len(tests))
	for _, tt := range tests {
		items = append(items, tt.item)
	}

	results, err := permService.BatchCheckPermission(ctx, user.ID, items)
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, results[tt.item.Resource+":"+tt.item.Action])
		})
	}
}
