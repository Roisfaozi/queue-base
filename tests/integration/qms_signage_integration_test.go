package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/qms/model"
	"github.com/Roisfaozi/queue-base/tests/integration/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignageIntegration(t *testing.T) {
	app, db, _, cleanup := testutils.SetupIntegrationApp(t)
	defer cleanup()

	tenant := testutils.CreateTestTenant(t, db, "Tenant S")
	branch := testutils.CreateTestBranch(t, db, tenant.ID, "Branch S")
	service := testutils.CreateTestService(t, db, tenant.ID, branch.ID, "Service S", "SS")
	counter := testutils.CreateTestCounter(t, db, tenant.ID, branch.ID, "Counter S")

	// Create active tickets
	t1 := testutils.CreateTestQueueJourney(t, db, tenant.ID, branch.ID, service.ID, "SS001")
	t2 := testutils.CreateTestQueueJourney(t, db, tenant.ID, branch.ID, service.ID, "SS002")
	t3 := testutils.CreateTestQueueJourney(t, db, tenant.ID, branch.ID, service.ID, "SS003")

	// Set states
	db.Model(&t1).Updates(map[string]interface{}{"status": model.QueueStatusWaiting})
	db.Model(&t2).Updates(map[string]interface{}{"status": model.QueueStatusCalled, "counter_id": counter.ID})
	db.Model(&t3).Updates(map[string]interface{}{"status": model.QueueStatusServing, "counter_id": counter.ID})

	t.Run("Get Active Signage State", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/qms/signage/state?branch_id="+branch.ID, nil)
		req.Header.Set("X-Tenant-ID", tenant.ID)
		w := httptest.NewRecorder()

		app.Router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp struct {
			Data struct {
				CurrentlyCalled []model.QueueJourney `json:"currently_called"`
				WaitingList     []model.QueueJourney `json:"waiting_list"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

		// Check response structure
		assert.NotEmpty(t, resp.Data.CurrentlyCalled)
		assert.NotEmpty(t, resp.Data.WaitingList)

		calledFound := false
		for _, q := range resp.Data.CurrentlyCalled {
			if q.ID == t2.ID {
				calledFound = true
			}
		}
		assert.True(t, calledFound, "Called ticket should be in CurrentlyCalled")

		waitFound := false
		for _, q := range resp.Data.WaitingList {
			if q.ID == t1.ID {
				waitFound = true
			}
		}
		assert.True(t, waitFound, "Waiting ticket should be in WaitingList")
	})
}
