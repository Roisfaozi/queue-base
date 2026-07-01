package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/scanner/model"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type stubScannerControllerUseCase struct {
	called bool
	last   *usecase.CheckInRequest
	res    *usecase.CheckInResponse
	err    error
}

func (s *stubScannerControllerUseCase) CheckIn(ctx context.Context, req *usecase.CheckInRequest) (*usecase.CheckInResponse, error) {
	s.called = true
	s.last = req
	return s.res, s.err
}

func TestScannerController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("CheckIn", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			headers  map[string]string
			setup    func() *stubScannerControllerUseCase
			tenantID string
			branchID string
			wantCode int
			assert   func(t *testing.T, uc *stubScannerControllerUseCase)
		}{
			{
				name:    "Positive_UsesHeaders",
				reqBody: model.CheckInRequest{Action: model.CheckInRequest{Action: "register"}.Action, BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "service-1", PatientName: "John Doe"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{res: &usecase.CheckInResponse{Action: usecase.ActionRegister}}
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.True(t, uc.called)
					assert.Equal(t, "client-1", uc.last.ClientID)
					assert.Equal(t, "key-1", uc.last.APIKey)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uc.last.BranchID)
				},
			},
			{
				name:    "Positive_ForwardsPharmacyPayload",
				reqBody: model.CheckInRequest{Action: "forward", BranchID: "550e8400-e29b-41d4-a716-446655440000", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{res: &usecase.CheckInResponse{Action: usecase.ActionForward}}
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.True(t, uc.called)
					assert.Equal(t, "forward", uc.last.Action)
					assert.Equal(t, "pharmacy-svc", uc.last.DestinationServiceID)
					assert.Equal(t, "counter-1", uc.last.DestinationCounterID)
				},
			},
			{
				name:    "Negative_RejectsMissingHeaders",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "service-1", PatientName: "John Doe"},
				headers: map[string]string{},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{}
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.False(t, uc.called)
				},
			},
			{
				name:    "Negative_RejectsInvalidAction",
				reqBody: map[string]interface{}{"action": "drop-table", "branch_id": "550e8400-e29b-41d4-a716-446655440000", "service_id": "s-1", "patient_name": "John"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{}
				},
				wantCode: http.StatusUnprocessableEntity,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.False(t, uc.called)
				},
			},
			{
				name:    "Negative_RejectsEmptyBody",
				reqBody: map[string]interface{}{},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{}
				},
				wantCode: http.StatusUnprocessableEntity,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.False(t, uc.called)
				},
			},
			{
				name:    "Security_PropagatesUnauthorized",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "s-1", PatientName: "John"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "bad-key"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{err: exception.ErrUnauthorized}
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusUnauthorized,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.True(t, uc.called)
				},
			},
			{
				name:    "Security_PropagatesWorkflowForbidden",
				reqBody: model.CheckInRequest{Action: "forward", BranchID: "550e8400-e29b-41d4-a716-446655440000", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{err: exception.ErrForbidden}
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusForbidden,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.True(t, uc.called)
				},
			},
			{
				name:    "Negative_RejectsMalformedJSON",
				reqBody: nil,
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func() *stubScannerControllerUseCase {
					return &stubScannerControllerUseCase{}
				},
				wantCode: http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubScannerControllerUseCase) {
					assert.False(t, uc.called)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewScannerController(uc, validator.New())
				router := gin.New()

				if tt.tenantID != "" || tt.branchID != "" {
					router.Use(func(c *gin.Context) {
						ctx := c.Request.Context()
						if tt.tenantID != "" {
							ctx = database.SetOrganizationContext(ctx, tt.tenantID)
						}
						if tt.branchID != "" {
							ctx = database.SetBranchContext(ctx, tt.branchID)
						}
						c.Request = c.Request.WithContext(ctx)
						c.Next()
					})
				}

				router.POST("/scanner/check-in", controller.CheckIn)

				var req *http.Request
				if tt.name == "Negative_RejectsMalformedJSON" {
					req, _ = http.NewRequest("POST", "/scanner/check-in", bytes.NewBufferString("{"))
				} else {
					body, err := json.Marshal(tt.reqBody)
					if err != nil {
						t.Fatalf("Failed to marshal body: %v", err)
					}
					req, _ = http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
				}

				for k, v := range tt.headers {
					req.Header.Set(k, v)
				}

				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})
}
