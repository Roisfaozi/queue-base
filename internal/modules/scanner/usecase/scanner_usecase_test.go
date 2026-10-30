package usecase_test

import (
	"context"
	"strings"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	scannerMocks "github.com/Roisfaozi/queue-base/internal/modules/scanner/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestScannerAuditLogging(t *testing.T) {
	t.Run("Edge_Register_EmitsAuditAndSurvivesFailure", func(t *testing.T) {
		qh := scannerMocks.NewMockQueueHandler(t)
		qh.On("RegisterQueue", mock.Anything, mock.Anything).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()

		auth := scannerMocks.NewMockAuthenticator(t)
		auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "k-1").Return(nil).Once()

		val := scannerMocks.NewMockRelationValidator(t)
		val.On("Validate", mock.Anything, "t-1", "b-1", "svc-1", "").Return(nil).Once()

		var loggedReq auditModel.CreateAuditLogRequest
		audit := scannerMocks.NewMockAuditLogger(t)
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			loggedReq = req
			return true
		})).Return(assert.AnError).Once()

		uc := usecase.NewScannerUseCase(qh, auth, val, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		ctx = authcontext.WithUserID(ctx, "u-1")

		res, err := uc.CheckIn(ctx, &usecase.CheckInRequest{
			Action:      usecase.ActionRegister,
			BranchID:    "b-1",
			ClientID:    "c-1",
			APIKey:      "k-1",
			ServiceID:   "svc-1",
			PatientName: "John",
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		values, ok := loggedReq.NewValues.(map[string]string)
		assert.True(t, ok)
		assert.Equal(t, "SCANNER_REGISTER", loggedReq.Action)
		assert.Equal(t, "scanner", loggedReq.Entity)
		assert.Equal(t, "t-1", loggedReq.OrganizationID)
		assert.Equal(t, "u-1", loggedReq.UserID)
		assert.Equal(t, "b-1", values["branch_id"])
	})

	t.Run("Vulnerability_Forward_DoesNotLeakAPIKey", func(t *testing.T) {
		qh := scannerMocks.NewMockQueueHandler(t)
		qh.On("ForwardQueue", mock.Anything, "q-1", mock.Anything).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()

		auth := scannerMocks.NewMockAuthenticator(t)
		auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "super-secret").Return(nil).Once()

		val := scannerMocks.NewMockRelationValidator(t)
		val.On("Validate", mock.Anything, "t-1", "b-1", "svc-2", "ctr-2").Return(nil).Once()

		var loggedReq auditModel.CreateAuditLogRequest
		audit := scannerMocks.NewMockAuditLogger(t)
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			loggedReq = req
			return true
		})).Return(nil).Once()

		uc := usecase.NewScannerUseCase(qh, auth, val, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.CheckIn(ctx, &usecase.CheckInRequest{
			Action:               usecase.ActionForward,
			BranchID:             "b-1",
			ClientID:             "c-1",
			APIKey:               "super-secret",
			QueueID:              "q-1",
			DestinationServiceID: "svc-2",
			DestinationCounterID: "ctr-2",
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		values, ok := loggedReq.NewValues.(map[string]string)
		assert.True(t, ok)
		assert.NotContains(t, values, "api_key")
		assert.Equal(t, "SCANNER_FORWARD", loggedReq.Action)
		assert.Equal(t, "q-1", loggedReq.EntityID)
	})

	t.Run("Vulnerability_Register_DoesNotLeakPatientData", func(t *testing.T) {
		qh := scannerMocks.NewMockQueueHandler(t)
		qh.On("RegisterQueue", mock.Anything, mock.Anything).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()

		auth := scannerMocks.NewMockAuthenticator(t)
		auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "k-1").Return(nil).Once()

		val := scannerMocks.NewMockRelationValidator(t)
		val.On("Validate", mock.Anything, "t-1", "b-1", "svc-1", "").Return(nil).Once()

		var loggedReq auditModel.CreateAuditLogRequest
		audit := scannerMocks.NewMockAuditLogger(t)
		audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
			loggedReq = req
			return true
		})).Return(nil).Once()

		uc := usecase.NewScannerUseCase(qh, auth, val, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.CheckIn(ctx, &usecase.CheckInRequest{
			Action:      usecase.ActionRegister,
			BranchID:    "b-1",
			ClientID:    "c-1",
			APIKey:      "k-1",
			ServiceID:   "svc-1",
			PatientID:   "p-1",
			PatientName: "Sensitive Patient",
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		values, ok := loggedReq.NewValues.(map[string]string)
		assert.True(t, ok)
		assert.NotContains(t, values, "patient_id")
		assert.NotContains(t, values, "patient_name")
	})

	t.Run("Vulnerability_RejectsBranchMismatch", func(t *testing.T) {
		qh := scannerMocks.NewMockQueueHandler(t)
		auth := scannerMocks.NewMockAuthenticator(t)
		val := scannerMocks.NewMockRelationValidator(t)
		audit := scannerMocks.NewMockAuditLogger(t)

		uc := usecase.NewScannerUseCase(qh, auth, val, audit)
		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.CheckIn(ctx, &usecase.CheckInRequest{
			Action:      usecase.ActionRegister,
			BranchID:    "b-2",
			ClientID:    "c-1",
			APIKey:      "k-1",
			ServiceID:   "svc-1",
			PatientName: "John",
		})
		assert.ErrorIs(t, err, exception.ErrForbidden)
		assert.Nil(t, res)
	})
}

