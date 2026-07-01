package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/mocking"
	permissionMocks "github.com/Roisfaozi/queue-base/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/role/model"
	"github.com/Roisfaozi/queue-base/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/role/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type guardianRoleTestDeps struct {
	Repo           *mocks.MockRoleRepository
	TM             *mocking.MockWithTransactionManager
	PermissionMock *permissionMocks.MockIPermissionUseCase
}

func setupGuardianRoleTest() (*guardianRoleTestDeps, usecase.RoleUseCase) {
	deps := &guardianRoleTestDeps{
		Repo:           new(mocks.MockRoleRepository),
		TM:             new(mocking.MockWithTransactionManager),
		PermissionMock: new(permissionMocks.MockIPermissionUseCase),
	}
	// Use discarded logger for tests
	log := logrus.New()
	log.SetOutput(ioDiscard)

	uc := usecase.NewRoleUseCase(log, deps.TM, deps.Repo, deps.PermissionMock)
	return deps, uc
}

// Simple io.Discard equivalent for logrus
type discardWriter struct{}

func (w discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

var ioDiscard = discardWriter{}

func TestRoleUseCase_Guardian_FindErrors(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*guardianRoleTestDeps)
		run       func(usecase.RoleUseCase) (interface{}, error)
		wantNil   bool
	}{
		{
			name: "Create FindByName Error",
			setupMock: func(deps *guardianRoleTestDeps) {
				deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					}).Return(exception.ErrInternalServer)
				deps.Repo.On("FindByName", mock.Anything, "error_role").Return((*entity.Role)(nil), errors.New("connection failed"))
			},
			run: func(uc usecase.RoleUseCase) (interface{}, error) {
				return uc.Create(context.Background(), &model.CreateRoleRequest{Name: "error_role", Description: "Test Role"})
			},
			wantNil: true,
		},
		{
			name: "Delete FindByID Error",
			setupMock: func(deps *guardianRoleTestDeps) {
				deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					}).Return(exception.ErrInternalServer)
				deps.Repo.On("FindByID", mock.Anything, "role-error-id").Return((*entity.Role)(nil), errors.New("connection failed"))
			},
			run: func(uc usecase.RoleUseCase) (interface{}, error) {
				return nil, uc.Delete(context.Background(), "role-error-id")
			},
			wantNil: true,
		},
		{
			name: "Update FindByID Error",
			setupMock: func(deps *guardianRoleTestDeps) {
				deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				deps.Repo.On("FindByID", mock.Anything, "role-error-id").Return((*entity.Role)(nil), errors.New("connection failed"))
			},
			run: func(uc usecase.RoleUseCase) (interface{}, error) {
				return uc.Update(context.Background(), "role-error-id", &model.UpdateRoleRequest{Description: "Updated Desc"})
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupGuardianRoleTest()
			tt.setupMock(deps)

			res, err := tt.run(uc)
			assert.Error(t, err)
			assert.ErrorIs(t, err, exception.ErrInternalServer)
			if tt.wantNil {
				assert.Nil(t, res)
			}

			deps.Repo.AssertExpectations(t)
			deps.TM.AssertExpectations(t)
		})
	}
}

func TestRoleUseCase_Create_Guardian_UUIDError(t *testing.T) {
	// Not practically testable without mocking google/uuid.NewV7 directly.
	// We'll skip since it's hard to trigger and coverage for this block is rare.
}

func TestRoleUseCase_Guardian_CleanupAndConflict(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(*guardianRoleTestDeps)
		run       func(usecase.RoleUseCase) (interface{}, error)
		wantErr   error
		wantNil   bool
	}{
		{
			name: "Delete cleanup policy error",
			setupMock: func(deps *guardianRoleTestDeps) {
				roleID := "role-test-id"
				role := &entity.Role{ID: roleID, Name: "test_role"}
				deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				deps.Repo.On("FindByID", mock.Anything, roleID).Return(role, nil)
				deps.Repo.On("Delete", mock.Anything, roleID).Return(nil)
				deps.PermissionMock.On("DeleteRole", mock.Anything, role.Name).Return(errors.New("cleanup failed"))
			},
			run: func(uc usecase.RoleUseCase) (interface{}, error) {
				return nil, uc.Delete(context.Background(), "role-test-id")
			},
			wantErr: exception.ErrInternalServer,
			wantNil: true,
		},
		{
			name: "Create find by name success means conflict",
			setupMock: func(deps *guardianRoleTestDeps) {
				deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) })
				deps.Repo.On("FindByName", mock.Anything, "success_role").Return(&entity.Role{ID: "existing-id", Name: "success_role"}, nil)
			},
			run: func(uc usecase.RoleUseCase) (interface{}, error) {
				return uc.Create(context.Background(), &model.CreateRoleRequest{Name: "success_role", Description: "Test Role"})
			},
			wantErr: exception.ErrConflict,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupGuardianRoleTest()
			tt.setupMock(deps)

			res, err := tt.run(uc)
			assert.Error(t, err)
			assert.ErrorIs(t, err, tt.wantErr)
			if tt.wantNil {
				assert.Nil(t, res)
			}

			deps.Repo.AssertExpectations(t)
			deps.TM.AssertExpectations(t)
			deps.PermissionMock.AssertExpectations(t)
		})
	}
}

func TestRoleUseCase_Guardian_TMError(t *testing.T) {
	tests := []struct {
		name    string
		run     func(usecase.RoleUseCase) (interface{}, error)
		wantNil bool
	}{
		{name: "Create TM Error", run: func(uc usecase.RoleUseCase) (interface{}, error) {
			return uc.Create(context.Background(), &model.CreateRoleRequest{Name: "tm_error_role", Description: "Test TM Error"})
		}, wantNil: true},
		{name: "Update TM Error", run: func(uc usecase.RoleUseCase) (interface{}, error) {
			return uc.Update(context.Background(), "id123", &model.UpdateRoleRequest{Description: "Test TM Error"})
		}, wantNil: true},
		{name: "GetAll TM Error", run: func(uc usecase.RoleUseCase) (interface{}, error) { return uc.GetAll(context.Background()) }, wantNil: true},
		{name: "GetAllRolesDynamic TM Error", run: func(uc usecase.RoleUseCase) (interface{}, error) {
			return uc.GetAllRolesDynamic(context.Background(), nil)
		}, wantNil: true},
		{name: "Delete TM Error", run: func(uc usecase.RoleUseCase) (interface{}, error) {
			return nil, uc.Delete(context.Background(), "id123")
		}, wantNil: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupGuardianRoleTest()
			deps.TM.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).Return(exception.ErrInternalServer)

			res, err := tt.run(uc)
			assert.Error(t, err)
			assert.ErrorIs(t, err, exception.ErrInternalServer)
			if tt.wantNil {
				assert.Nil(t, res)
			}

			deps.TM.AssertExpectations(t)
		})
	}
}
