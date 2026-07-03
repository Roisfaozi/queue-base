package usecase

import (
	"context"
	"strings"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	auditUsecase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	authUsecase "github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/caller/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"gorm.io/gorm"
)

type CallerUseCase interface {
	Login(ctx context.Context, clientID string, req model.CallerLoginRequest) (*model.CallerLoginResponse, string, error)
	Me(ctx context.Context, clientID string) (*model.CallerMeResponse, error)
	ExecuteAction(ctx context.Context, clientID, journeyID, action string) (*model.CallerActionResponse, error)
}

type callerUseCase struct {
	db      *gorm.DB
	queueUC queueUsecase.QueueUseCase
	authUC  authUsecase.AuthUseCase
	audit   auditUsecase.AuditUseCase
}

func NewCallerUseCase(db *gorm.DB, qu queueUsecase.QueueUseCase, authUC authUsecase.AuthUseCase, auditUC auditUsecase.AuditUseCase) CallerUseCase {
	return &callerUseCase{db: db, queueUC: qu, authUC: authUC, audit: auditUC}
}

func (u *callerUseCase) Login(ctx context.Context, clientID string, req model.CallerLoginRequest) (*model.CallerLoginResponse, string, error) {
	if clientID == "" {
		return nil, "", exception.ErrUnauthorized
	}
	if u.authUC == nil {
		return nil, "", exception.ErrInternalServer
	}

	binding, err := u.resolveCallerBinding(ctx, clientID)
	if err != nil {
		return nil, "", err
	}

	loginRes, refreshToken, err := u.authUC.Login(ctx, authModel.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, "", err
	}

	if err := u.validateOperatorAccess(ctx, binding, loginRes.User.ID); err != nil {
		return nil, "", err
	}

	return mapCallerLoginResponse(loginRes.AccessToken, binding), refreshToken, nil
}

func (u *callerUseCase) Me(ctx context.Context, clientID string) (*model.CallerMeResponse, error) {
	if clientID == "" {
		return nil, exception.ErrUnauthorized
	}
	userID, ok := authcontext.UserIDFromContext(ctx)
	if !ok || userID == "" {
		return nil, exception.ErrUnauthorized
	}

	binding, err := u.resolveCallerBinding(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if err := u.validateOperatorAccess(ctx, binding, userID); err != nil {
		return nil, err
	}

	res := &model.CallerMeResponse{}
	res.CallerLoginResponse = *mapCallerLoginResponse("", binding)
	return res, nil
}

func (u *callerUseCase) ExecuteAction(ctx context.Context, clientID, journeyID, action string) (*model.CallerActionResponse, error) {
	if clientID == "" {
		return nil, exception.ErrUnauthorized
	}
	type journeyRow struct {
		QueueID   string
		TenantID  string
		BranchID  string
		CounterID *string
	}
	var jr journeyRow
	if err := u.db.WithContext(ctx).Table("queue_journeys").
		Select("queue_id, tenant_id, branch_id, counter_id").
		Where("id = ?", journeyID).
		First(&jr).Error; err != nil {
		return nil, exception.ErrNotFound
	}

	binding, err := u.resolveCallerBinding(ctx, clientID)
	if err != nil {
		return nil, err
	}

	if binding.TenantID != jr.TenantID || binding.BranchID != jr.BranchID {
		return nil, exception.ErrForbidden
	}
	if binding.CounterID != "" {
		if jr.CounterID == nil || *jr.CounterID == "" || binding.CounterID != *jr.CounterID {
			return nil, exception.ErrForbidden
		}
	}

	if userID, ok := authcontext.UserIDFromContext(ctx); ok && userID != "" && binding.CounterID != "" {
		if err := u.validateOperatorAccess(ctx, binding, userID); err != nil {
			return nil, err
		}
	}

	ctx = database.SetOrganizationContext(ctx, jr.TenantID)
	ctx = database.SetBranchContext(ctx, jr.BranchID)

	transitionReq := &queueModel.QueueTransitionRequest{Action: action}
	res, err := u.queueUC.TransitionQueue(ctx, jr.QueueID, transitionReq)
	if err != nil {
		return nil, err
	}
	u.tryAudit(ctx, userIDOrSystem(ctx), "CALLER_"+strings.ToUpper(action), jr.QueueID, map[string]string{
		"branch_id":  jr.BranchID,
		"journey_id": journeyID,
		"client_id":  clientID,
		"status":     res.Status,
	})
	return &model.CallerActionResponse{
		Success:   true,
		TrackNo:   res.TicketNo,
		QueueNo:   res.QueueNo,
		Status:    res.Status,
		JourneyID: journeyID,
	}, nil
}

type callerBinding struct {
	TenantID           string
	TenantName         string
	BranchID           string
	BranchName         string
	BranchServiceID    string
	ServiceName        string
	CounterID          string
	CounterName        string
	CounterDisplayName string
}

func (u *callerUseCase) resolveCallerBinding(ctx context.Context, clientID string) (*callerBinding, error) {
	type bindingRow struct {
		TenantID        string
		TenantName      string
		BranchID        string
		BranchName      string
		BranchServiceID *string
		ServiceName     *string
		CounterID       *string
		CounterName     *string
		DisplayName     *string
	}
	var row bindingRow
	if err := u.db.WithContext(ctx).Table("qms_clients qc").
		Select(strings.Join([]string{
			"qc.tenant_id",
			"org.name AS tenant_name",
			"qc.branch_id",
			"branches.name AS branch_name",
			"qc.branch_service_id",
			"services.name AS service_name",
			"qc.counter_id",
			"counters.name AS counter_name",
			"counters.display_name AS display_name",
		}, ", ")).
		Joins("JOIN organizations org ON org.id = qc.tenant_id").
		Joins("JOIN branches ON branches.id = qc.branch_id AND branches.tenant_id = qc.tenant_id").
		Joins("LEFT JOIN branch_services ON branch_services.id = qc.branch_service_id AND branch_services.tenant_id = qc.tenant_id").
		Joins("LEFT JOIN services ON services.id = branch_services.service_id").
		Joins("LEFT JOIN counters ON counters.id = qc.counter_id AND counters.tenant_id = qc.tenant_id").
		Where("qc.id = ? AND qc.is_active = ? AND qc.client_type = ?", clientID, true, "caller").
		First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, exception.ErrUnauthorized
		}
		return nil, err
	}
	return &callerBinding{
		TenantID:           row.TenantID,
		TenantName:         row.TenantName,
		BranchID:           row.BranchID,
		BranchName:         row.BranchName,
		BranchServiceID:    derefString(row.BranchServiceID),
		ServiceName:        derefString(row.ServiceName),
		CounterID:          derefString(row.CounterID),
		CounterName:        derefString(row.CounterName),
		CounterDisplayName: derefString(row.DisplayName),
	}, nil
}

