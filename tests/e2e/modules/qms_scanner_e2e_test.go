//go:build e2e
// +build e2e

package modules

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/scanner/model"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/tests/e2e/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScannerE2E_APIKeyCheckInFlow(t *testing.T) {
	server := setup.SetupTestServer(t)
	defer server.Server.Close()

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	regServiceID := uuid.New().String()
	pharmacyServiceID := uuid.New().String()
	counterID := uuid.New().String()
	clientID := uuid.New().String()
	rawAPIKey := "sk_live_test_" + uuid.New().String()

	// 2. Setup Data
	// Organizations and Branches
	server.DB.Exec("INSERT INTO organizations (id, name, slug, owner_id, status, deleted_at) VALUES (?, ?, ?, ?, ?, 0)", tenantID, "Scanner Tenant", "scanner-tenant-"+tenantID[:6], "system", "active")
	server.DB.Exec("INSERT INTO branches (id, tenant_id, code, name, status, deleted_at) VALUES (?, ?, ?, ?, ?, 0)", branchID, tenantID, "BR-SCAN", "Scanner Branch", "active")

	// Services
	server.DB.Exec("INSERT INTO services (id, tenant_id, code, name, status, is_pharmacy, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", regServiceID, tenantID, "REG", "Registration", "active", false)
	server.DB.Exec("INSERT INTO services (id, tenant_id, code, name, status, is_pharmacy, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", pharmacyServiceID, tenantID, "PHA", "Pharmacy", "active", true)

	// Counters
	server.DB.Exec("INSERT INTO counters (id, tenant_id, branch_id, code, name, status, deleted_at) VALUES (?, ?, ?, ?, ?, ?, 0)", counterID, tenantID, branchID, "C-PHA", "Pharmacy Counter", "active")

	// Settings
	server.DB.Exec("INSERT INTO settings (id, tenant_id, scope_type, scope_id, `key`, value, value_type, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), tenantID, "service", pharmacyServiceID, settingsModel.SettingKeyPharmacyFlowEnabled, "true", "boolean", true)

	// API Key and User
	server.DB.Exec("INSERT INTO users (id, username, email, password, status, deleted_at) VALUES (?, ?, ?, ?, ?, ?)", clientID, "scanner-client", "scanner@example.com", "hash", "active", 0)

	// Add user as organization member with "active" status
	server.DB.Exec("INSERT INTO organization_members (id, organization_id, user_id, role_id, status) VALUES (?, ?, ?, ?, ?)", uuid.New().String(), tenantID, clientID, "admin", "active")

	// Add Casbin policy for the scanner check-in endpoint
	server.DB.Exec("INSERT INTO casbin_rule (ptype, v0, v1, v2, v3) VALUES ('p', ?, ?, '/api/v1/scanner/check-in', 'POST')", clientID, tenantID)
	// Refresh the policy manually
	if server.Enforcer != nil {
		server.Enforcer.LoadPolicy()
	}

	hash := sha256.Sum256([]byte(strings.TrimPrefix(rawAPIKey, "sk_live_"))) // actualKey is stripped of "sk_live_" during auth.
	keyHashHex := hex.EncodeToString(hash[:])
	server.DB.Exec("INSERT INTO api_keys (id, key_hash, organization_id, user_id, name, scopes, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), keyHashHex, tenantID, clientID, "Scanner CheckIn Key", `["*"]`, true)

	// Give the server time to start up and cache to be ready (Redis delay mitigation if any)
	time.Sleep(100 * time.Millisecond)

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_RegisterViaScanner",
			category: "positive",
			run: func(t *testing.T) {
				req := model.CheckInRequest{
					Action:      usecase.ActionRegister,
					BranchID:    branchID,
					ServiceID:   regServiceID,
					PatientName: "John Doe Scanner",
				}

				resp := server.Client.POST("/api/v1/scanner/check-in", req, func(r *http.Request) {
					r.Header.Set("X-Client-ID", clientID)
					r.Header.Set("X-API-Key", rawAPIKey)
					r.Header.Set("X-Organization-ID", tenantID)
				})

				require.Equal(t, http.StatusOK, resp.StatusCode)

				var resData struct {
					Data struct {
						Action string `json:"action"`
						Queue  struct {
							ID          string `json:"id"`
							PatientName string `json:"patient_name"`
						} `json:"queue"`
					} `json:"data"`
				}
				err := resp.JSON(&resData)
				require.NoError(t, err)

				assert.Equal(t, usecase.ActionRegister, resData.Data.Action)
				assert.NotEmpty(t, resData.Data.Queue.ID)
				assert.Equal(t, "John Doe Scanner", resData.Data.Queue.PatientName)
			},
		},
		{
			name:     "Failure_MissingHeaders",
			category: "negative",
			run: func(t *testing.T) {
				req := model.CheckInRequest{
					Action:      usecase.ActionRegister,
					BranchID:    branchID,
					ServiceID:   regServiceID,
					PatientName: "No Headers",
				}

				resp := server.Client.POST("/api/v1/scanner/check-in", req) // No headers added
				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name:     "Failure_InvalidAPIKey",
			category: "negative",
			run: func(t *testing.T) {
				req := model.CheckInRequest{
					Action:      usecase.ActionRegister,
					BranchID:    branchID,
					ServiceID:   regServiceID,
					PatientName: "Invalid Key",
				}

				resp := server.Client.POST("/api/v1/scanner/check-in", req, func(r *http.Request) {
					r.Header.Set("X-Client-ID", clientID)
					r.Header.Set("X-API-Key", "sk_live_invalid_key")
				})

				assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
			},
		},
		{
			name:     "Success_ForwardViaScanner",
			category: "positive",
			run: func(t *testing.T) {
				// First register to get a Queue ID
				regReq := model.CheckInRequest{
					Action:      usecase.ActionRegister,
					BranchID:    branchID,
					ServiceID:   regServiceID,
					PatientName: "To Be Forwarded",
				}
				regResp := server.Client.POST("/api/v1/scanner/check-in", regReq, func(r *http.Request) {
					r.Header.Set("X-Client-ID", clientID)
					r.Header.Set("X-API-Key", rawAPIKey)
					r.Header.Set("X-Organization-ID", tenantID)
				})
				require.Equal(t, http.StatusOK, regResp.StatusCode)

				var regResData struct {
					Data struct {
						Queue struct {
							ID string `json:"id"`
						} `json:"queue"`
					} `json:"data"`
				}
				_ = regResp.JSON(&regResData)
				queueID := regResData.Data.Queue.ID
				require.NotEmpty(t, queueID)

				// Now forward it
				forwardReq := model.CheckInRequest{
					Action:               usecase.ActionForward,
					BranchID:             branchID,
					QueueID:              queueID,
					DestinationServiceID: pharmacyServiceID,
					DestinationCounterID: counterID,
				}

				forwardResp := server.Client.POST("/api/v1/scanner/check-in", forwardReq, func(r *http.Request) {
					r.Header.Set("X-Client-ID", clientID)
					r.Header.Set("X-API-Key", rawAPIKey)
					r.Header.Set("X-Organization-ID", tenantID)
				})

				require.Equal(t, http.StatusOK, forwardResp.StatusCode)

				var fwdResData struct {
					Data struct {
						Action string `json:"action"`
						Queue  struct {
							ID string `json:"id"`
						} `json:"queue"`
					} `json:"data"`
				}
				err := forwardResp.JSON(&fwdResData)
				require.NoError(t, err)

				assert.Equal(t, usecase.ActionForward, fwdResData.Data.Action)
				assert.Equal(t, queueID, fwdResData.Data.Queue.ID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
