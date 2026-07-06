package usecase

import (
	"context"
	"fmt"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/signage/model"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SignageUseCase interface {
	GetMe(ctx context.Context, clientID string) (*model.SignageMeResponse, error)
	GetCurrentCalls(ctx context.Context, clientID string) ([]model.SignageCurrentCallResponse, error)
	GetQueues(ctx context.Context, clientID string) ([]queueModel.QueueResponse, error)
}

type signageUseCase struct {
	db      *gorm.DB
	queueUC queueUsecase.QueueUseCase
	log     *logrus.Logger
}

func NewSignageUseCase(db *gorm.DB, qu queueUsecase.QueueUseCase, log *logrus.Logger) SignageUseCase {
	return &signageUseCase{db: db, queueUC: qu, log: log}
}

func (u *signageUseCase) logEntry(ctx context.Context, action string, fields logrus.Fields) *logrus.Entry {
	if u.log == nil {
		return nil
	}
	f := logrus.Fields{"module": "signage", "action": action}
	for k, v := range fields {
		f[k] = v
	}
	if tenantID := database.GetTenantID(ctx); tenantID != "" {
		f["tenant_id"] = tenantID
	}
	if branchID := database.GetBranchID(ctx); branchID != "" {
		f["branch_id"] = branchID
	}
	if userID, ok := authcontext.UserIDFromContext(ctx); ok && userID != "" {
		f["user_id"] = userID
	}
	return u.log.WithFields(f)
}

func (u *signageUseCase) GetMe(ctx context.Context, clientID string) (*model.SignageMeResponse, error) {
	entry := u.logEntry(ctx, "GetMe", logrus.Fields{"qms_client_id": clientID})
	if entry != nil {
		entry.Info("start")
	}

	if clientID == "" {
		return nil, exception.ErrUnauthorized
	}
	type clientRow struct {
		ID              string
		TenantID        string
		BranchID        string
		BranchServiceID *string
		CounterID       *string
		ClientType      string
		Name            string
	}
	var cr clientRow
	if err := u.db.WithContext(ctx).Table("qms_clients").
		Select("id, tenant_id, branch_id, branch_service_id, counter_id, client_type, name").
		Where("id = ? AND is_active = ?", clientID, true).
		First(&cr).Error; err != nil {
		if entry != nil {
			entry.Error("client not found")
		}
		return nil, exception.ErrNotFound
	}

	type branchRow struct {
		Name        string
		RunningText *string
		LogoAssetID *string
	}
	type tenantRow struct {
		RunningText *string
		LogoAssetID *string
	}
	var br branchRow
	_ = u.db.WithContext(ctx).Table("branches").
		Select("name, running_text, logo_asset_id").
		Where("id = ? AND tenant_id = ?", cr.BranchID, cr.TenantID).
		First(&br).Error
	var tenant tenantRow
	_ = u.db.WithContext(ctx).Table("organizations").
		Select("running_text, logo_asset_id").
		Where("id = ?", cr.TenantID).
		First(&tenant).Error

	res := &model.SignageMeResponse{
		ClientID:   cr.ID,
		TenantID:   cr.TenantID,
		BranchID:   cr.BranchID,
		ClientType: cr.ClientType,
		Name:       cr.Name,
		BranchName: br.Name,
	}
	if br.RunningText != nil {
		res.RunningText = *br.RunningText
	} else if tenant.RunningText != nil {
		res.RunningText = *tenant.RunningText
	}
	if br.LogoAssetID != nil {
		res.LogoAssetID = *br.LogoAssetID
	} else if tenant.LogoAssetID != nil {
		res.LogoAssetID = *tenant.LogoAssetID
	}

	if cr.BranchServiceID != nil && *cr.BranchServiceID != "" {
		res.BranchServiceID = *cr.BranchServiceID
		type svcRow struct {
			ServiceName            string `gorm:"column:service_name"`
			AudioID                *string
			AudioEN                *string
			NarrativeInstructionID *string
			NarrativeInstructionEN *string
		}
		var sr svcRow
		_ = u.db.WithContext(ctx).Table("branch_services").
			Joins("JOIN services ON services.id = branch_services.service_id").
			Select("services.name AS service_name, services.audio_id, services.audio_en, services.narrative_instruction_id, services.narrative_instruction_en").
			Where("branch_services.id = ? AND branch_services.tenant_id = ?", *cr.BranchServiceID, cr.TenantID).
			Take(&sr).Error
		res.ServiceName = sr.ServiceName
		if sr.AudioID != nil {
			res.AudioID = *sr.AudioID
		}
		if sr.AudioEN != nil {
			res.AudioEN = *sr.AudioEN
		}
		if sr.NarrativeInstructionID != nil {
			res.NarrativeInstructionID = *sr.NarrativeInstructionID
		}
		if sr.NarrativeInstructionEN != nil {
			res.NarrativeInstructionEN = *sr.NarrativeInstructionEN
		}
	}

	if cr.CounterID != nil && *cr.CounterID != "" {
		res.CounterID = *cr.CounterID
		type ctrRow struct {
			DisplayName *string
		}
		var ctr ctrRow
		_ = u.db.WithContext(ctx).Table("counters").
			Select("display_name").
			Where("id = ? AND tenant_id = ?", *cr.CounterID, cr.TenantID).
			Take(&ctr).Error
		if ctr.DisplayName != nil {
			res.CounterDisplayName = *ctr.DisplayName
		}
	}

	if entry != nil {
		entry.WithField("branch_id", cr.BranchID).Info("ok")
	}
	return res, nil
}

