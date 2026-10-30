package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/scanner/model"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScannerController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Positive_CheckIn", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			headers  map[string]string
			setup    func(t *testing.T) *mocks.MockScannerUseCase
			tenantID string
			branchID string
			wantCode int
		}{
			{
				name:    "Positive_UsesHeaders",
				reqBody: model.CheckInRequest{Action: model.CheckInRequest{Action: "register"}.Action, BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "service-1", PatientName: "John Doe"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					m := mocks.NewMockScannerUseCase(t)
					m.On("CheckIn", mock.Anything, mock.MatchedBy(func(req *usecase.CheckInRequest) bool {
						return req != nil && req.ClientID == "client-1" && req.APIKey == "key-1" && req.BranchID == "550e8400-e29b-41d4-a716-446655440000"
					})).Return(&usecase.CheckInResponse{Action: usecase.ActionRegister}, nil).Once()
					return m
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusOK,
			},
			{
				name:    "Positive_ForwardsPharmacyPayload",
				reqBody: model.CheckInRequest{Action: "forward", BranchID: "550e8400-e29b-41d4-a716-446655440000", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					m := mocks.NewMockScannerUseCase(t)
					m.On("CheckIn", mock.Anything, mock.MatchedBy(func(req *usecase.CheckInRequest) bool {
						return req != nil && req.Action == "forward" && req.DestinationServiceID == "pharmacy-svc" && req.DestinationCounterID == "counter-1"
					})).Return(&usecase.CheckInResponse{Action: usecase.ActionForward}, nil).Once()
					return m
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusOK,
			},
			{
				name:    "Negative_RejectsMissingHeaders",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "service-1", PatientName: "John Doe"},
				headers: map[string]string{},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusBadRequest,
			},
			{
				name:    "Negative_RejectsInvalidAction",
				reqBody: map[string]interface{}{"action": "drop-table", "branch_id": "550e8400-e29b-41d4-a716-446655440000", "service_id": "s-1", "patient_name": "John"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
			{
				name:    "Negative_RejectsEmptyBody",
				reqBody: map[string]interface{}{},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
			{
				name:    "Negative_RejectsInvalidBranchID",
				reqBody: map[string]interface{}{"action": "register", "branch_id": "bad", "service_id": "s-1", "patient_name": "John"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
			{
				name:    "Negative_RejectsMissingAPIKeyHeader",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "s-1", PatientName: "John"},
				headers: map[string]string{"X-Client-ID": "client-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusBadRequest,
			},
			{
				name:    "Negative_RejectsMissingClientHeader",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "s-1", PatientName: "John"},
				headers: map[string]string{"X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusBadRequest,
			},
			{
				name:    "Negative_RejectsMissingBranchID",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "", ServiceID: "s-1", PatientName: "John"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusUnprocessableEntity,
			},
			{
				name:    "Security_PropagatesUnauthorized",
				reqBody: model.CheckInRequest{Action: "register", BranchID: "550e8400-e29b-41d4-a716-446655440000", ServiceID: "s-1", PatientName: "John"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "bad-key"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					m := mocks.NewMockScannerUseCase(t)
					m.On("CheckIn", mock.Anything, mock.Anything).Return(nil, exception.ErrUnauthorized).Once()
					return m
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusUnauthorized,
			},
			{
				name:    "Security_PropagatesWorkflowForbidden",
				reqBody: model.CheckInRequest{Action: "forward", BranchID: "550e8400-e29b-41d4-a716-446655440000", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"},
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					m := mocks.NewMockScannerUseCase(t)
					m.On("CheckIn", mock.Anything, mock.Anything).Return(nil, exception.ErrForbidden).Once()
					return m
				},
				tenantID: "t-1",
				branchID: "b-1",
				wantCode: http.StatusForbidden,
			},
			{
				name:    "Negative_RejectsMalformedJSON",
				reqBody: nil,
				headers: map[string]string{"X-Client-ID": "client-1", "X-API-Key": "key-1"},
				setup: func(t *testing.T) *mocks.MockScannerUseCase {
					return mocks.NewMockScannerUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
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
				uc.AssertExpectations(t)
			})
		}
	})
}
