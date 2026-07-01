package test

import (
	"context"
	"errors"
	"testing"

	accessEntity "github.com/Roisfaozi/queue-base/internal/modules/access/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/model"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetRoleAccessRights(t *testing.T) {
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
				accessRights := []*accessEntity.AccessRight{
					{ID: "ar1", Name: "Role Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/roles"}, {Method: "POST", Path: "/api/roles"}}},
					{ID: "ar2", Name: "User Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/users"}, {Method: "DELETE", Path: "/api/users/:id"}}},
					{ID: "ar3", Name: "Stats", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/stats"}}},
				}
				deps.AccessRepo.On("GetAccessRights", mock.Anything).Return(accessRights, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "GET"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/users", "GET"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/users/:id", "DELETE"}).Return(false, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/stats", "GET"}).Return(false, nil)

				res, err := uc.GetRoleAccessRights(context.Background(), "admin", "")
				assert.NoError(t, err)
				assert.Len(t, res, 3)

				assert.Equal(t, "Role Management", res[0].Name)
				assert.True(t, res[0].Assigned)
				assert.False(t, res[0].Partial)

				assert.Equal(t, "User Management", res[1].Name)
				assert.False(t, res[1].Assigned)
				assert.True(t, res[1].Partial)

				assert.Equal(t, "Stats", res[2].Name)
				assert.False(t, res[2].Assigned)
				assert.False(t, res[2].Partial)
			},
		},
		{
			name:     "Negative_RepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.AccessRepo.On("GetAccessRights", mock.Anything).Return(nil, errors.New("db error"))

				res, err := uc.GetRoleAccessRights(context.Background(), "admin", "")
				assert.Error(t, err)
				assert.EqualError(t, err, "failed to fetch access rights: db error")
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

func TestAssignAccessRight(t *testing.T) {
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
				access := &accessEntity.AccessRight{ID: "ar1", Name: "Role Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/roles"}, {Method: "POST", Path: "/api/roles"}}}
				deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(access, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "GET"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(false, nil)
				deps.Enforcer.On("AddPolicy", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(true, nil)
				deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

				err := uc.AssignAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(nil, exception.ErrNotFound)

				err := uc.AssignAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_NoEndpoints",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				access := &accessEntity.AccessRight{ID: "ar1", Name: "Empty Access Right", Endpoints: []accessEntity.Endpoint{}}
				deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(access, nil)

				err := uc.AssignAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "no endpoints configured")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestRevokeAccessRight(t *testing.T) {
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
				access := &accessEntity.AccessRight{ID: "ar1", Name: "Role Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/roles"}, {Method: "POST", Path: "/api/roles"}}}
				deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(access, nil)
				deps.Enforcer.On("RemovePolicy", []interface{}{"admin", "global", "/api/roles", "GET"}).Return(true, nil)
				deps.Enforcer.On("RemovePolicy", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(true, nil)
				deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil)

				err := uc.RevokeAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupPermissionTest()
				deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(nil, exception.ErrNotFound)

				err := uc.RevokeAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
