package usecase_test

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/mocking"
	permMocks "github.com/Roisfaozi/queue-base/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/role/model"
	"github.com/Roisfaozi/queue-base/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/role/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type roleTestDeps struct {
	Repo         *mocks.MockRoleRepository
	TM           *mocking.MockWithTransactionManager
	PermissionUC *permMocks.MockIPermissionUseCase
}

func setupRoleTest() (*roleTestDeps, usecase.RoleUseCase) {
	deps := &roleTestDeps{
		Repo:         new(mocks.MockRoleRepository),
		TM:           new(mocking.MockWithTransactionManager),
		PermissionUC: new(permMocks.MockIPermissionUseCase),
	}

	deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
		Return(func(ctx context.Context, fn func(context.Context) error) error {
			return fn(ctx)
		})

	log := logrus.New()
	log.SetOutput(io.Discard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, deps.PermissionUC)
	return deps, uc
}

func TestRoleUseCase_Create(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				req := &model.CreateRoleRequest{Name: "NewRole", Description: "Desc"}

				deps.Repo.On("FindByName", ctx, req.Name).Return(nil, gorm.ErrRecordNotFound)
				deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(nil)

				res, err := uc.Create(ctx, req)
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "NewRole", res.Name)
			},
		},
		{
			name:     "Negative_Conflict",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				req := &model.CreateRoleRequest{Name: "ExistingRole", Description: "Desc"}

				deps.Repo.On("FindByName", ctx, req.Name).Return(&entity.Role{}, nil)

				res, err := uc.Create(ctx, req)
				assert.ErrorIs(t, err, exception.ErrConflict)
				assert.Nil(t, res)
			},
		},
		{
			name:     "Negative_DBErrorFind",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				req := &model.CreateRoleRequest{Name: "Role", Description: "Desc"}

				deps.Repo.On("FindByName", ctx, req.Name).Return(nil, errors.New("db error"))

				res, err := uc.Create(ctx, req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
				assert.Nil(t, res)
			},
		},
		{
			name:     "Negative_DBErrorCreate",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				req := &model.CreateRoleRequest{Name: "Role", Description: "Desc"}

				deps.Repo.On("FindByName", ctx, req.Name).Return(nil, gorm.ErrRecordNotFound)
				deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(errors.New("db error"))

				res, err := uc.Create(ctx, req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
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

func TestRoleUseCase_Update(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				req := &model.UpdateRoleRequest{Description: "NewDesc"}

				deps.Repo.On("FindByID", ctx, id).Return(&entity.Role{ID: id, Name: "Role", Description: "OldDesc"}, nil)
				deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).Return(nil)

				res, err := uc.Update(ctx, id, req)
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, "NewDesc", res.Description)
			},
		},
		{
			name:     "Negative_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				req := &model.UpdateRoleRequest{Description: "NewDesc"}

				deps.Repo.On("FindByID", ctx, id).Return(nil, gorm.ErrRecordNotFound)

				res, err := uc.Update(ctx, id, req)
				assert.ErrorIs(t, err, exception.ErrNotFound)
				assert.Nil(t, res)
			},
		},
		{
			name:     "Negative_DBErrorFind",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				req := &model.UpdateRoleRequest{Description: "NewDesc"}

				deps.Repo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

				res, err := uc.Update(ctx, id, req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
				assert.Nil(t, res)
			},
		},
		{
			name:     "Negative_DBErrorUpdate",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				req := &model.UpdateRoleRequest{Description: "NewDesc"}

				deps.Repo.On("FindByID", ctx, id).Return(&entity.Role{ID: id, Name: "Role"}, nil)
				deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).Return(errors.New("db error"))

				res, err := uc.Update(ctx, id, req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
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

func TestRoleUseCase_GetAll(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()

				deps.Repo.On("FindAll", ctx).Return([]*entity.Role{{ID: "1", Name: "Role1"}, {ID: "2", Name: "Role2"}}, nil)

				res, err := uc.GetAll(ctx)
				assert.NoError(t, err)
				assert.Len(t, res, 2)
			},
		},
		{
			name:     "Negative_DBError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()

				deps.Repo.On("FindAll", ctx).Return(nil, errors.New("db error"))

				res, err := uc.GetAll(ctx)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
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

func TestRoleUseCase_Delete(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				role := &entity.Role{ID: "role-1", Name: "NormalRole"}

				deps.Repo.On("FindByID", ctx, id).Return(role, nil)
				deps.Repo.On("Delete", ctx, id).Return(nil)
				deps.PermissionUC.On("DeleteRole", ctx, role.Name).Return(nil)

				err := uc.Delete(ctx, id)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Security_ForbiddenSuperadmin",
			category: "vulnerability",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-super"
				role := &entity.Role{ID: "role-super", Name: "role:superadmin"}

				deps.Repo.On("FindByID", ctx, id).Return(role, nil)

				err := uc.Delete(ctx, id)
				assert.ErrorIs(t, err, exception.ErrForbidden)
			},
		},
		{
			name:     "Negative_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"

				deps.Repo.On("FindByID", ctx, id).Return(nil, gorm.ErrRecordNotFound)

				err := uc.Delete(ctx, id)
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_DBErrorFind",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"

				deps.Repo.On("FindByID", ctx, id).Return(nil, errors.New("db error"))

				err := uc.Delete(ctx, id)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_DBErrorDelete",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				role := &entity.Role{ID: "role-1", Name: "NormalRole"}

				deps.Repo.On("FindByID", ctx, id).Return(role, nil)
				deps.Repo.On("Delete", ctx, id).Return(errors.New("db error"))

				err := uc.Delete(ctx, id)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_DBErrorPermissionCleanup",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				id := "role-1"
				role := &entity.Role{ID: "role-1", Name: "NormalRole"}

				deps.Repo.On("FindByID", ctx, id).Return(role, nil)
				deps.Repo.On("Delete", ctx, id).Return(nil)
				deps.PermissionUC.On("DeleteRole", ctx, role.Name).Return(errors.New("perm error"))

				err := uc.Delete(ctx, id)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRoleUseCase_GetAllRolesDynamic(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				filter := &querybuilder.DynamicFilter{}

				deps.Repo.On("FindAllDynamic", ctx, filter).Return([]*entity.Role{{ID: "1", Name: "Role1"}}, nil)

				res, err := uc.GetAllRolesDynamic(ctx, filter)
				assert.NoError(t, err)
				assert.Len(t, res, 1)
			},
		},
		{
			name:     "Negative_DBError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupRoleTest()
				ctx := context.Background()
				filter := &querybuilder.DynamicFilter{}

				deps.Repo.On("FindAllDynamic", ctx, filter).Return(nil, errors.New("db error"))

				res, err := uc.GetAllRolesDynamic(ctx, filter)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
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
