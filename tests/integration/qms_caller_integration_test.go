//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/qms/model"
	"github.com/Roisfaozi/queue-base/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallerActionsIntegration(t *testing.T) {
	app, db, _, cleanup := testutils.SetupIntegrationApp(t)
	defer cleanup()

	// 1. Setup tenant, branch, service, counter, caller user
	tenant := testutils.CreateTestTenant(t, db, "Tenant A")
	branch := testutils.CreateTestBranch(t, db, tenant.ID, "Branch 1")
	service := testutils.CreateTestService(t, db, tenant.ID, branch.ID, "Service A", "SA")
	counter := testutils.CreateTestCounter(t, db, tenant.ID, branch.ID, "Counter 1")
	callerUser := testutils.CreateTestUser(t, db, tenant.ID, branch.ID, "caller1@test.com", "caller")

	// 2. Setup token
	token, _ := app.JwtService.GenerateToken(callerUser.ID, tenant.ID, branch.ID, "caller")
	authHeader := "Bearer " + token

	// 3. Create a ticket to act on
	ticket := testutils.CreateTestQueueJourney(t, db, tenant.ID, branch.ID, service.ID, "SA001")

	// 4. Test endpoints
	t.Run("Call Next Ticket", func(t *testing.T) {
		reqBody := `{"counter_id":"` + counter.ID + `","service_ids":["` + service.ID + `"]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/qms/caller/call-next", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("X-Tenant-ID", tenant.ID)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data model.QueueJourney `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
		assert.Equal(t, ticket.ID, resp.Data.ID)
		assert.Equal(t, model.QueueStatusCalled, resp.Data.Status)
		assert.Equal(t, counter.ID, *resp.Data.CounterID)

		// Verify ticket status in DB
		var updatedTicket model.QueueJourney
		err := db.First(&updatedTicket, "id = ?", ticket.ID).Error
		require.NoError(t, err)
		assert.Equal(t, model.QueueStatusCalled, updatedTicket.Status)
	})

	t.Run("Serve Ticket", func(t *testing.T) {
		reqBody := `{"action":"serve"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/qms/caller/journeys/"+ticket.ID+"/action", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("X-Tenant-ID", tenant.ID)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify DB
		var updatedTicket model.QueueJourney
		db.First(&updatedTicket, "id = ?", ticket.ID)
		assert.Equal(t, model.QueueStatusServing, updatedTicket.Status)
	})

	t.Run("Complete Ticket", func(t *testing.T) {
		reqBody := `{"action":"complete"}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/qms/caller/journeys/"+ticket.ID+"/action", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("X-Tenant-ID", tenant.ID)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify DB
		var updatedTicket model.QueueJourney
		db.First(&updatedTicket, "id = ?", ticket.ID)
		assert.Equal(t, model.QueueStatusCompleted, updatedTicket.Status)
	})
}
