package usecase

import (
	"context"
	"fmt"
	"strings"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/telemetry"
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

type AuditLogger interface {
	LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error
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
	audit             AuditLogger
}

func NewScannerUseCase(queueHandler QueueHandler, authenticator Authenticator, relationValidator RelationValidator, audit ...AuditLogger) ScannerUseCase {
	var auditLogger AuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &scannerUseCase{queueHandler: queueHandler, authenticator: authenticator, relationValidator: relationValidator, audit: auditLogger}
}

func (u *scannerUseCase) CheckIn(ctx context.Context, req *CheckInRequest) (*CheckInResponse, error) {
	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" || req == nil {
		telemetry.ScannerCheckInsTotal.WithLabelValues("unknown", "bad_request").Inc()
		return nil, exception.ErrBadRequest
	}
	if req.BranchID == "" || req.BranchID != branchID {
		telemetry.ScannerCheckInsTotal.WithLabelValues("unknown", "forbidden").Inc()
		return nil, exception.ErrForbidden
	}

	action := strings.TrimSpace(strings.ToLower(req.Action))
	if action != ActionRegister && action != ActionForward {
		telemetry.ScannerCheckInsTotal.WithLabelValues("unknown", "bad_request").Inc()
		return nil, exception.ErrBadRequest
	}

	if u.authenticator != nil {
		if err := u.authenticator.Authenticate(ctx, tenantID, branchID, req.ClientID, req.APIKey); err != nil {
			telemetry.ScannerCheckInsTotal.WithLabelValues(action, "unauthorized").Inc()
			return nil, fmt.Errorf("authenticator failed (%v): %w", err, exception.ErrUnauthorized)
		}
	}

	serviceID := req.ServiceID
	if action == ActionForward {
		if req.QueueID == "" || req.DestinationServiceID == "" {
			telemetry.ScannerCheckInsTotal.WithLabelValues(action, "bad_request").Inc()
			return nil, exception.ErrBadRequest
		}
		serviceID = req.DestinationServiceID
	} else if serviceID == "" {
		telemetry.ScannerCheckInsTotal.WithLabelValues(action, "bad_request").Inc()
		return nil, exception.ErrBadRequest
	}

	if u.relationValidator != nil {
		if err := u.relationValidator.Validate(ctx, tenantID, branchID, serviceID, req.DestinationCounterID); err != nil {
			telemetry.ScannerCheckInsTotal.WithLabelValues(action, "forbidden").Inc()
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
			telemetry.ScannerCheckInsTotal.WithLabelValues(ActionRegister, "failed").Inc()
			return nil, err
		}
		u.tryAudit(ctx, "SCANNER_REGISTER", queueRes.ID, req.BranchID)
		telemetry.ScannerCheckInsTotal.WithLabelValues(ActionRegister, "success").Inc()
		return &CheckInResponse{Action: ActionRegister, Queue: queueRes}, nil
	case ActionForward:
		queueRes, err := u.queueHandler.ForwardQueue(ctx, req.QueueID, &queueModel.ForwardQueueRequest{
			DestinationServiceID: req.DestinationServiceID,
			DestinationCounterID: req.DestinationCounterID,
		})
		if err != nil {
			telemetry.ScannerCheckInsTotal.WithLabelValues(ActionForward, "failed").Inc()
			return nil, err
		}
		u.tryAudit(ctx, "SCANNER_FORWARD", queueRes.ID, req.BranchID)
		telemetry.ScannerCheckInsTotal.WithLabelValues(ActionForward, "success").Inc()
		return &CheckInResponse{Action: ActionForward, Queue: queueRes}, nil
	default:
		telemetry.ScannerCheckInsTotal.WithLabelValues(action, "bad_request").Inc()
		return nil, exception.ErrBadRequest
	}
}

func (u *scannerUseCase) tryAudit(ctx context.Context, action, entityID, branchID string) {
	if u.audit == nil {
		return
	}
	userID, ok := authcontext.UserIDFromContext(ctx)
	if !ok || userID == "" {
		userID = "system"
	}
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{
		OrganizationID: database.GetTenantID(ctx),
		UserID:         userID,
		Action:         action,
		Entity:         "scanner",
		EntityID:       entityID,
		NewValues:      map[string]string{"branch_id": branchID},
	})
}
