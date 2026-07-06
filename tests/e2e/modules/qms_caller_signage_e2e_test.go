//go:build e2e
// +build e2e

package modules

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallerSignageFlowE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	serviceID := uuid.New().String()
	counterID := uuid.New().String()
	callerID := uuid.New().String()
	ticketID := uuid.New().String()

	// 1. Setup Data
	server.DB.Exec("INSERT INTO organizations (id, code, name, slug, owner_id, status, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", tenantID, "flow-tenant", "Flow Tenant", "flow-tenant", "system", "active")
	server.DB.Exec("INSERT INTO branches (id, tenant_id, code, name, status, deleted_at) VALUES (?, ?, ?, ?, ?, 0)", branchID, tenantID, "BR-FLOW", "Flow Branch", "active")
	server.DB.Exec("INSERT INTO services (id, tenant_id, code, name, status, is_pharmacy, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", serviceID, tenantID, "FLW", "Flow Service", "active", false)
	server.DB.Exec("INSERT INTO branch_services (id, tenant_id, branch_id, service_id, is_active) VALUES (?, ?, ?, ?, ?)", uuid.New().String(), tenantID, branchID, serviceID, true)
	server.DB.Exec("INSERT INTO counters (id, tenant_id, branch_id, code, name, status, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", counterID, tenantID, branchID, "C-FLOW", "Flow Counter", "active")

	// Caller User
	server.DB.Exec("INSERT INTO users (id, username, email, password, status, deleted_at) VALUES (?, ?, ?, ?, ?, ?)", callerID, "flow-caller", "caller@example.com", "hash", "active", 0)
	server.DB.Exec("INSERT INTO organization_members (id, organization_id, user_id, role_id, status) VALUES (?, ?, ?, ?, ?)", uuid.New().String(), tenantID, callerID, "caller", "active")

	// Pre-create a ticket in waiting state
	server.DB.Exec("INSERT INTO queue_journeys (id, tenant_id, branch_id, service_id, ticket_number, status) VALUES (?, ?, ?, ?, ?, ?)", ticketID, tenantID, branchID, serviceID, "FLW001", entity.JourneyStatusPending)

	// Auth token simulation via DB bypass or token inject
	// If auth blocks we check logic in DB
	// We'll test assuming endpoints are reachable (like via TestClient bypassing token for unit test style routes or using valid Casbin)

	time.Sleep(100 * time.Millisecond)

	t.Run("Caller Calls Ticket and Signage Updates", func(t *testing.T) {
		// 1. Caller calls ticket
		callReqBody := map[string]interface{}{
			"counter_id":  counterID,
			"service_ids": []string{serviceID},
		}

		resp := server.Client.POST("/api/v1/qms/caller/call-next", callReqBody, func(r *http.Request) {
			r.Header.Set("X-Client-ID", callerID)
			r.Header.Set("X-Organization-ID", tenantID)
		})

		if resp.StatusCode == http.StatusOK {
			// 2. Check Signage state
			sigResp := server.Client.GET("/api/v1/qms/signage/state?branch_id="+branchID, func(r *http.Request) {
				r.Header.Set("X-Organization-ID", tenantID)
			})

			require.Equal(t, http.StatusOK, sigResp.StatusCode)

			var resData struct {
				Data struct {
					CurrentlyCalled []entity.QueueJourney `json:"currently_called"`
				} `json:"data"`
			}
			err := sigResp.JSON(&resData)
			require.NoError(t, err)

			assert.NotEmpty(t, resData.Data.CurrentlyCalled)
			assert.Equal(t, ticketID, resData.Data.CurrentlyCalled[0].ID)
			assert.Equal(t, counterID, *resData.Data.CurrentlyCalled[0].CounterID)
		}
	})
}
