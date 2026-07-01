//go:build e2e
// +build e2e

package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	apiKeyModel "github.com/Roisfaozi/queue-base/internal/modules/api_key/model"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	integrationSetup "github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func loginQueueAdmin(t *testing.T, server *setup.TestServer) (string, string, string) {
	unique := fmt.Sprintf("%d", time.Now().UnixNano())
	user := integrationSetup.CreateTestUser(t, server.DB, "queue_admin_"+unique, "queue_"+unique+"@test.com", "Password123!")
	org := &orgEntity.Organization{ID: uuid.New().String(), Name: "Queue Org", Slug: "queue-org-" + unique, OwnerID: user.ID, Status: orgEntity.OrgStatusActive}
	require.NoError(t, server.DB.Create(org).Error)
	require.NoError(t, server.DB.Create(&orgEntity.OrganizationMember{ID: uuid.New().String(), OrganizationID: org.ID, UserID: user.ID, RoleID: "role:owner", Status: orgEntity.MemberStatusActive}).Error)
	_, err := server.Enforcer.AddGroupingPolicy(user.ID, "role:superadmin", org.ID)
	require.NoError(t, err)
	_, err = server.Enforcer.AddPolicy("role:superadmin", org.ID, "*", "*")
	require.NoError(t, err)
	require.NoError(t, server.Enforcer.SavePolicy())

	loginResp := server.Client.POST("/api/v1/auth/login", map[string]string{"username": user.Username, "password": "Password123!"})
	require.Equal(t, http.StatusOK, loginResp.StatusCode)
	var loginData struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, loginResp.JSON(&loginData))
	return loginData.Data.AccessToken, org.ID, user.ID
}

