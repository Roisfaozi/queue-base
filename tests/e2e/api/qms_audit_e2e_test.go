//go:build e2e
// +build e2e

package api

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQMSAuditE2E_Visibility(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, server *setup.TestServer)
	}{
		{
			name:     "Positive_QMSCRUDProducesAuditLogs",
			category: "positive",
			run: func(t *testing.T, server *setup.TestServer) {
				token, orgID, _ := loginQueueAdmin(t, server)

				// 1. Create Branch
				branchPayload := map[string]any{"code": fmt.Sprintf("AUD-%d", time.Now().UnixMilli()), "name": "Audit Branch"}
				branchResp := server.Client.POST("/api/v1/branches", branchPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, branchResp.StatusCode, branchResp.String())
				var branchData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, branchResp.JSON(&branchData))

				// 2. Create Service
				svcPayload := map[string]any{"code": "SA", "name": "Service Audit"}
				svcResp := server.Client.POST("/api/v1/services", svcPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, svcResp.StatusCode, svcResp.String())
				var svcData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, svcResp.JSON(&svcData))

				// 3. Create Counter
				ctrPayload := map[string]any{"branch_id": branchData.Data.ID, "code": "CA", "name": "Counter Audit"}
				ctrResp := server.Client.POST("/api/v1/counters", ctrPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, ctrResp.StatusCode, ctrResp.String())
				var ctrData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, ctrResp.JSON(&ctrData))

				// 4. Create Branch Service
				bsPayload := map[string]any{"service_id": svcData.Data.ID}
				bsResp := server.Client.POST("/api/v1/branches/"+branchData.Data.ID+"/services", bsPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusCreated, bsResp.StatusCode, bsResp.String())
				var bsData struct {
					Data struct {
						ID string `json:"id"`
					} `json:"data"`
				}
				require.NoError(t, bsResp.JSON(&bsData))

				// 5. Update Branch Service
				bsUpdatePayload := map[string]any{"custom_name": "Updated BS"}
				bsUpdateResp := server.Client.PUT("/api/v1/branches/"+branchData.Data.ID+"/services/"+bsData.Data.ID, bsUpdatePayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, bsUpdateResp.StatusCode, bsUpdateResp.String())

				// 6. Delete Branch Service
				bsDelResp := server.Client.DELETE("/api/v1/branches/"+branchData.Data.ID+"/services/"+bsData.Data.ID, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusNoContent, bsDelResp.StatusCode, bsDelResp.String())

				// Wait for Async Flush
				time.Sleep(2 * time.Second)

				// 7. Search Audit Logs for Service creation
				auditSearchPayload := map[string]any{
					"filter": map[string]any{
						"entity_id": map[string]any{"type": "equals", "from": svcData.Data.ID},
					},
				}
				auditResp := server.Client.POST("/api/v1/audit-logs/search", auditSearchPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, auditResp.StatusCode, auditResp.String())

				assert.Contains(t, auditResp.String(), "SERVICE_CREATE")

				// 8. Search Audit Logs for Counter creation
				auditCtrPayload := map[string]any{
					"filter": map[string]any{
						"entity_id": map[string]any{"type": "equals", "from": ctrData.Data.ID},
					},
				}
				auditCtrResp := server.Client.POST("/api/v1/audit-logs/search", auditCtrPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, auditCtrResp.StatusCode, auditCtrResp.String())
				assert.Contains(t, auditCtrResp.String(), "COUNTER_CREATE")

				// 9. Search Audit Logs for Branch Service lifecycle
				auditBsPayload := map[string]any{
					"filter": map[string]any{
						"entity_id": map[string]any{"type": "equals", "from": bsData.Data.ID},
					},
				}
				auditBsResp := server.Client.POST("/api/v1/audit-logs/search", auditBsPayload, setup.WithAuth(token), setup.WithOrg(orgID))
				require.Equal(t, http.StatusOK, auditBsResp.StatusCode, auditBsResp.String())
				assert.Contains(t, auditBsResp.String(), "BRANCH_SERVICE_CREATE")
				assert.Contains(t, auditBsResp.String(), "BRANCH_SERVICE_UPDATE")
				assert.Contains(t, auditBsResp.String(), "BRANCH_SERVICE_DELETE")
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
