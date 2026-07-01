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
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer, adminToken string, createdID *string)
	}{
		{
			name:     "Success_CreateAccessRight",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string, createdID *string) {
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
				*createdID = result.Data.ID
				assert.Equal(t, "manage_reports", result.Data.Name)
			},
		},
		{
			name:     "Success_GetAllAccessRights",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string, createdID *string) {
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
			name:     "Success_DeleteAccessRight",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string, createdID *string) {
				// We need a createdID, let's create one first
				respCreate := server.Client.POST("/api/v1/access-rights", map[string]any{
					"name": "manage_reports_del",
				}, setup.WithAuth(adminToken))
				var result struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				respCreate.JSON(&result)

				resp := server.Client.DELETE("/api/v1/access-rights/"+result.Data.ID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name:     "Negative_DeleteNonExistent",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer, adminToken string, createdID *string) {
				resp := server.Client.DELETE("/api/v1/access-rights/nonexistent-id", setup.WithAuth(adminToken))
				assert.Equal(t, 404, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, server, adminToken, &createdID)
		})
	}
}

func TestAccessE2E_EndpointsCRUD(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer, adminToken string)
	}{
		{
			name:     "Success_CreateEndpoint",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string) {
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
				assert.Equal(t, "/api/v1/reports", result.Data.Path)

				// Also delete it
				delResp := server.Client.DELETE("/api/v1/endpoints/"+result.Data.ID, setup.WithAuth(adminToken))
				assert.Equal(t, 200, delResp.StatusCode)
			},
		},
		{
			name:     "Negative_DeleteNonExistent",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer, adminToken string) {
				resp := server.Client.DELETE("/api/v1/endpoints/nonexistent-id", setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setup.SetupTestServer(t)
			defer server.Cleanup()
			adminToken := createAccessAdminAndLogin(t, server)
			tt.run(t, server, adminToken)
		})
	}
}

func TestAccessE2E_LinkEndpointToAccessRight(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer, adminToken string)
	}{
		{
			name:     "Success_LinkEndpointToAccessRight",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string) {
				arResp := server.Client.POST("/api/v1/access-rights", map[string]any{
					"name": "link_test_access",
				}, setup.WithAuth(adminToken))
				require.Equal(t, 201, arResp.StatusCode)
				var arResult struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				arResp.JSON(&arResult)

				epResp := server.Client.POST("/api/v1/endpoints", map[string]any{
					"path":   "/api/v1/linked-resource",
					"method": "POST",
				}, setup.WithAuth(adminToken))
				require.Equal(t, 201, epResp.StatusCode)
				var epResult struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				epResp.JSON(&epResult)

				resp := server.Client.POST("/api/v1/access-rights/link", map[string]any{
					"access_right_id": arResult.Data.ID,
					"endpoint_id":     epResult.Data.ID,
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name:     "Negative_InvalidIDs",
			category: "negative",
			run: func(t *testing.T, server *setup.TestServer, adminToken string) {
				resp := server.Client.POST("/api/v1/access-rights/link", map[string]any{
					"access_right_id": "",
					"endpoint_id":     "",
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 422, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setup.SetupTestServer(t)
			defer server.Cleanup()
			adminToken := createAccessAdminAndLogin(t, server)
			tt.run(t, server, adminToken)
		})
	}
}

func TestAccessE2E_DynamicSearch(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer, adminToken string)
	}{
		{
			name:     "Success_SearchAccessRights",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string) {
				server.Client.POST("/api/v1/access-rights", map[string]any{"name": "searchable_access_alpha"}, setup.WithAuth(adminToken))
				server.Client.POST("/api/v1/access-rights", map[string]any{"name": "searchable_access_beta"}, setup.WithAuth(adminToken))

				resp := server.Client.POST("/api/v1/access-rights/search", map[string]any{
					"filter": map[string]any{"name": map[string]any{"type": "contains", "from": "searchable"}},
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
		{
			name:     "Success_SearchEndpoints",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer, adminToken string) {
				server.Client.POST("/api/v1/endpoints", map[string]any{"path": "/api/v1/searchable1", "method": "GET"}, setup.WithAuth(adminToken))
				server.Client.POST("/api/v1/endpoints", map[string]any{"path": "/api/v1/searchable2", "method": "POST"}, setup.WithAuth(adminToken))

				resp := server.Client.POST("/api/v1/endpoints/search", map[string]any{
					"filter": map[string]any{"path": map[string]any{"type": "contains", "from": "searchable"}},
				}, setup.WithAuth(adminToken))
				assert.Equal(t, 200, resp.StatusCode)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setup.SetupTestServer(t)
			defer server.Cleanup()
			adminToken := createAccessAdminAndLogin(t, server)
			tt.run(t, server, adminToken)
		})
	}
}
