package usecase

import (
	"context"
	"testing"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Stubs
// =============================================================================

type stubQueueHandler struct {
	registerCalled bool
	forwardCalled  bool
	registerReq    *queueModel.RegisterQueueRequest
	forwardReq     *queueModel.ForwardQueueRequest
	forwardQueueID string
	registerRes    *queueModel.QueueResponse
	forwardRes     *queueModel.QueueResponse
	registerErr    error
	forwardErr     error
}

func (s *stubQueueHandler) RegisterQueue(ctx context.Context, req *queueModel.RegisterQueueRequest) (*queueModel.QueueResponse, error) {
	s.registerCalled = true
	s.registerReq = req
	return s.registerRes, s.registerErr
}

func (s *stubQueueHandler) ForwardQueue(ctx context.Context, queueID string, req *queueModel.ForwardQueueRequest) (*queueModel.QueueResponse, error) {
	s.forwardCalled = true
	s.forwardQueueID = queueID
	s.forwardReq = req
	return s.forwardRes, s.forwardErr
}

type stubScannerAuthenticator struct {
	err error
}

func (s stubScannerAuthenticator) Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error {
	return s.err
}

type stubRelationValidator struct {
	err            error
	serviceID      string
	counterID      string
	validateCalled bool
}

func (s *stubRelationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	s.validateCalled = true
	s.serviceID = serviceID
	s.counterID = counterID
	return s.err
}

// =============================================================================
// TestScannerCheckIn
// =============================================================================

func TestScannerCheckIn(t *testing.T) {
	tests := []struct {
		name          string
		category      string
		req           *CheckInRequest
		queueHandler  *stubQueueHandler
		authenticator stubScannerAuthenticator
		validator     *stubRelationValidator
		tenantID      string
		branchID      string
		wantErr       error
		wantRes       func(t *testing.T, queueHandler *stubQueueHandler, validator *stubRelationValidator, res *CheckInResponse)
	}{
		{
			name:     "Positive_RegisterSuccess",
			category: "positive",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			queueHandler: &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}},
			validator:    &stubRelationValidator{},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantRes: func(t *testing.T, qh *stubQueueHandler, v *stubRelationValidator, res *CheckInResponse) {
				assert.Equal(t, "register", res.Action)
				assert.True(t, qh.registerCalled)
				assert.Equal(t, "service-1", qh.registerReq.ServiceID)
				assert.True(t, v.validateCalled)
				assert.Equal(t, "service-1", v.serviceID)
				assert.False(t, qh.forwardCalled)
			},
		},
		{
			name:     "Negative_InvalidCredential",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "bad",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			authenticator: stubScannerAuthenticator{err: exception.ErrUnauthorized},
			tenantID:      "t-1",
			branchID:      "b-1",
			wantErr:       exception.ErrUnauthorized,
		},
		{
			name:     "Security_AuthenticatorInternalErrorMapsToUnauthorized",
			category: "security",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "bad",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			authenticator: stubScannerAuthenticator{err: assert.AnError},
			tenantID:      "t-1",
			branchID:      "b-1",
			wantErr:       exception.ErrUnauthorized,
		},
		{
			name:     "Edge_WhitespaceAction",
			category: "edge",
			req: &CheckInRequest{
				Action:               " forward ",
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				QueueID:              "q-1",
				DestinationServiceID: "service-2",
				DestinationCounterID: "counter-2",
			},
			queueHandler: &stubQueueHandler{forwardRes: &queueModel.QueueResponse{ID: "q-1"}},
			validator:    &stubRelationValidator{},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantRes: func(t *testing.T, qh *stubQueueHandler, v *stubRelationValidator, res *CheckInResponse) {
				assert.Equal(t, "forward", res.Action)
				assert.True(t, qh.forwardCalled)
				assert.Equal(t, "q-1", qh.forwardQueueID)
				assert.True(t, v.validateCalled)
				assert.Equal(t, "counter-2", v.counterID)
			},
		},
		{
			name:     "Vulnerability_RejectsUnknownAction",
			category: "vulnerability",
			req: &CheckInRequest{
				Action:   "drop-table",
				BranchID: "b-1",
				ClientID: "client-1",
				APIKey:   "key-1",
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_InvalidRelation",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			validator: &stubRelationValidator{err: exception.ErrForbidden},
			tenantID:  "t-1",
			branchID:  "b-1",
			wantErr:   exception.ErrForbidden,
		},
		{
			name:     "Negative_ForwardPropagatesWorkflowRejection",
			category: "negative",
			req: &CheckInRequest{
				Action:               ActionForward,
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				QueueID:              "q-1",
				DestinationServiceID: "pharmacy-svc",
				DestinationCounterID: "counter-1",
			},
			validator: &stubRelationValidator{err: exception.ErrForbidden},
			tenantID:  "t-1",
			branchID:  "b-1",
			wantErr:   exception.ErrForbidden,
			wantRes: func(t *testing.T, qh *stubQueueHandler, v *stubRelationValidator, res *CheckInResponse) {
				assert.False(t, qh.forwardCalled)
				assert.True(t, v.validateCalled)
				assert.Equal(t, "pharmacy-svc", v.serviceID)
			},
		},
		{
			name:     "Negative_NilRequest",
			category: "negative",
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_MissingTenantContext",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "c-1",
				APIKey:      "k-1",
				ServiceID:   "s-1",
				PatientName: "John",
			},
			wantErr: exception.ErrBadRequest,
		},
		{
			name:     "Negative_MissingBranchContext",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "c-1",
				APIKey:      "k-1",
				ServiceID:   "s-1",
				PatientName: "John",
			},
			tenantID: "t-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_RegisterPropagatesQueueError",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "c-1",
				APIKey:      "k-1",
				ServiceID:   "s-1",
				PatientName: "John",
			},
			queueHandler: &stubQueueHandler{registerErr: exception.ErrConflict},
			validator:    &stubRelationValidator{},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantErr:      exception.ErrConflict,
		},
		{
			name:     "Negative_ForwardPropagatesQueueError",
			category: "negative",
			req: &CheckInRequest{
				Action:               ActionForward,
				BranchID:             "b-1",
				ClientID:             "c-1",
				APIKey:               "k-1",
				QueueID:              "q-1",
				DestinationServiceID: "s-2",
			},
			queueHandler: &stubQueueHandler{forwardErr: exception.ErrNotFound},
			validator:    &stubRelationValidator{},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantErr:      exception.ErrNotFound,
		},
		{
			name:     "Edge_CaseInsensitiveAction",
			category: "edge",
			req: &CheckInRequest{
				Action:      "REGISTER",
				BranchID:    "b-1",
				ClientID:    "c-1",
				APIKey:      "k-1",
				ServiceID:   "s-1",
				PatientName: "John",
			},
			queueHandler: &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}},
			validator:    &stubRelationValidator{},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantRes: func(t *testing.T, qh *stubQueueHandler, v *stubRelationValidator, res *CheckInResponse) {
				assert.Equal(t, ActionRegister, res.Action)
				assert.True(t, qh.registerCalled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qh := tt.queueHandler
			if qh == nil {
				qh = &stubQueueHandler{}
			}
			uc := NewScannerUseCase(qh, tt.authenticator, tt.validator)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.CheckIn(ctx, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.wantRes != nil {
					tt.wantRes(t, qh, tt.validator, res)
				}
				return
			}
			assert.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, qh, tt.validator, res)
			}
		})
	}
}
