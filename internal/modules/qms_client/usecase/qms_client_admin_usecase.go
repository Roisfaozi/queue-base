package usecase

import (
	"context"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/model"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type AdminAuditLogger interface {
	LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error
}

type QMSClientAdminUseCase interface {
	CreateClient(ctx context.Context, req *model.QMSClientRequest) (*model.QMSClientResponse, error)
	CreateCredential(ctx context.Context, req *model.QMSClientCredentialRequest) (*model.QMSClientCredentialResponse, error)
	GetAll(ctx context.Context) ([]model.QMSClientResponse, error)
	GetByID(ctx context.Context, id string) (*model.QMSClientResponse, error)
	Update(ctx context.Context, id string, req *model.QMSClientUpdateRequest) (*model.QMSClientResponse, error)
	Delete(ctx context.Context, id string) error
}

type qmsClientAdminUseCase struct {
	log   *logrus.Logger
	db    *gorm.DB
	audit AdminAuditLogger
}

func NewQMSClientAdminUseCase(db *gorm.DB, log *logrus.Logger, audit ...AdminAuditLogger) QMSClientAdminUseCase {
	var auditLogger AdminAuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &qmsClientAdminUseCase{db: db, audit: auditLogger, log: log}
}

func (u *qmsClientAdminUseCase) CreateClient(ctx context.Context, req *model.QMSClientRequest) (*model.QMSClientResponse, error) {
	entry := u.logEntry(ctx, "CreateClient", nil)
	if entry != nil {
		entry.Info("start")
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || req == nil || req.BranchID == "" || req.ClientType == "" || req.Name == "" {
		return nil, exception.ErrBadRequest
	}
	cType := entity.ClientType(req.ClientType)
	if !cType.IsValid() {
		return nil, exception.ErrBadRequest
	}
	if err := u.ensureBranchInTenant(ctx, tenantID, req.BranchID); err != nil {
		return nil, err
	}
	now := time.Now().UnixMilli()
	client := &entity.QMSClient{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		BranchID:   req.BranchID,
		ClientType: cType,
		Name:       req.Name,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := u.db.WithContext(ctx).Create(client).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("create failed")
		}
		return nil, err
	}
	u.tryAudit(ctx, "QMS_CLIENT_CREATE", client.ID, map[string]string{"branch_id": client.BranchID, "client_type": string(client.ClientType), "name": client.Name})
	if entry != nil {
		entry.WithField("client_id", client.ID).Info("ok")
	}
	return toClientResponse(client), nil
}

func (u *qmsClientAdminUseCase) CreateCredential(ctx context.Context, req *model.QMSClientCredentialRequest) (*model.QMSClientCredentialResponse, error) {
	entry := u.logEntry(ctx, "CreateCredential", nil)
	if entry != nil {
		entry.Info("start")
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || req == nil || req.ClientID == "" || req.APIKey == "" {
		return nil, exception.ErrBadRequest
	}
	var client entity.QMSClient
	if err := u.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", req.ClientID, tenantID).First(&client).Error; err != nil {
		return nil, exception.ErrNotFound
	}
	hash, err := pkg.HashPassword(req.APIKey)
	if err != nil {
		if entry != nil {
			entry.WithError(err).Error("hash failed")
		}
		return nil, err
	}
	now := time.Now().UnixMilli()
	cred := &entity.QMSClientCredential{ID: uuid.New().String(), TenantID: tenantID, ClientID: req.ClientID, ClientSecretHash: hash, CreatedAt: now, UpdatedAt: now}
	if err := u.db.WithContext(ctx).Create(cred).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("create credential failed")
		}
		return nil, err
	}
	u.tryAudit(ctx, "QMS_CLIENT_CREDENTIAL_CREATE", cred.ID, map[string]string{"client_id": cred.ClientID})
	if entry != nil {
		entry.WithField("credential_id", cred.ID).Info("ok")
	}
	return &model.QMSClientCredentialResponse{ID: cred.ID, ClientID: cred.ClientID, ExpiresAt: cred.ExpiresAt, CreatedAt: cred.CreatedAt}, nil
}

func (u *qmsClientAdminUseCase) GetAll(ctx context.Context) ([]model.QMSClientResponse, error) {
	entry := u.logEntry(ctx, "GetAll", nil)
	if entry != nil {
		entry.Info("start")
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	var clients []entity.QMSClient
	if err := u.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&clients).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("query failed")
		}
		return nil, err
	}
	res := make([]model.QMSClientResponse, 0, len(clients))
	for i := range clients {
		res = append(res, *toClientResponse(&clients[i]))
	}
	if entry != nil {
		entry.WithField("count", len(res)).Info("ok")
	}
	return res, nil
}

