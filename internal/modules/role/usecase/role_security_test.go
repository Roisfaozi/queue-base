package usecase_test

import (
	"context"
	"testing"

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

// MockTransactionManager is a local mock for testing
type MockTransactionManager struct {
	mock.Mock
}

func (m *MockTransactionManager) WithinTransaction(ctx context.Context, fn func(context.Context) error) error {
	// Simple implementation that executes the function directly
	return fn(ctx)
}

func setupSecurityRoleTest() (*mocks.MockRoleRepository, *permissionMocks.MockIPermissionUseCase, usecase.RoleUseCase) {
	mockRepo := new(mocks.MockRoleRepository)
	mockPerm := new(permissionMocks.MockIPermissionUseCase)
	logger := logrus.New()
	mockTM := new(MockTransactionManager)

	uc := usecase.NewRoleUseCase(logger, mockTM, mockRepo, mockPerm)
	return mockRepo, mockPerm, uc
}

func TestDeleteRole_SuperadminProtection(t *testing.T) {
	tests := []struct {
		name         string
		roleID       string
		roleName     string
		setupDelete  bool
		wantErr      error
		assertDelete bool
	}{
		{name: "forbid exact superadmin", roleID: "role-superadmin-id", roleName: "role:superadmin", wantErr: exception.ErrForbidden, assertDelete: false},
		{name: "allow case variation", roleID: "role-fake-superadmin", roleName: "Role:SuperAdmin", setupDelete: true, assertDelete: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, perm, uc := setupSecurityRoleTest()
			role := &entity.Role{ID: tt.roleID, Name: tt.roleName}
			repo.On("FindByID", mock.Anything, tt.roleID).Return(role, nil)
			if tt.setupDelete {
				repo.On("Delete", mock.Anything, tt.roleID).Return(nil)
				perm.On("DeleteRole", mock.Anything, tt.roleName).Return(nil)
			}

			err := uc.Delete(context.Background(), tt.roleID)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
				return
			}

			assert.NoError(t, err)
			if tt.assertDelete {
				repo.AssertCalled(t, "Delete", mock.Anything, tt.roleID)
			}
		})
	}
}

func TestUpdateRole_SuperadminProtection(t *testing.T) {
	tests := []struct {
		name   string
		roleID string
		role   *entity.Role
		body   *model.UpdateRoleRequest
	}{
		{name: "update allowed for exact superadmin", roleID: "role-superadmin-id", role: &entity.Role{ID: "role-superadmin-id", Name: "role:superadmin"}, body: &model.UpdateRoleRequest{Description: "Updated description"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, _, uc := setupSecurityRoleTest()
			repo.On("FindByID", mock.Anything, tt.roleID).Return(tt.role, nil)
			repo.On("Update", mock.Anything, tt.role).Return(nil)

			_, err := uc.Update(context.Background(), tt.roleID, tt.body)
			assert.NoError(t, err)
		})
	}
}
