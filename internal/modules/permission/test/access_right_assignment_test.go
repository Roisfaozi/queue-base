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
		name         string
		category     string
		accessRights []*accessEntity.AccessRight
		repoErr      error
		setupEnforce func(*permissionTestDeps)
		wantErr      error
		wantLen      int
		assertRes    func(t *testing.T, res []model.RoleAccessRightStatus)
	}{
		{
			name:     "Success",
			category: "positive",
			accessRights: []*accessEntity.AccessRight{{ID: "ar1", Name: "Role Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/roles"}, {Method: "POST", Path: "/api/roles"}}}, {ID: "ar2", Name: "User Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/users"}, {Method: "DELETE", Path: "/api/users/:id"}}}, {ID: "ar3", Name: "Stats", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/stats"}}}},
			setupEnforce: func(deps *permissionTestDeps) {
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "GET"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/users", "GET"}).Return(true, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/users/:id", "DELETE"}).Return(false, nil)
				deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/stats", "GET"}).Return(false, nil)
			},
			wantLen: 3,
			assertRes: func(t *testing.T, res []model.RoleAccessRightStatus) {
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
		{name: "RepoError", category: "negative", repoErr: errors.New("db error"), wantErr: errors.New("db error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			deps.AccessRepo.On("GetAccessRights", mock.Anything).Return(tt.accessRights, tt.repoErr)
			if tt.setupEnforce != nil {
				tt.setupEnforce(deps)
			}
			res, err := uc.GetRoleAccessRights(context.Background(), "admin", "")
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, res)
				return
			}
			assert.NoError(t, err)
			assert.Len(t, res, tt.wantLen)
			if tt.assertRes != nil {
				tt.assertRes(t, res)
			}
		})
	}
}

func TestAssignAccessRight(t *testing.T) {
	tests := []struct {
		name     string
		category string
		access   *accessEntity.AccessRight
		repoErr  error
		setupCas func(*permissionTestDeps)
		wantErr  string
	}{
		{name: "Success", category: "positive", access: &accessEntity.AccessRight{ID: "ar1", Name: "Role Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/roles"}, {Method: "POST", Path: "/api/roles"}}}, setupCas: func(deps *permissionTestDeps) { deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "GET"}).Return(true, nil); deps.Enforcer.On("Enforce", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(false, nil); deps.Enforcer.On("AddPolicy", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(true, nil); deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil) }},
		{name: "NotFound", category: "negative", repoErr: exception.ErrNotFound, wantErr: exception.ErrNotFound.Error()},
		{name: "NoEndpoints", category: "negative", access: &accessEntity.AccessRight{ID: "ar1", Name: "Empty Access Right", Endpoints: []accessEntity.Endpoint{}}, wantErr: "no endpoints configured"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(tt.access, tt.repoErr)
			if tt.setupCas != nil {
				tt.setupCas(deps)
			}
			err := uc.AssignAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestRevokeAccessRight(t *testing.T) {
	tests := []struct {
		name     string
		category string
		access   *accessEntity.AccessRight
		repoErr  error
		setupCas func(*permissionTestDeps)
		wantErr  string
	}{
		{name: "Success", category: "positive", access: &accessEntity.AccessRight{ID: "ar1", Name: "Role Management", Endpoints: []accessEntity.Endpoint{{Method: "GET", Path: "/api/roles"}, {Method: "POST", Path: "/api/roles"}}}, setupCas: func(deps *permissionTestDeps) { deps.Enforcer.On("RemovePolicy", []interface{}{"admin", "global", "/api/roles", "GET"}).Return(true, nil); deps.Enforcer.On("RemovePolicy", []interface{}{"admin", "global", "/api/roles", "POST"}).Return(true, nil); deps.AuditUC.On("LogActivity", mock.Anything, mock.Anything).Return(nil) }},
		{name: "NotFound", category: "negative", repoErr: exception.ErrNotFound, wantErr: exception.ErrNotFound.Error()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupPermissionTest()
			deps.AccessRepo.On("GetAccessRightByID", mock.Anything, "ar1").Return(tt.access, tt.repoErr)
			if tt.setupCas != nil {
				tt.setupCas(deps)
			}
			err := uc.RevokeAccessRight(context.Background(), model.AssignAccessRightRequest{AccessRightID: "ar1", Role: "admin"})
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
