package usecase

import (
	"context"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/model"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
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
	db    *gorm.DB
	audit AdminAuditLogger
}

func NewQMSClientAdminUseCase(db *gorm.DB, audit ...AdminAuditLogger) QMSClientAdminUseCase {
	var auditLogger AdminAuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &qmsClientAdminUseCase{db: db, audit: auditLogger}
}

func (u *qmsClientAdminUseCase) CreateClient(ctx context.Context, req *model.QMSClientRequest) (*model.QMSClientResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || req == nil || req.BranchID == "" || req.ClientType == "" || req.Name == "" {
		return nil, exception.ErrBadRequest
	}
	now := time.Now().UnixMilli()
	client := &entity.QMSClient{
		ID:         uuid.New().String(),
		TenantID:   tenantID,
		BranchID:   req.BranchID,
		ClientType: entity.ClientType(req.ClientType),
		Name:       req.Name,
		IsActive:   true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := u.db.WithContext(ctx).Create(client).Error; err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "QMS_CLIENT_CREATE", client.ID, map[string]string{"branch_id": client.BranchID, "client_type": string(client.ClientType), "name": client.Name})
	return toClientResponse(client), nil
}

func (u *qmsClientAdminUseCase) CreateCredential(ctx context.Context, req *model.QMSClientCredentialRequest) (*model.QMSClientCredentialResponse, error) {
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
		return nil, err
	}
	now := time.Now().UnixMilli()
	cred := &entity.QMSClientCredential{ID: uuid.New().String(), TenantID: tenantID, ClientID: req.ClientID, ClientSecretHash: hash, CreatedAt: now, UpdatedAt: now}
	if err := u.db.WithContext(ctx).Create(cred).Error; err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "QMS_CLIENT_CREDENTIAL_CREATE", cred.ID, map[string]string{"client_id": cred.ClientID})
	return &model.QMSClientCredentialResponse{ID: cred.ID, ClientID: cred.ClientID, ExpiresAt: cred.ExpiresAt, CreatedAt: cred.CreatedAt}, nil
}

func (u *qmsClientAdminUseCase) GetAll(ctx context.Context) ([]model.QMSClientResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	var clients []entity.QMSClient
	if err := u.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&clients).Error; err != nil {
		return nil, err
	}
	res := make([]model.QMSClientResponse, 0, len(clients))
	for i := range clients {
		res = append(res, *toClientResponse(&clients[i]))
	}
	return res, nil
}

func (u *qmsClientAdminUseCase) GetByID(ctx context.Context, id string) (*model.QMSClientResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || id == "" {
		return nil, exception.ErrBadRequest
	}
	client, err := u.findClient(ctx, tenantID, id)
	if err != nil {
		return nil, err
	}
	return toClientResponse(client), nil
}

func (u *qmsClientAdminUseCase) Update(ctx context.Context, id string, req *model.QMSClientUpdateRequest) (*model.QMSClientResponse, error) {
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
		client.BranchServiceID = req.BranchServiceID
	}
	if req.CounterID != nil {
		client.CounterID = req.CounterID
	}
	if req.IsActive != nil {
		client.IsActive = *req.IsActive
	}
	client.UpdatedAt = time.Now().UnixMilli()
	if err := u.db.WithContext(ctx).Save(client).Error; err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "QMS_CLIENT_UPDATE", client.ID, map[string]string{"name": client.Name})
	return toClientResponse(client), nil
}

func (u *qmsClientAdminUseCase) Delete(ctx context.Context, id string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || id == "" {
		return exception.ErrBadRequest
	}
	client, err := u.findClient(ctx, tenantID, id)
	if err != nil {
		return err
	}
	if !client.IsActive {
		return nil
	}
	client.IsActive = false
	client.UpdatedAt = time.Now().UnixMilli()
	if err := u.db.WithContext(ctx).Save(client).Error; err != nil {
		return err
	}
	u.tryAudit(ctx, "QMS_CLIENT_DEACTIVATE", client.ID, map[string]string{"is_active": "false"})
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