func (u *callerUseCase) validateOperatorAccess(ctx context.Context, binding *callerBinding, userID string) error {
	var userCount int64
	if err := u.db.WithContext(ctx).Table("organization_members").
		Where("organization_id = ? AND user_id = ? AND status = ? AND (deleted_at = 0 OR deleted_at IS NULL)", binding.TenantID, userID, "active").
		Count(&userCount).Error; err != nil {
		return err
	}
	if userCount == 0 {
		return exception.ErrForbidden
	}
	if binding.CounterID == "" {
		return nil
	}

	var assignmentCount int64
	if err := u.db.WithContext(ctx).Table("operator_counter_assignments").
		Where("tenant_id = ? AND branch_id = ? AND user_id = ? AND counter_id = ? AND unassigned_at IS NULL", binding.TenantID, binding.BranchID, userID, binding.CounterID).
		Count(&assignmentCount).Error; err != nil {
		return err
	}
	if assignmentCount == 0 {
		return exception.ErrForbidden
	}
	return nil
}

func (u *callerUseCase) tryAudit(ctx context.Context, userID, action, entityID string, values map[string]string) {
	if u.audit == nil {
		return
	}
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{
		OrganizationID: database.GetTenantID(ctx),
		UserID:         userID,
		Action:         action,
		Entity:         "queue",
		EntityID:       entityID,
		NewValues:      values,
	})
}

func mapCallerLoginResponse(accessToken string, binding *callerBinding) *model.CallerLoginResponse {
	return &model.CallerLoginResponse{
		AccessToken: accessToken,
		Permissions: []string{"queue.read", "queue.call", "queue.start", "queue.forward", "queue.complete", "queue.skip", "queue.cancel"},
		Context: struct {
			TenantID           string `json:"tenant_id"`
			TenantName         string `json:"tenant_name,omitempty"`
			BranchID           string `json:"branch_id"`
			BranchName         string `json:"branch_name,omitempty"`
			BranchServiceID    string `json:"branch_service_id,omitempty"`
			ServiceName        string `json:"service_name,omitempty"`
			CounterID          string `json:"counter_id,omitempty"`
			CounterName        string `json:"counter_name,omitempty"`
			CounterDisplayName string `json:"display_name,omitempty"`
		}{
			TenantID:           binding.TenantID,
			TenantName:         binding.TenantName,
			BranchID:           binding.BranchID,
			BranchName:         binding.BranchName,
			BranchServiceID:    binding.BranchServiceID,
			ServiceName:        binding.ServiceName,
			CounterID:          binding.CounterID,
			CounterName:        binding.CounterName,
			CounterDisplayName: binding.CounterDisplayName,
		},
	}
}

func userIDOrSystem(ctx context.Context) string {
	userID, ok := authcontext.UserIDFromContext(ctx)
	if !ok || userID == "" {
		return "system"
	}
	return userID
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
