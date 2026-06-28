//go:build e2e
// +build e2e

package api

import (
	"testing"

	roleEntity "github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func createRoleAdminAndLogin(t *testing.T, server *setup.TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("RoleAdmin123!"), bcrypt.DefaultCost)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "role_admin"
		u.Email = "role_admin@test.com"
		u.Password = string(hash)
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "RoleAdmin123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	return loginRes.Data.AccessToken
}

func TestRoleE2E_CreateRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_CreateRole",
			run: func(t *testing.T) {
				roleName := "TestRole_" + uuid.New().String()[:8]
				resp := server.Client.POST("/api/v1/roles", map[string]any{
					"name":        roleName,
					"description": "A test role",
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 201, resp.StatusCode)

				var result struct {
					Data struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.Equal(t, roleName, result.Data.Name)
			},
		},
		{
			name: "Negative_DuplicateRoleName",
			run: func(t *testing.T) {
				roleName := "DuplicateRole"
				server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: roleName})

				resp := server.Client.POST("/api/v1/roles", map[string]any{"name": roleName}, setup.WithAuth(adminToken))
				assert.Equal(t, 409, resp.StatusCode)
			},
		},
		{
			name: "Negative_EmptyName",
			run: func(t *testing.T) {
				resp := server.Client.POST("/api/v1/roles", map[string]any{"name": ""}, setup.WithAuth(adminToken))
				assert.Equal(t, 422, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRoleE2E_DeleteRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_DeleteRole",
			run: func(t *testing.T) {
				roleToDelete := &roleEntity.Role{ID: uuid.New().String(), Name: "RoleToDelete"}
				server.DB.Create(roleToDelete)

				resp := server.Client.DELETE("/api/v1/roles/"+roleToDelete.ID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name: "Negative_DeleteNonExistent",
			run: func(t *testing.T) {
				resp := server.Client.DELETE("/api/v1/roles/nonexistent-role-id", setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRoleE2E_GetAllRoles(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Role_List_1"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "Role_List_2"})

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_GetAllRoles",
			run: func(t *testing.T) {
				resp := server.Client.GET("/api/v1/roles", setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(result.Data), 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRoleE2E_UpdateRole(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)
	roleToUpdate := &roleEntity.Role{ID: uuid.New().String(), Name: "RoleToUpdate", Description: "Original"}
	server.DB.Create(roleToUpdate)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_UpdateRole",
			run: func(t *testing.T) {
				resp := server.Client.PUT("/api/v1/roles/"+roleToUpdate.ID, map[string]any{"description": "Updated Description"}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						Description string `json:"description"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.Equal(t, "Updated Description", result.Data.Description)
			},
		},
		{
			name: "Negative_UpdateNonExistent",
			run: func(t *testing.T) {
				resp := server.Client.PUT("/api/v1/roles/nonexistent-role-id", map[string]any{"description": "Updated Description"}, setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestRoleE2E_DynamicSearch(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createRoleAdminAndLogin(t, server)

	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "SearchableRole_Alpha"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "SearchableRole_Beta"})
	server.DB.Create(&roleEntity.Role{ID: uuid.New().String(), Name: "OtherRole"})

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_SearchByName",
			run: func(t *testing.T) {
				resp := server.Client.POST("/api/v1/roles/search", map[string]any{
					"filter": map[string]any{
						"name": map[string]any{"type": "contains", "from": "Searchable"},
					},
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data []struct {
						Name string `json:"name"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(result.Data), 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
