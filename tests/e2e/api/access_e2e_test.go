//go:build e2e
// +build e2e

package api

import (
	"testing"

	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func createAccessAdminAndLogin(t *testing.T, server *setup.TestServer) string {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("AccessAdmin123!"), bcrypt.DefaultCost)

	admin := f.Create(func(u *userEntity.User) {
		u.Username = "access_admin"
		u.Email = "access_admin@test.com"
		u.Password = string(hash)
	})

	server.Enforcer.AddGroupingPolicy(admin.ID, "role:superadmin", "global")
	server.Enforcer.AddPolicy("role:superadmin", "global", "*", "*")
	server.Enforcer.SavePolicy()

	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": admin.Username,
		"password": "AccessAdmin123!",
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

func TestAccessE2E_AccessRightsCRUD(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)
	var createdID string

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_CreateAccessRight",
			run: func(t *testing.T) {
				resp := server.Client.POST("/api/v1/access-rights", map[string]any{
					"name":        "manage_reports",
					"description": "Can manage reports",
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
				createdID = result.Data.ID
				assert.Equal(t, "manage_reports", result.Data.Name)
			},
		},
		{
			name: "Success_GetAllAccessRights",
			run: func(t *testing.T) {
				resp := server.Client.GET("/api/v1/access-rights", setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data struct {
						Data []struct {
							ID   string `json:"id"`
							Name string `json:"name"`
						} `json:"data"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, len(result.Data.Data), 1)
			},
		},
		{
			name: "Success_DeleteAccessRight",
			run: func(t *testing.T) {
				resp := server.Client.DELETE("/api/v1/access-rights/"+createdID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name: "Negative_DeleteNonExistent",
			run: func(t *testing.T) {
				resp := server.Client.DELETE("/api/v1/access-rights/nonexistent-id", setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestAccessE2E_EndpointsCRUD(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)
	var createdID string

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Success_CreateEndpoint",
			run: func(t *testing.T) {
				resp := server.Client.POST("/api/v1/endpoints", map[string]any{
					"path":   "/api/v1/reports",
					"method": "GET",
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 201, resp.StatusCode)

				var result struct {
					Data struct {
						ID     string `json:"id"`
						Path   string `json:"path"`
						Method string `json:"method"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				createdID = result.Data.ID
				assert.Equal(t, "/api/v1/reports", result.Data.Path)
			},
		},
		{
			name: "Success_DeleteEndpoint",
			run: func(t *testing.T) {
				resp := server.Client.DELETE("/api/v1/endpoints/"+createdID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}

func TestAccessE2E_LinkEndpointToAccessRight(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)

	resp := server.Client.POST("/api/v1/access-rights", map[string]any{
		"name": "link_test_access",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var arResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&arResult)
	accessRightID := arResult.Data.ID

	resp = server.Client.POST("/api/v1/endpoints", map[string]any{
		"path":   "/api/v1/linked-resource",
		"method": "POST",
	}, setup.WithAuth(adminToken))
	require.Equal(t, 201, resp.StatusCode)
	var epResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&epResult)
	endpointID := epResult.Data.ID

	tests := []struct {
		name    string
		payload map[string]any
		status  int
	}{
		{name: "Success_LinkEndpointToAccessRight", payload: map[string]any{"access_right_id": accessRightID, "endpoint_id": endpointID}, status: 200},
		{name: "Negative_InvalidIDs", payload: map[string]any{"access_right_id": "", "endpoint_id": ""}, status: 422},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.Client.POST("/api/v1/access-rights/link", tt.payload, setup.WithAuth(adminToken))
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}

func TestAccessE2E_DynamicSearch(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	adminToken := createAccessAdminAndLogin(t, server)

	server.Client.POST("/api/v1/access-rights", map[string]any{"name": "searchable_access_alpha"}, setup.WithAuth(adminToken))
	server.Client.POST("/api/v1/access-rights", map[string]any{"name": "searchable_access_beta"}, setup.WithAuth(adminToken))
	server.Client.POST("/api/v1/endpoints", map[string]any{"path": "/api/v1/searchable", "method": "GET"}, setup.WithAuth(adminToken))
	server.Client.POST("/api/v1/endpoints", map[string]any{"path": "/api/v1/searchable", "method": "POST"}, setup.WithAuth(adminToken))

	tests := []struct {
		name    string
		url     string
		payload map[string]any
		status  int
	}{
		{
			name:    "Success_SearchAccessRights",
			url:     "/api/v1/access-rights/search",
			status:  200,
			payload: map[string]any{"filter": map[string]any{"name": map[string]any{"type": "contains", "from": "searchable"}}},
		},
		{
			name:    "Success_SearchEndpoints",
			url:     "/api/v1/endpoints/search",
			status:  200,
			payload: map[string]any{"filter": map[string]any{"path": map[string]any{"type": "contains", "from": "searchable"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := server.Client.POST(tt.url, tt.payload, setup.WithAuth(adminToken))
			assert.Equal(t, tt.status, resp.StatusCode)
		})
	}
}