func TestQMSQueueE2E_LifecycleAndScannerGuard(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer)
	}{
		{
			name:     "Positive_LifecycleAndScannerGuard",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				token, orgID, userID := loginQueueAdmin(t, server)

				createBranchResp := server.Client.POST("/api/v1/branches", map[string]any{"code": "BD", "name": "Branch Desk"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, createBranchResp.StatusCode, createBranchResp.String())
				var branchData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, createBranchResp.JSON(&branchData))

				createServiceResp := server.Client.POST("/api/v1/services", map[string]any{"code": "RG", "name": "Registration"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, createServiceResp.StatusCode, createServiceResp.String())
				var regServiceData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, createServiceResp.JSON(&regServiceData))

				createPharmacyResp := server.Client.POST("/api/v1/services", map[string]any{"code": "PH", "name": "Pharmacy", "is_pharmacy": true}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, createPharmacyResp.StatusCode, createPharmacyResp.String())
				var pharmacyData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, createPharmacyResp.JSON(&pharmacyData))

				createCounterResp := server.Client.POST("/api/v1/counters", map[string]any{"branch_id": branchData.Data.ID, "code": "C1", "name": "Counter 1"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, createCounterResp.StatusCode, createCounterResp.String())
				var counterData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, createCounterResp.JSON(&counterData))

				for _, payload := range []map[string]any{
					{"scope_type": "service", "scope_id": pharmacyData.Data.ID, "key": settingsModel.SettingKeyPharmacyFlowEnabled, "value": "true", "value_type": "boolean"},
					{"scope_type": "service", "scope_id": pharmacyData.Data.ID, "key": settingsModel.SettingKeyRequireCounterForService, "value": "true", "value_type": "boolean"},
				} {
					resp := server.Client.POST("/api/v1/settings", payload, setup.WithAuth(token), setup.WithOrg(orgID))
					require.Equal(t, http.StatusCreated, resp.StatusCode, resp.String())
				}

				queueResp := server.Client.POST("/api/v1/queues", map[string]any{"branch_id": branchData.Data.ID, "service_id": regServiceData.Data.ID, "patient_name": "Queue Patient"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, queueResp.StatusCode, queueResp.String())
				var queueData struct {
					Data struct {
						ID               string `json:"id"`
						CurrentJourneyID string `json:"current_journey_id"`
					} `json:"data"`
				}
				require.NoError(t, queueResp.JSON(&queueData))

				listByServiceResp := server.Client.GET("/api/v1/branches/"+branchData.Data.ID+"/services/"+regServiceData.Data.ID+"/queue-journeys", setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, listByServiceResp.StatusCode, listByServiceResp.String())
				assert.Contains(t, listByServiceResp.String(), "queue_id")

				forwardResp := server.Client.POST("/api/v1/queues/"+queueData.Data.ID+"/forward", map[string]any{"destination_service_id": pharmacyData.Data.ID, "destination_counter_id": counterData.Data.ID}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, forwardResp.StatusCode, forwardResp.String())
				var forwarded struct {
					Data struct {
						CurrentJourneyID string `json:"current_journey_id"`
					} `json:"data"`
				}
				require.NoError(t, forwardResp.JSON(&forwarded))
				assert.NotEqual(t, queueData.Data.CurrentJourneyID, forwarded.Data.CurrentJourneyID)

				transitionResp := server.Client.POST("/api/v1/queues/"+queueData.Data.ID+"/transition", map[string]any{"action": "call"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, transitionResp.StatusCode, transitionResp.String())

				repeatedTransitionResp := server.Client.POST("/api/v1/queues/"+queueData.Data.ID+"/transition", map[string]any{"action": "call"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusBadRequest, repeatedTransitionResp.StatusCode, repeatedTransitionResp.String())

				invalidTransitionResp := server.Client.POST("/api/v1/queues/"+queueData.Data.ID+"/transition", map[string]any{"action": "drop-table"}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusUnprocessableEntity, invalidTransitionResp.StatusCode, invalidTransitionResp.String())

				visitResp := server.Client.GET("/api/v1/queues/"+queueData.Data.ID+"/visit-journeys", setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, visitResp.StatusCode, visitResp.String())
				assert.Contains(t, visitResp.String(), "registration")
				assert.Contains(t, visitResp.String(), "forward")
				assert.Contains(t, visitResp.String(), "call")

				statsResp := server.Client.GET("/api/v1/branches/"+branchData.Data.ID+"/queue-stats", setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, statsResp.StatusCode, statsResp.String())
				assert.Contains(t, statsResp.String(), "total_queues_today")

				createAPIKeyResp := server.Client.POST("/api/v1/api-keys", apiKeyModel.CreateApiKeyRequest{Name: "scanner-key", Scopes: []string{"scanner:create"}}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, createAPIKeyResp.StatusCode, createAPIKeyResp.String())
				var apiKeyData struct {
					Data apiKeyModel.CreateApiKeyResponse `json:"data"`
				}
				require.NoError(t, createAPIKeyResp.JSON(&apiKeyData))

				createReadOnlyQueueKeyResp := server.Client.POST("/api/v1/api-keys", apiKeyModel.CreateApiKeyRequest{Name: "queue-view-key", Scopes: []string{"queue:view"}}, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, createReadOnlyQueueKeyResp.StatusCode, createReadOnlyQueueKeyResp.String())
				var queueViewKeyData struct {
					Data apiKeyModel.CreateApiKeyResponse `json:"data"`
				}
				require.NoError(t, createReadOnlyQueueKeyResp.JSON(&queueViewKeyData))

				queueCreateBlockedResp := server.Client.POST("/api/v1/queues", map[string]any{"branch_id": branchData.Data.ID, "service_id": regServiceData.Data.ID, "patient_name": "Blocked Patient"}, setup.WithOrg(orgID), setup.WithHeader("X-API-Key", queueViewKeyData.Data.Key))
				require.Equal(t, http.StatusForbidden, queueCreateBlockedResp.StatusCode, queueCreateBlockedResp.String())

				scannerForbiddenResp := server.Client.POST("/api/v1/scanner/check-in", map[string]any{"action": "forward", "branch_id": branchData.Data.ID, "queue_id": queueData.Data.ID, "destination_service_id": pharmacyData.Data.ID}, setup.WithAuth(token), setup.WithOrg(orgID), setup.WithHeader("X-Client-ID", userID), setup.WithHeader("X-API-Key", apiKeyData.Data.Key))
				require.Equal(t, http.StatusForbidden, scannerForbiddenResp.StatusCode, scannerForbiddenResp.String())

				scannerOKResp := server.Client.POST("/api/v1/scanner/check-in", map[string]any{"action": "forward", "branch_id": branchData.Data.ID, "queue_id": queueData.Data.ID, "destination_service_id": pharmacyData.Data.ID, "destination_counter_id": counterData.Data.ID}, setup.WithAuth(token), setup.WithOrg(orgID), setup.WithHeader("X-Client-ID", userID), setup.WithHeader("X-API-Key", apiKeyData.Data.Key))
				require.Equal(t, http.StatusOK, scannerOKResp.StatusCode, scannerOKResp.String())

				foreignBranch := &branchEntity.Branch{ID: uuid.New().String(), TenantID: uuid.New().String(), Code: "FX", Name: "Foreign", Status: branchEntity.BranchStatusActive}
				require.NoError(t, server.DB.Create(foreignBranch).Error)
				foreignStatsResp := server.Client.GET("/api/v1/branches/"+foreignBranch.ID+"/queue-stats", setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusForbidden, foreignStatsResp.StatusCode, foreignStatsResp.String())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setup.SetupTestServer(t)
			defer server.Cleanup()
			tt.run(t, server)
		})
	}
}