func (u *signageUseCase) GetCurrentCalls(ctx context.Context, clientID string) ([]model.SignageCurrentCallResponse, error) {
	entry := u.logEntry(ctx, "GetCurrentCalls", logrus.Fields{"qms_client_id": clientID})
	if entry != nil {
		entry.Info("start")
	}

	if clientID == "" {
		return nil, exception.ErrUnauthorized
	}
	type clientScope struct {
		TenantID        string
		BranchID        string
		BranchServiceID *string
		CounterID       *string
	}
	var cs clientScope
	if err := u.db.WithContext(ctx).Table("qms_clients").
		Select("tenant_id, branch_id, branch_service_id, counter_id").
		Where("id = ? AND is_active = ?", clientID, true).
		First(&cs).Error; err != nil {
		if entry != nil {
			entry.Error("client not found")
		}
		return nil, exception.ErrNotFound
	}

	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	if cs.TenantID != tenantID || cs.BranchID != branchID {
		if entry != nil {
			entry.Warn("tenant/branch mismatch")
		}
		return nil, exception.ErrForbidden
	}

	type callRow struct {
		QueueID                string
		TicketNo               string
		CounterID              string
		ServiceID              string
		CounterDisplayName     *string
		ServiceType            string
		AudioID                *string
		AudioEN                *string
		NarrativeInstructionID *string
		NarrativeInstructionEN *string
	}
	var rows []callRow
	query := u.db.WithContext(ctx).Table("queue_journeys AS qj").
		Select("qj.queue_id, q.ticket_no, COALESCE(qj.counter_id,'') AS counter_id, qj.service_id, c.display_name AS counter_display_name, COALESCE(s.type,'general') AS service_type, s.audio_id, s.audio_en, s.narrative_instruction_id, s.narrative_instruction_en").
		Joins("JOIN queues q ON q.id = qj.queue_id AND q.tenant_id = qj.tenant_id AND q.branch_id = qj.branch_id").
		Joins("LEFT JOIN counters c ON c.id = qj.counter_id AND c.tenant_id = qj.tenant_id").
		Joins("JOIN services s ON s.id = qj.service_id").
		Where("qj.tenant_id = ? AND qj.branch_id = ? AND qj.status = ?", tenantID, branchID, "calling").
		Order("qj.created_at DESC")

	if cs.BranchServiceID != nil && *cs.BranchServiceID != "" {
		type bsService struct{ ServiceID string }
		var bs bsService
		if err := u.db.WithContext(ctx).Table("branch_services").
			Select("service_id").
			Where("id = ? AND tenant_id = ?", *cs.BranchServiceID, tenantID).
			First(&bs).Error; err == nil && bs.ServiceID != "" {
			query = query.Where("qj.service_id = ?", bs.ServiceID)
		}
	}
	if cs.CounterID != nil && *cs.CounterID != "" {
		query = query.Where("qj.counter_id = ?", *cs.CounterID)
	}

	if err := query.Find(&rows).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("query failed")
		}
		return nil, fmt.Errorf("query current calls: %w", err)
	}

	if entry != nil {
		entry.WithField("count", len(rows)).Info("ok")
	}

	res := make([]model.SignageCurrentCallResponse, 0, len(rows))
	for _, r := range rows {
		item := model.SignageCurrentCallResponse{
			QueueID:     r.QueueID,
			TicketNo:    r.TicketNo,
			CounterID:   r.CounterID,
			ServiceID:   r.ServiceID,
			ServiceType: r.ServiceType,
		}
		if r.CounterDisplayName != nil {
			item.CounterDisplayName = *r.CounterDisplayName
		}
		if r.AudioID != nil {
			item.AudioID = *r.AudioID
		}
		if r.AudioEN != nil {
			item.AudioEN = *r.AudioEN
		}
		if r.NarrativeInstructionID != nil {
			item.NarrativeInstructionID = *r.NarrativeInstructionID
		}
		if r.NarrativeInstructionEN != nil {
			item.NarrativeInstructionEN = *r.NarrativeInstructionEN
		}
		res = append(res, item)
	}
	return res, nil
}

