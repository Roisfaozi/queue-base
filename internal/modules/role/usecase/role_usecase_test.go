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
		name        string
		category    string
		req         *model.CreateRoleRequest
		findResult  *entity.Role
		findErr     error
		createErr   error
		wantErr     error
		wantResName string
	}{
		{name: "Success", category: "positive", req: &model.CreateRoleRequest{Name: "NewRole", Description: "Desc"}, findErr: gorm.ErrRecordNotFound, wantResName: "NewRole"},
		{name: "Conflict", category: "negative", req: &model.CreateRoleRequest{Name: "ExistingRole", Description: "Desc"}, findResult: &entity.Role{}, wantErr: exception.ErrConflict},
		{name: "DBErrorFind", category: "negative", req: &model.CreateRoleRequest{Name: "Role", Description: "Desc"}, findErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "DBErrorCreate", category: "negative", req: &model.CreateRoleRequest{Name: "Role", Description: "Desc"}, findErr: gorm.ErrRecordNotFound, createErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupRoleTest()
			ctx := context.Background()

			deps.Repo.On("FindByName", ctx, tt.req.Name).Return(tt.findResult, tt.findErr)
			if tt.findErr == gorm.ErrRecordNotFound {
				deps.Repo.On("Create", ctx, mock.AnythingOfType("*entity.Role")).Return(tt.createErr).Maybe()
			}

			res, err := uc.Create(ctx, tt.req)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantResName, res.Name)
		})
	}
}

func TestRoleUseCase_Update(t *testing.T) {
	tests := []struct {
		name        string
		category    string
		id          string
		req         *model.UpdateRoleRequest
		findResult  *entity.Role
		findErr     error
		updateErr   error
		wantErr     error
		wantResDesc string
	}{
		{name: "Success", category: "positive", id: "role-1", req: &model.UpdateRoleRequest{Description: "NewDesc"}, findResult: &entity.Role{ID: "role-1", Name: "Role", Description: "OldDesc"}, wantResDesc: "NewDesc"},
		{name: "NotFound", category: "negative", id: "role-1", req: &model.UpdateRoleRequest{Description: "NewDesc"}, findErr: gorm.ErrRecordNotFound, wantErr: exception.ErrNotFound},
		{name: "DBErrorFind", category: "negative", id: "role-1", req: &model.UpdateRoleRequest{Description: "NewDesc"}, findErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "DBErrorUpdate", category: "negative", id: "role-1", req: &model.UpdateRoleRequest{Description: "NewDesc"}, findResult: &entity.Role{ID: "role-1", Name: "Role"}, updateErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupRoleTest()
			ctx := context.Background()

			deps.Repo.On("FindByID", ctx, tt.id).Return(tt.findResult, tt.findErr)
			if tt.findErr == nil && tt.findResult != nil {
				deps.Repo.On("Update", ctx, mock.AnythingOfType("*entity.Role")).Return(tt.updateErr)
			}

			res, err := uc.Update(ctx, tt.id, tt.req)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, res)
			assert.Equal(t, tt.wantResDesc, res.Description)
		})
	}
}

func TestRoleUseCase_GetAll(t *testing.T) {
	tests := []struct {
		name     string
		category string
		roles    []*entity.Role
		repoErr  error
		wantLen  int
		wantErr  error
	}{
		{name: "Success", category: "positive", roles: []*entity.Role{{ID: "1", Name: "Role1"}, {ID: "2", Name: "Role2"}}, wantLen: 2},
		{name: "DBError", category: "negative", repoErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupRoleTest()
			ctx := context.Background()

			deps.Repo.On("FindAll", ctx).Return(tt.roles, tt.repoErr)

			res, err := uc.GetAll(ctx)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, res, tt.wantLen)
		})
	}
}

func TestRoleUseCase_Delete(t *testing.T) {
	tests := []struct {
		name       string
		category   string
		id         string
		role       *entity.Role
		findErr    error
		deleteErr  error
		cleanupErr error
		wantErr    error
	}{
		{name: "Success", category: "positive", id: "role-1", role: &entity.Role{ID: "role-1", Name: "NormalRole"}},
		{name: "ForbiddenSuperadmin", category: "vulnerability", id: "role-super", role: &entity.Role{ID: "role-super", Name: "role:superadmin"}, wantErr: exception.ErrForbidden},
		{name: "NotFound", category: "negative", id: "role-1", findErr: gorm.ErrRecordNotFound, wantErr: exception.ErrNotFound},
		{name: "DBErrorFind", category: "negative", id: "role-1", findErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "DBErrorDelete", category: "negative", id: "role-1", role: &entity.Role{ID: "role-1", Name: "NormalRole"}, deleteErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
		{name: "DBErrorPermissionCleanup", category: "negative", id: "role-1", role: &entity.Role{ID: "role-1", Name: "NormalRole"}, cleanupErr: errors.New("perm error"), wantErr: exception.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupRoleTest()
			ctx := context.Background()

			deps.Repo.On("FindByID", ctx, tt.id).Return(tt.role, tt.findErr)
			if tt.findErr == nil && tt.role != nil && tt.role.Name != "role:superadmin" {
				deps.Repo.On("Delete", ctx, tt.id).Return(tt.deleteErr).Maybe()
				if tt.deleteErr == nil {
					deps.PermissionUC.On("DeleteRole", ctx, tt.role.Name).Return(tt.cleanupErr).Maybe()
				}
			}

			err := uc.Delete(ctx, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestRoleUseCase_GetAllRolesDynamic(t *testing.T) {
	tests := []struct {
		name     string
		category string
		roles    []*entity.Role
		repoErr  error
		wantLen  int
		wantErr  error
	}{
		{name: "Success", category: "positive", roles: []*entity.Role{{ID: "1", Name: "Role1"}}, wantLen: 1},
		{name: "DBError", category: "negative", repoErr: errors.New("db error"), wantErr: exception.ErrInternalServer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupRoleTest()
			ctx := context.Background()
			filter := &querybuilder.DynamicFilter{}

			deps.Repo.On("FindAllDynamic", ctx, filter).Return(tt.roles, tt.repoErr)

			res, err := uc.GetAllRolesDynamic(ctx, filter)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, res)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, res, tt.wantLen)
		})
	}
}