func (u *qmsClientAdminUseCase) GetByID(ctx context.Context, id string) (*model.QMSClientResponse, error) {
	entry := u.logEntry(ctx, "GetByID", logrus.Fields{"client_id": id})
	if entry != nil {
		entry.Info("start")
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || id == "" {
		return nil, exception.ErrBadRequest
	}
	client, err := u.findClient(ctx, tenantID, id)
	if err != nil {
		if entry != nil {
			entry.Error("client not found")
		}
		return nil, err
	}
	if entry != nil {
		entry.Info("ok")
	}
	return toClientResponse(client), nil
}

func (u *qmsClientAdminUseCase) Update(ctx context.Context, id string, req *model.QMSClientUpdateRequest) (*model.QMSClientResponse, error) {
	entry := u.logEntry(ctx, "Update", logrus.Fields{"client_id": id})
	if entry != nil {
		entry.Info("start")
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || id == "" || req == nil {
		return nil, exception.ErrBadRequest
	}
	client, err := u.findClient(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != "" {
		client.Name = req.Name
	}
	if req.BranchServiceID != nil {
		if *req.BranchServiceID != "" {
			if err := u.ensureBranchServiceInTenantBranch(ctx, tenantID, client.BranchID, *req.BranchServiceID); err != nil {
				return nil, err
			}
		}
		client.BranchServiceID = req.BranchServiceID
	}
	if req.CounterID != nil {
		if *req.CounterID != "" {
			if err := u.ensureCounterInTenantBranch(ctx, tenantID, client.BranchID, *req.CounterID); err != nil {
				return nil, err
			}
		}
		client.CounterID = req.CounterID
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}
	client.UpdatedAt = time.Now().UnixMilli()
	if err := u.db.WithContext(ctx).Save(client).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("update failed")
		}
		return nil, err
	}
	u.tryAudit(ctx, "QMS_CLIENT_UPDATE", client.ID, map[string]string{"name": client.Name})
	if entry != nil {
		entry.WithField("is_active", client.IsActive).Info("ok")
	}
	return toClientResponse(client), nil
}

func (u *qmsClientAdminUseCase) Delete(ctx context.Context, id string) error {
	entry := u.logEntry(ctx, "Delete", logrus.Fields{"client_id": id})
	if entry != nil {
		entry.Info("start")
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || id == "" {
		return exception.ErrBadRequest
	}
	client, err := u.findClient(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if !client.IsActive {
		if entry != nil {
			entry.Info("already inactive")
		}
		return nil
	}
	client.IsActive = false
	client.UpdatedAt = time.Now().UnixMilli()
	if err := u.db.WithContext(ctx).Save(client).Error; err != nil {
		if entry != nil {
			entry.WithError(err).Error("deactivate failed")
		}
		return err
	}
	u.tryAudit(ctx, "QMS_CLIENT_DEACTIVATE", client.ID, map[string]string{"is_active": "false"})
	if entry != nil {
		entry.Info("ok")
	}
	return nil
}

func (u *qmsClientAdminUseCase) findClient(ctx context.Context, tenantID, id string) (*entity.QMSClient, error) {
	var client entity.QMSClient
	if err := u.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&client).Error; err != nil {
		return nil, exception.ErrNotFound
	}
	return &client, nil
}

func toClientResponse(client *entity.QMSClient) *model.QMSClientResponse {
	return &model.QMSClientResponse{ID: client.ID, TenantID: client.TenantID, BranchID: client.BranchID, ClientType: string(client.ClientType), Name: client.Name, BranchServiceID: client.BranchServiceID, CounterID: client.CounterID, IsActive: client.IsActive, CreatedAt: client.CreatedAt}
}

func (u *qmsClientAdminUseCase) logEntry(ctx context.Context, action string, fields logrus.Fields) *logrus.Entry {
	if u.log == nil {
		return nil
	}
	f := logrus.Fields{"module": "qms_client", "action": action}
	for k, v := range fields {
		f[k] = v
	}
	if tenantID := database.GetTenantID(ctx); tenantID != "" {
		f["tenant_id"] = tenantID
	}
	if userID, ok := authcontext.UserIDFromContext(ctx); ok && userID != "" {
		f["user_id"] = userID
	}
	return u.log.WithFields(f)
}

func (u *qmsClientAdminUseCase) tryAudit(ctx context.Context, action, entityID string, values map[string]string) {
	if u.audit == nil {
		return
	}
	userID, _ := authcontext.UserIDFromContext(ctx)
	_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{
		OrganizationID: database.GetTenantID(ctx),
		UserID:         userID,
		Action:         action,
		Entity:         "qms_client",
		EntityID:       entityID,
		NewValues:      values,
	})
}

func (u *qmsClientAdminUseCase) ensureBranchInTenant(ctx context.Context, tenantID, branchID string) error {
	var count int64
	if err := u.db.WithContext(ctx).Model(&branchEntity.Branch{}).Where("id = ? AND tenant_id = ?", branchID, tenantID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (u *qmsClientAdminUseCase) ensureBranchServiceInTenantBranch(ctx context.Context, tenantID, branchID, branchServiceID string) error {
	var count int64
	if err := u.db.WithContext(ctx).Model(&serviceEntity.BranchService{}).Where("id = ? AND tenant_id = ? AND branch_id = ?", branchServiceID, tenantID, branchID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return exception.ErrNotFound
	}
	return nil
}

func (u *qmsClientAdminUseCase) ensureCounterInTenantBranch(ctx context.Context, tenantID, branchID, counterID string) error {
	var count int64
	if err := u.db.WithContext(ctx).Model(&counterEntity.Counter{}).Where("id = ? AND tenant_id = ? AND branch_id = ?", counterID, tenantID, branchID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return exception.ErrNotFound
	}
	return nil
}
