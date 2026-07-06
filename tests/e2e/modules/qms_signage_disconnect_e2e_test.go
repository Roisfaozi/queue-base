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

// TestSignageDisconnectReconnectE2E tests the signage reconnect after disconnect behavior.
// Implementation note: if reconnection state restoration is not yet implemented,
// this test will serve as a placeholder marking the gap.
func TestSignageDisconnectReconnectE2E(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Cleanup()

	tenantID := uuid.New().String()
	branchID := uuid.New().String()

	// Setup data
	server.DB.Exec("INSERT INTO organizations (id, code, name, slug, owner_id, status, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", tenantID, "sig-disc", "Signage Disconnect", "sig-disc", "system", "active")
	server.DB.Exec("INSERT INTO branches (id, tenant_id, code, name, status, deleted_at) VALUES (?, ?, ?, ?, ?, 0)", branchID, tenantID, "BR-SIG", "Signage Branch", "active")

	time.Sleep(100 * time.Millisecond)

	t.Run("Signage Reconnects After Disconnect", func(t *testing.T) {
		// Check that signage GET endpoint is reachable
		resp := server.Client.GET("/api/v1/qms/signage/state?branch_id="+branchID, func(r *http.Request) {
			r.Header.Set("X-Organization-ID", tenantID)
		})

		require.Equal(t, http.StatusOK, resp.StatusCode)

		// The initial state should be empty branch state (no tickets yet)
		var resData struct {
			Data struct {
				CurrentlyCalled []entity.QueueJourney `json:"currently_called"`
				WaitingList     []entity.QueueJourney `json:"waiting_list"`
			} `json:"data"`
		}
		err := resp.JSON(&resData)
		require.NoError(t, err)

		assert.Empty(t, resData.Data.CurrentlyCalled)
		assert.Empty(t, resData.Data.WaitingList)
	})

	// TODO: When SSE/WebSocket reconnection logic is implemented for signage,
	// add test for:
	// 1. Client connects via SSE/WS, gets current state as initial snapshot
	// 2. Client disconnects and reconnects after state changes
	// 3. Client receives full current state on reconnection
	// 4. Verify no duplicate events or missed state transitions
}