func TestScannerCheckIn(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      *usecase.CheckInRequest
		setup    func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator)
		tenantID string
		branchID string
		wantErr  error
		assert   func(t *testing.T, res *usecase.CheckInResponse)
	}{
		{
			name:     "Positive_RegisterSuccess",
			category: "positive",
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "service-1", "").Return(nil).Once()
				qh.On("RegisterQueue", mock.Anything, mock.MatchedBy(func(req *queueModel.RegisterQueueRequest) bool {
					return req.ServiceID == "service-1" && req.BranchID == "b-1"
				})).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			assert: func(t *testing.T, res *usecase.CheckInResponse) {
				assert.Equal(t, "register", res.Action)
			},
		},
		{
			name:     "Negative_InvalidCredential",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "bad",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "bad").Return(exception.ErrUnauthorized).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrUnauthorized,
		},
		{
			name:     "Security_AuthenticatorInternalErrorMapsToUnauthorized",
			category: "security",
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "bad",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "bad").Return(assert.AnError).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrUnauthorized,
		},
		{
			name:     "Edge_WhitespaceAction",
			category: "edge",
			req: &usecase.CheckInRequest{
				Action:               " forward ",
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				QueueID:              "q-1",
				DestinationServiceID: "service-2",
				DestinationCounterID: "counter-2",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "service-2", "counter-2").Return(nil).Once()
				qh.On("ForwardQueue", mock.Anything, "q-1", mock.MatchedBy(func(req *queueModel.ForwardQueueRequest) bool {
					return req.DestinationServiceID == "service-2" && req.DestinationCounterID == "counter-2"
				})).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			assert: func(t *testing.T, res *usecase.CheckInResponse) {
				assert.Equal(t, "forward", res.Action)
			},
		},
		{
			name:     "Negative_ForwardMissingDestinationService",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionForward,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				QueueID:     "q-1",
				PatientName: "John",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_ForwardMissingQueueID",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:               usecase.ActionForward,
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				DestinationServiceID: "service-2",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Edge_ForwardMissingCounterIDAllowed",
			category: "edge",
			req: &usecase.CheckInRequest{
				Action:               usecase.ActionForward,
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				QueueID:              "q-1",
				DestinationServiceID: "service-2",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "service-2", "").Return(nil).Once()
				qh.On("ForwardQueue", mock.Anything, "q-1", mock.MatchedBy(func(req *queueModel.ForwardQueueRequest) bool {
					return req.DestinationServiceID == "service-2" && req.DestinationCounterID == ""
				})).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			assert: func(t *testing.T, res *usecase.CheckInResponse) {
				assert.Equal(t, usecase.ActionForward, res.Action)
			},
		},
		{
			name:     "Negative_RegisterMissingService",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				PatientName: "John Doe",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_RejectsUnknownAction",
			category: "vulnerability",
			req: &usecase.CheckInRequest{
				Action:   "drop-table",
				BranchID: "b-1",
				ClientID: "client-1",
				APIKey:   "key-1",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_InvalidRelation",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
				BranchID:    "b-1",
				ClientID:    "client-1",
				APIKey:      "key-1",
				ServiceID:   "service-1",
				PatientName: "John Doe",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "service-1", "").Return(exception.ErrForbidden).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrForbidden,
		},
		{
			name:     "Negative_ForwardPropagatesWorkflowRejection",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:               usecase.ActionForward,
				BranchID:             "b-1",
				ClientID:             "client-1",
				APIKey:               "key-1",
				QueueID:              "q-1",
				DestinationServiceID: "pharmacy-svc",
				DestinationCounterID: "counter-1",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "client-1", "key-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "pharmacy-svc", "counter-1").Return(exception.ErrForbidden).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrForbidden,
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
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
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
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
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
			req: &usecase.CheckInRequest{
				Action:      usecase.ActionRegister,
				BranchID:    "b-1",
				ClientID:    "c-1",
				APIKey:      "k-1",
				ServiceID:   "s-1",
				PatientName: "John",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "k-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "s-1", "").Return(nil).Once()
				qh.On("RegisterQueue", mock.Anything, mock.Anything).Return(nil, exception.ErrConflict).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrConflict,
		},
		{
			name:     "Negative_ForwardPropagatesQueueError",
			category: "negative",
			req: &usecase.CheckInRequest{
				Action:               usecase.ActionForward,
				BranchID:             "b-1",
				ClientID:             "c-1",
				APIKey:               "k-1",
				QueueID:              "q-1",
				DestinationServiceID: "s-2",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "k-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "s-2", "").Return(nil).Once()
				qh.On("ForwardQueue", mock.Anything, "q-1", mock.Anything).Return(nil, exception.ErrNotFound).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Edge_CaseInsensitiveAction",
			category: "edge",
			req: &usecase.CheckInRequest{
				Action:      "REGISTER",
				BranchID:    "b-1",
				ClientID:    "c-1",
				APIKey:      "k-1",
				ServiceID:   "s-1",
				PatientName: "John",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "k-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "s-1", "").Return(nil).Once()
				qh.On("RegisterQueue", mock.Anything, mock.Anything).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			assert: func(t *testing.T, res *usecase.CheckInResponse) {
				assert.Equal(t, usecase.ActionRegister, res.Action)
			},
		},
		{
			name:     "Edge_CaseInsensitiveForwardAction",
			category: "edge",
			req: &usecase.CheckInRequest{
				Action:               "FORWARD",
				BranchID:             "b-1",
				ClientID:             "c-1",
				APIKey:               "k-1",
				QueueID:              "q-1",
				DestinationServiceID: "s-2",
			},
			setup: func(t *testing.T, qh *scannerMocks.MockQueueHandler, auth *scannerMocks.MockAuthenticator, val *scannerMocks.MockRelationValidator) {
				auth.On("Authenticate", mock.Anything, "t-1", "b-1", "c-1", "k-1").Return(nil).Once()
				val.On("Validate", mock.Anything, "t-1", "b-1", "s-2", "").Return(nil).Once()
				qh.On("ForwardQueue", mock.Anything, "q-1", mock.MatchedBy(func(req *queueModel.ForwardQueueRequest) bool {
					return req.DestinationServiceID == "s-2"
				})).Return(&queueModel.QueueResponse{ID: "q-1"}, nil).Once()
			},
			tenantID: "t-1",
			branchID: "b-1",
			assert: func(t *testing.T, res *usecase.CheckInResponse) {
				assert.Equal(t, usecase.ActionForward, res.Action)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qh := scannerMocks.NewMockQueueHandler(t)
			auth := scannerMocks.NewMockAuthenticator(t)
			val := scannerMocks.NewMockRelationValidator(t)

			if tt.setup != nil {
				tt.setup(t, qh, auth, val)
			}

			uc := usecase.NewScannerUseCase(qh, auth, val)

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
				return
			}
			assert.NoError(t, err)
			if tt.assert != nil {
				tt.assert(t, res)
			}
		})
	}
}

// silence unused warning
var _ = strings.ToUpper
