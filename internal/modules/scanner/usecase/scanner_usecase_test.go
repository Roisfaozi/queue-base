package usecase

import (
	"context"
	"errors"
	"testing"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
)

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

func TestScannerUseCase_CheckIn_RegisterSuccess(t *testing.T) {
	queueHandler := &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}}
	validator := &stubRelationValidator{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, validator)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.CheckIn(ctx, &CheckInRequest{
		Action:      ActionRegister,
		ClientID:    "client-1",
		APIKey:      "key-1",
		ServiceID:   "service-1",
		PatientName: "John Doe",
	})

	assert.NoError(t, err)
	assert.Equal(t, "register", res.Action)
	assert.True(t, queueHandler.registerCalled)
	assert.Equal(t, "service-1", queueHandler.registerReq.ServiceID)
	assert.True(t, validator.validateCalled)
	assert.Equal(t, "service-1", validator.serviceID)
	assert.False(t, queueHandler.forwardCalled)
}

func TestScannerUseCase_CheckIn_NegativeInvalidCredential(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{err: errors.New("invalid credential")}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, ClientID: "client-1", APIKey: "bad", ServiceID: "service-1", PatientName: "John Doe"})

	assert.ErrorIs(t, err, exception.ErrUnauthorized)
	assert.False(t, queueHandler.registerCalled)
}

func TestScannerUseCase_CheckIn_EdgeWhitespaceAction(t *testing.T) {
	queueHandler := &stubQueueHandler{forwardRes: &queueModel.QueueResponse{ID: "q-1"}}
	validator := &stubRelationValidator{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, validator)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.CheckIn(ctx, &CheckInRequest{
		Action:               " forward ",
		ClientID:             "client-1",
		APIKey:               "key-1",
		QueueID:              "q-1",
		DestinationServiceID: "service-2",
		DestinationCounterID: "counter-2",
	})

	assert.NoError(t, err)
	assert.Equal(t, "forward", res.Action)
	assert.True(t, queueHandler.forwardCalled)
	assert.Equal(t, "q-1", queueHandler.forwardQueueID)
	assert.True(t, validator.validateCalled)
	assert.Equal(t, "counter-2", validator.counterID)
}

func TestScannerUseCase_CheckIn_SecurityRejectsUnknownAction(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: "drop-table", ClientID: "client-1", APIKey: "key-1"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
	assert.False(t, queueHandler.registerCalled)
	assert.False(t, queueHandler.forwardCalled)
}

func TestScannerUseCase_CheckIn_NegativeInvalidRelation(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	validator := &stubRelationValidator{err: exception.ErrForbidden}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, validator)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, ClientID: "client-1", APIKey: "key-1", ServiceID: "service-1", PatientName: "John Doe"})

	assert.ErrorIs(t, err, exception.ErrForbidden)
	assert.False(t, queueHandler.registerCalled)
}

func TestScannerUseCase_CheckIn_ForwardPropagatesWorkflowRejection(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	validator := &stubRelationValidator{err: exception.ErrForbidden}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, validator)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionForward, ClientID: "client-1", APIKey: "key-1", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"})

	assert.ErrorIs(t, err, exception.ErrForbidden)
	assert.False(t, queueHandler.forwardCalled)
	assert.True(t, validator.validateCalled)
	assert.Equal(t, "pharmacy-svc", validator.serviceID)
}

func TestScannerUseCase_CheckIn_NegativeNilRequest(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, nil)
	assert.ErrorIs(t, err, exception.ErrBadRequest)
	assert.False(t, queueHandler.registerCalled)
	assert.False(t, queueHandler.forwardCalled)
}

func TestScannerUseCase_CheckIn_NegativeMissingTenantContext(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, nil)

	_, err := uc.CheckIn(context.Background(), &CheckInRequest{Action: ActionRegister, ClientID: "c-1", APIKey: "k-1", ServiceID: "s-1", PatientName: "John"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
	assert.False(t, queueHandler.registerCalled)
}

func TestScannerUseCase_CheckIn_NegativeMissingBranchContext(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, ClientID: "c-1", APIKey: "k-1", ServiceID: "s-1", PatientName: "John"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
	assert.False(t, queueHandler.registerCalled)
}

func TestScannerUseCase_CheckIn_RegisterPropagatesQueueError(t *testing.T) {
	queueHandler := &stubQueueHandler{registerErr: exception.ErrConflict}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, &stubRelationValidator{})
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, ClientID: "c-1", APIKey: "k-1", ServiceID: "s-1", PatientName: "John"})
	assert.ErrorIs(t, err, exception.ErrConflict)
	assert.True(t, queueHandler.registerCalled)
}

func TestScannerUseCase_CheckIn_ForwardPropagatesQueueError(t *testing.T) {
	queueHandler := &stubQueueHandler{forwardErr: exception.ErrNotFound}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, &stubRelationValidator{})
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionForward, ClientID: "c-1", APIKey: "k-1", QueueID: "q-1", DestinationServiceID: "s-2"})
	assert.ErrorIs(t, err, exception.ErrNotFound)
	assert.True(t, queueHandler.forwardCalled)
}

func TestScannerUseCase_CheckIn_SecurityCaseInsensitiveAction(t *testing.T) {
	queueHandler := &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{}, &stubRelationValidator{})
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.CheckIn(ctx, &CheckInRequest{Action: "REGISTER", ClientID: "c-1", APIKey: "k-1", ServiceID: "s-1", PatientName: "John"})
	assert.NoError(t, err)
	assert.Equal(t, ActionRegister, res.Action)
	assert.True(t, queueHandler.registerCalled)
}
