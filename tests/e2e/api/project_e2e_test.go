//go:build e2e
// +build e2e

package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/Roisfaozi/queue-base/tests/fixtures"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func setupProjectE2E(t *testing.T, server *setup.TestServer) (token string, orgID string) {
	f := fixtures.NewUserFactory(server.DB)
	hash, _ := bcrypt.GenerateFromPassword([]byte("ProjPass123!"), bcrypt.DefaultCost)

	uniqueSuffix := fmt.Sprintf("%d", time.Now().UnixNano())

	f.Create(func(u *userEntity.User) {
		u.Username = "proj_user_" + uniqueSuffix
		u.Email = "proj_" + uniqueSuffix + "@test.com"
		u.Password = string(hash)
	})

	// Login
	resp := server.Client.POST("/api/v1/auth/login", map[string]any{
		"username": "proj_user_" + uniqueSuffix,
		"password": "ProjPass123!",
	})
	require.Equal(t, 200, resp.StatusCode)

	var loginRes struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	resp.JSON(&loginRes)
	token = loginRes.Data.AccessToken

	// Create an organization (required for tenant-scoped project routes)
	orgResp := server.Client.POST("/api/v1/organizations", map[string]string{
		"name": "Project Test Org " + uniqueSuffix,
		"slug": "project-org-" + uniqueSuffix,
	}, setup.WithAuth(token))
	require.Equal(t, 201, orgResp.StatusCode)

	var orgResult struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	orgResp.JSON(&orgResult)
	orgID = orgResult.Data.ID

	return token, orgID
}

func TestProjectE2E_CRUD_Lifecycle(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, orgID := setupProjectE2E(t, server)
	var projectID string

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "1. Create Project",
			category: "positive",
			run: func(t *testing.T) {
				payload := map[string]string{
					"name":   "Test Project Alpha",
					"domain": "alpha.example.com",
				}

				resp := server.Client.POST("/api/v1/projects", payload,
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)

				if resp.StatusCode != 201 {
					t.Logf("Create Response: %s", resp.String())
				}
				require.Equal(t, 201, resp.StatusCode)

				var result struct {
					Data struct {
						ID             string `json:"id"`
						Name           string `json:"name"`
						Domain         string `json:"domain"`
						Status         string `json:"status"`
						OrganizationID string `json:"organization_id"`
					} `json:"data"`
				}
				err := resp.JSON(&result)
				require.NoError(t, err)
				assert.NotEmpty(t, result.Data.ID)
				assert.Equal(t, "Test Project Alpha", result.Data.Name)
				assert.Equal(t, "alpha.example.com", result.Data.Domain)
				assert.Equal(t, "active", result.Data.Status)
				projectID = result.Data.ID
			},
		},
		{
			name:     "2. Get All Projects",
			category: "positive",
			run: func(t *testing.T) {
				resp := server.Client.GET("/api/v1/projects",
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data []struct {
						ID   string `json:"id"`
						Name string `json:"name"`
					} `json:"data"`
				}
				resp.JSON(&result)
				assert.GreaterOrEqual(t, len(result.Data), 1)
			},
		},
		{
			name:     "3. Get Project By ID",
			category: "positive",
			run: func(t *testing.T) {
				resp := server.Client.GET("/api/v1/projects/"+projectID,
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)
				assert.Equal(t, 200, resp.StatusCode)

				var result struct {
					Data struct {
						ID     string `json:"id"`
						Name   string `json:"name"`
						Domain string `json:"domain"`
					} `json:"data"`
				}
				resp.JSON(&result)
				assert.Equal(t, projectID, result.Data.ID)
				assert.Equal(t, "Test Project Alpha", result.Data.Name)
			},
		},
		{
			name:     "4. Update Project",
			category: "positive",
			run: func(t *testing.T) {
				payload := map[string]string{
					"name":   "Updated Project Alpha",
					"domain": "updated-alpha.example.com",
				}
				resp := server.Client.PUT("/api/v1/projects/"+projectID, payload,
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)
				assert.Equal(t, 200, resp.StatusCode)

				getResp := server.Client.GET("/api/v1/projects/"+projectID,
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)
				var result struct {
					Data struct {
						Name   string `json:"name"`
						Domain string `json:"domain"`
					} `json:"data"`
				}
				getResp.JSON(&result)
				assert.Equal(t, "Updated Project Alpha", result.Data.Name)
				assert.Equal(t, "updated-alpha.example.com", result.Data.Domain)
			},
		},
		{
			name:     "5. Delete Project",
			category: "positive",
			run: func(t *testing.T) {
				resp := server.Client.DELETE("/api/v1/projects/"+projectID,
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)
				assert.Equal(t, 200, resp.StatusCode)

				getResp := server.Client.GET("/api/v1/projects/"+projectID,
					setup.WithAuth(token),
					setup.WithOrg(orgID),
				)
				assert.Equal(t, 404, getResp.StatusCode)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// createTestProject creates a project for testing get/update/delete endpoints in stateless TDT
func createTestProject(t *testing.T, server *setup.TestServer, token, orgID string) string {
	payload := map[string]string{
		"name":   "Precreated Project",
		"domain": "pre.example.com",
	}
	resp := server.Client.POST("/api/v1/projects", payload,
		setup.WithAuth(token),
		setup.WithOrg(orgID),
	)
	require.Equal(t, 201, resp.StatusCode)

	var result struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	resp.JSON(&result)
	return result.Data.ID
}

func TestProjectE2E_Stateless(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	token, orgID := setupProjectE2E(t, server)
	// Create one project beforehand for get-by-id tests
	_ = createTestProject(t, server, token, orgID)

	tests := []struct {
		name           string
		endpoint       string
		method         string
		payload        any
		token          string
		orgID          string
		expectedStatus int
		validateFunc   func(t *testing.T, resp *setup.Response)
	}{
		{
			name:           "Success_CreateMultipleProjects",
			endpoint:       "/api/v1/projects",
			method:         http.MethodPost,
			payload:        map[string]string{"name": "Multiple 1", "domain": "m1.example.com"},
			token:          token,
			orgID:          orgID,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Error_CreateWithoutAuth",
			endpoint:       "/api/v1/projects",
			method:         http.MethodPost,
			payload:        map[string]string{"name": "Unauth Project", "domain": "unauth.com"},
			token:          "",
			orgID:          orgID,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Error_GetAllWithoutAuth",
			endpoint:       "/api/v1/projects",
			method:         http.MethodGet,
			token:          "",
			orgID:          orgID,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Error_GetNonexistentProject",
			endpoint:       "/api/v1/projects/nonexistent-id",
			method:         http.MethodGet,
			token:          token,
			orgID:          orgID,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []setup.RequestOption
			if tt.token != "" {
				opts = append(opts, setup.WithAuth(tt.token))
			}
			if tt.orgID != "" {
				opts = append(opts, setup.WithOrg(tt.orgID))
			}

			var resp *setup.Response
			switch tt.method {
			case http.MethodGet:
				resp = server.Client.GET(tt.endpoint, opts...)
			case http.MethodPost:
				resp = server.Client.POST(tt.endpoint, tt.payload, opts...)
			default:
				t.Fatalf("unsupported method %s", tt.method)
			}

			assert.Equal(t, tt.expectedStatus, resp.StatusCode)
			if tt.validateFunc != nil {
				tt.validateFunc(t, resp)
			}
		})
	}
}
