//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"

	accessRepo "github.com/Roisfaozi/queue-base/internal/modules/access/repository"
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

func TestScenario_RoleHierarchy(t *testing.T) {
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

	parentRole := "Manager"
	childRole := "Staff"

	_, err := roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: parentRole})
	require.NoError(t, err)
	_, err = roleService.Create(ctx, &roleModel.CreateRoleRequest{Name: childRole})
	require.NoError(t, err)

	path := "/api/v1/work"
	method := "GET"
	err = permService.GrantPermissionToRole(ctx, childRole, path, method, "global")
	require.NoError(t, err)

	checks := []struct {
		name     string
		expected bool
	}{
		{name: "BeforeInheritance", expected: false},
		{name: "AfterInheritance", expected: true},
	}

	ok, err := env.Enforcer.Enforce(parentRole, "global", path, method)
	require.NoError(t, err)
	assert.Equal(t, checks[0].expected, ok, "Parent role should not have access yet")

	err = permService.AddParentRole(ctx, parentRole, childRole, "global")
	require.NoError(t, err)

	ok, err = env.Enforcer.Enforce(parentRole, "global", path, method)
	require.NoError(t, err)
	assert.Equal(t, checks[1].expected, ok, "Parent role (Manager) should inherit permissions from Child role (Staff)")
}
