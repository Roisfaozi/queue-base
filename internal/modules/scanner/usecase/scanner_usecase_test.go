package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
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

type stubAuditLogger struct {
	entries []auditModel.CreateAuditLogRequest
	err     error
}

func (s *stubAuditLogger) LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error {
	s.entries = append(s.entries, req)
	return s.err
}

func (s *stubRelationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	s.validateCalled = true
	s.serviceID = serviceID
	s.counterID = counterID
	return s.err
}

func TestScannerAuditLogging(t *testing.T) {
	t.Run("Register_EmitsAuditAndSurvivesFailure", func(t *testing.T) {
		qh := &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}}
		audit := &stubAuditLogger{err: assert.AnError}
		uc := NewScannerUseCase(qh, stubScannerAuthenticator{}, &stubRelationValidator{}, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		ctx = authcontext.WithUserID(ctx, "u-1")

		res, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, BranchID: "b-1", ClientID: "c-1", APIKey: "k-1", ServiceID: "svc-1", PatientName: "John"})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, audit.entries, 1)
		values, ok := audit.entries[0].NewValues.(map[string]string)
		assert.True(t, ok)
		assert.Equal(t, "SCANNER_REGISTER", audit.entries[0].Action)
		assert.Equal(t, "scanner", audit.entries[0].Entity)
		assert.Equal(t, "t-1", audit.entries[0].OrganizationID)
		assert.Equal(t, "u-1", audit.entries[0].UserID)
		assert.Equal(t, "b-1", values["branch_id"])
	})

	t.Run("Forward_DoesNotLeakAPIKey", func(t *testing.T) {
		qh := &stubQueueHandler{forwardRes: &queueModel.QueueResponse{ID: "q-1"}}
		audit := &stubAuditLogger{}
		uc := NewScannerUseCase(qh, stubScannerAuthenticator{}, &stubRelationValidator{}, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionForward, BranchID: "b-1", ClientID: "c-1", APIKey: "super-secret", QueueID: "q-1", DestinationServiceID: "svc-2", DestinationCounterID: "ctr-2"})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		require := assert.New(t)
		require.Len(audit.entries, 1)
		values, ok := audit.entries[0].NewValues.(map[string]string)
		require.True(ok)
		require.NotContains(values, "api_key")
		require.Equal("SCANNER_FORWARD", audit.entries[0].Action)
		require.Equal("q-1", audit.entries[0].EntityID)
	})

	t.Run("Register_DoesNotLeakPatientData", func(t *testing.T) {
		qh := &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}}
		audit := &stubAuditLogger{}
		uc := NewScannerUseCase(qh, stubScannerAuthenticator{}, &stubRelationValidator{}, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, BranchID: "b-1", ClientID: "c-1", APIKey: "k-1", ServiceID: "svc-1", PatientID: "p-1", PatientName: "Sensitive Patient"})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, audit.entries, 1)
		values, ok := audit.entries[0].NewValues.(map[string]string)
		assert.True(t, ok)
		assert.NotContains(t, values, "patient_id")
		assert.NotContains(t, values, "patient_name")
	})

	t.Run("RejectsBranchMismatch", func(t *testing.T) {
		uc := NewScannerUseCase(&stubQueueHandler{}, stubScannerAuthenticator{}, &stubRelationValidator{}, &stubAuditLogger{})
		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, BranchID: "b-2", ClientID: "c-1", APIKey: "k-1", ServiceID: "svc-1", PatientName: "John"})
		assert.ErrorIs(t, err, exception.ErrForbidden)
		assert.Nil(t, res)
	})
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
			name:     "Negative_ForwardMissingDestinationService",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionForward,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				QueueID:     "q-1",
				PatientName: "John",
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_ForwardMissingQueueID",
			category: "negative",
			req: &CheckInRequest{
				Action:               ActionForward,
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				DestinationServiceID: "service-2",
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_RegisterMissingService",
			category: "negative",
			req: &CheckInRequest{
				Action:      ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				PatientName: "John Doe",
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
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
		{
			name:     "Edge_CaseInsensitiveForwardAction",
			category: "edge",
			req: &CheckInRequest{
				Action:               "FORWARD",
				BranchID:             "b-1",
				ClientID:             "c-1",
				APIKey:               "k-1",
				QueueID:              "q-1",
				DestinationServiceID: "s-2",
			},
			queueHandler: &stubQueueHandler{forwardRes: &queueModel.QueueResponse{ID: "q-1"}},
			validator:    &stubRelationValidator{},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantRes: func(t *testing.T, qh *stubQueueHandler, v *stubRelationValidator, res *CheckInResponse) {
				assert.Equal(t, ActionForward, res.Action)
				assert.True(t, qh.forwardCalled)
				assert.Equal(t, "s-2", qh.forwardReq.DestinationServiceID)
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
