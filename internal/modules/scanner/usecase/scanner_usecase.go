package usecase

import (
	"context"
	"strings"

	queueModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
)

const (
	ActionRegister = "register"
	ActionForward  = "forward"
)

type QueueHandler interface {
	RegisterQueue(ctx context.Context, req *queueModel.RegisterQueueRequest) (*queueModel.QueueResponse, error)
	ForwardQueue(ctx context.Context, queueID string, req *queueModel.ForwardQueueRequest) (*queueModel.QueueResponse, error)
}

type Authenticator interface {
	Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error
}

type CheckInRequest struct {
	Action               string
	ClientID             string
	APIKey               string
	ServiceID            string
	PatientID            string
	PatientName          string
	QueueID              string
	DestinationServiceID string
	DestinationCounterID string
}

type CheckInResponse struct {
	Action string
	Queue  *queueModel.QueueResponse
}

type ScannerUseCase interface {
	CheckIn(ctx context.Context, req *CheckInRequest) (*CheckInResponse, error)
}

type scannerUseCase struct {
	queueHandler  QueueHandler
	authenticator Authenticator
}

func NewScannerUseCase(queueHandler QueueHandler, authenticator Authenticator) ScannerUseCase {
	return &scannerUseCase{queueHandler: queueHandler, authenticator: authenticator}
}

func (u *scannerUseCase) CheckIn(ctx context.Context, req *CheckInRequest) (*CheckInResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" || req == nil {
		return nil, exception.ErrBadRequest
	}

	action := strings.TrimSpace(strings.ToLower(req.Action))
	if action != ActionRegister && action != ActionForward {
		return nil, exception.ErrBadRequest
	}

	if u.authenticator != nil {
		if err := u.authenticator.Authenticate(ctx, tenantID, branchID, req.ClientID, req.APIKey); err != nil {
			return nil, exception.ErrUnauthorized
		}
	}

	switch action {
	case ActionRegister:
		queueRes, err := u.queueHandler.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{
			ServiceID:   req.ServiceID,
			PatientID:   req.PatientID,
			PatientName: req.PatientName,
		})
		if err != nil {
			return nil, err
		}
		return &CheckInResponse{Action: ActionRegister, Queue: queueRes}, nil
	case ActionForward:
		queueRes, err := u.queueHandler.ForwardQueue(ctx, req.QueueID, &queueModel.ForwardQueueRequest{
			DestinationServiceID: req.DestinationServiceID,
			DestinationCounterID: req.DestinationCounterID,
		})
		if err != nil {
			return nil, err
		}
		return &CheckInResponse{Action: ActionForward, Queue: queueRes}, nil
	default:
		return nil, exception.ErrBadRequest
	}
}
