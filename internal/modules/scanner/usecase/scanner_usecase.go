package usecase

import (
	"context"
	"fmt"
	"strings"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
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

type RelationValidator interface {
	Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error
}

type CheckInRequest struct {
	Action               string
	BranchID             string
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
	queueHandler      QueueHandler
	authenticator     Authenticator
	relationValidator RelationValidator
}

func NewScannerUseCase(queueHandler QueueHandler, authenticator Authenticator, relationValidator RelationValidator) ScannerUseCase {
	return &scannerUseCase{queueHandler: queueHandler, authenticator: authenticator, relationValidator: relationValidator}
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
			return nil, fmt.Errorf("authenticator failed (%v): %w", err, exception.ErrUnauthorized)
		}
	}

	serviceID := req.ServiceID
	if action == ActionForward {
		serviceID = req.DestinationServiceID
	}

	if u.relationValidator != nil {
		if err := u.relationValidator.Validate(ctx, tenantID, branchID, serviceID, req.DestinationCounterID); err != nil {
			return nil, fmt.Errorf("relation validator failed: %w", err)
		}
	}

	switch action {
	case ActionRegister:
		queueRes, err := u.queueHandler.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{
			BranchID:    req.BranchID,
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
