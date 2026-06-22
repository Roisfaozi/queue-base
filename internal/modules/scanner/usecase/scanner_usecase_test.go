package usecase

import (
	"context"
	"errors"
	"testing"

	queueModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
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

func TestScannerUseCase_CheckIn_RegisterSuccess(t *testing.T) {
	queueHandler := &stubQueueHandler{registerRes: &queueModel.QueueResponse{ID: "q-1"}}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{})
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
	assert.False(t, queueHandler.forwardCalled)
}

func TestScannerUseCase_CheckIn_NegativeInvalidCredential(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{err: errors.New("invalid credential")})
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: ActionRegister, ClientID: "client-1", APIKey: "bad", ServiceID: "service-1", PatientName: "John Doe"})

	assert.ErrorIs(t, err, exception.ErrUnauthorized)
	assert.False(t, queueHandler.registerCalled)
}

func TestScannerUseCase_CheckIn_EdgeWhitespaceAction(t *testing.T) {
	queueHandler := &stubQueueHandler{forwardRes: &queueModel.QueueResponse{ID: "q-1"}}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{})
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
}

func TestScannerUseCase_CheckIn_SecurityRejectsUnknownAction(t *testing.T) {
	queueHandler := &stubQueueHandler{}
	uc := NewScannerUseCase(queueHandler, stubScannerAuthenticator{})
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.CheckIn(ctx, &CheckInRequest{Action: "drop-table", ClientID: "client-1", APIKey: "key-1"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
	assert.False(t, queueHandler.registerCalled)
	assert.False(t, queueHandler.forwardCalled)
}