func (u *signageUseCase) GetQueues(ctx context.Context, clientID string) ([]queueModel.QueueResponse, error) {
	entry := u.logEntry(ctx, "GetQueues", logrus.Fields{"qms_client_id": clientID})
	if entry != nil {
		entry.Info("start")
	}

	if clientID == "" {
		return nil, exception.ErrUnauthorized
	}
	type clientScope struct {
		TenantID        string
		BranchID        string
		BranchServiceID *string
	}
	var cs clientScope
	if err := u.db.WithContext(ctx).Table("qms_clients").
		Select("tenant_id, branch_id, branch_service_id").
		Where("id = ? AND is_active = ?", clientID, true).
		First(&cs).Error; err != nil {
		if entry != nil {
			entry.Error("client not found")
		}
		return nil, exception.ErrNotFound
	}

	tenantID := database.GetTenantID(ctx)
	branchID := database.GetBranchID(ctx)
	if tenantID == "" || branchID == "" {
		return nil, exception.ErrBadRequest
	}
	if cs.TenantID != tenantID || cs.BranchID != branchID {
		if entry != nil {
			entry.Warn("tenant/branch mismatch")
		}
		return nil, exception.ErrForbidden
	}

	type queueRow struct {
		ID        string
		TenantID  string
		BranchID  string
		QueueDate string
		TicketNo  string
		QueueNo   int
		Status    string
		CreatedAt int64
		UpdatedAt int64
	}
	var rows []queueRow
	query := u.db.WithContext(ctx).Table("queue_journeys AS qj").
		Select("q.id, q.tenant_id, q.branch_id, q.queue_date, q.ticket_no, q.queue_no, q.status, q.created_at, q.updated_at").
		Joins("JOIN queues q ON q.id = qj.queue_id AND q.tenant_id = qj.tenant_id AND q.branch_id = qj.branch_id").
		Where("qj.tenant_id = ? AND qj.branch_id = ? AND qj.status = ?", tenantID, branchID, "pending").
		Order("qj.created_at ASC")

	if cs.BranchServiceID != nil && *cs.BranchServiceID != "" {
		type bsService struct{ ServiceID string }
		var bs bsService
		if err := u.db.WithContext(ctx).Table("branch_services").
			Select("service_id").
			Where("id = ? AND tenant_id = ?", *cs.BranchServiceID, tenantID).
			First(&bs).Error; err == nil && bs.ServiceID != "" {
			query = query.Where("qj.service_id = ?", bs.ServiceID)
		}
	}

	if err := query.Find(&rows).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("query failed")
		}
		return nil, fmt.Errorf("query waiting queues: %w", err)
	}

	if entry != nil {
		entry.WithField("count", len(rows)).Info("ok")
	}

	res := make([]queueModel.QueueResponse, 0, len(rows))
	for _, r := range rows {
		res = append(res, queueModel.QueueResponse{
			ID:        r.ID,
			TenantID:  r.TenantID,
			BranchID:  r.BranchID,
			QueueDate: r.QueueDate,
			TicketNo:  r.TicketNo,
			QueueNo:   r.QueueNo,
			Status:    r.Status,
			CreatedAt: r.CreatedAt,
			UpdatedAt: r.UpdatedAt,
		})
	}
	return res, nil
}
