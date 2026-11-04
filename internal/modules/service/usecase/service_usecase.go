package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/storage"
	"github.com/google/uuid"
)

type AuditLogger interface {
	LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error
}

type ServiceUseCase interface {
	CreateService(ctx context.Context, req *model.CreateServiceRequest) (*model.ServiceResponse, error)
	GetService(ctx context.Context, serviceID string) (*model.ServiceResponse, error)
	ListServices(ctx context.Context) ([]model.ServiceResponse, error)
	UpdateService(ctx context.Context, serviceID string, req *model.UpdateServiceRequest) (*model.ServiceResponse, error)
	DeleteService(ctx context.Context, serviceID string) error
	UploadAudio(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
}

type serviceUseCase struct {
	repo    repository.ServiceRepository
	storage storage.Provider
	audit   AuditLogger
}

func NewServiceUseCase(repo repository.ServiceRepository, storage storage.Provider, audit ...AuditLogger) ServiceUseCase {
	var auditLogger AuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &serviceUseCase{repo: repo, storage: storage, audit: auditLogger}
}

func (u *serviceUseCase) CreateService(ctx context.Context, req *model.CreateServiceRequest) (*model.ServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	serviceType := req.Type
	if serviceType == "" {
		serviceType = entity.ServiceTypeGeneral
	}
	estimatedDuration := req.DefaultEstimatedDuration
	if estimatedDuration == 0 {
		estimatedDuration = 5
	}
	now := time.Now().UnixMilli()
	service := &entity.Service{
		ID:                       uuid.New().String(),
		TenantID:                 tenantID,
		Code:                     req.Code,
		Name:                     req.Name,
		Type:                     serviceType,
		DefaultEstimatedDuration: estimatedDuration,
		AudioID:                  req.AudioID,
		AudioEN:                  req.AudioEN,
		NarrativeInstructionID:   req.NarrativeInstructionID,
		NarrativeInstructionEN:   req.NarrativeInstructionEN,
		Status:                   entity.ServiceStatusActive,
		IsPharmacy:               req.IsPharmacy,
		IsPharmacyReception:      req.IsPharmacyReception,
		CreatedAt:                now,
		UpdatedAt:                now,
	}
	if err := u.repo.Create(ctx, service); err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "SERVICE_CREATE", service.ID, map[string]any{"code": service.Code, "name": service.Name, "type": service.Type, "status": service.Status})
	return u.mapToResponse(service), nil
}

func (u *serviceUseCase) GetService(ctx context.Context, serviceID string) (*model.ServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || serviceID == "" {
		return nil, exception.ErrBadRequest
	}
	service, err := u.repo.FindByID(ctx, tenantID, serviceID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	return u.mapToResponse(service), nil
}

func (u *serviceUseCase) ListServices(ctx context.Context) ([]model.ServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	services, err := u.repo.FindAll(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	res := make([]model.ServiceResponse, len(services))
	for i, service := range services {
		res[i] = *u.mapToResponse(service)
	}
	return res, nil
}

func (u *serviceUseCase) UpdateService(ctx context.Context, serviceID string, req *model.UpdateServiceRequest) (*model.ServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || serviceID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	service, err := u.repo.FindByID(ctx, tenantID, serviceID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	if req.Code != nil {
		service.Code = *req.Code
	}
	if req.Name != nil {
		service.Name = *req.Name
	}
	if req.Type != nil {
		service.Type = *req.Type
	}
	if req.DefaultEstimatedDuration != nil {
		service.DefaultEstimatedDuration = *req.DefaultEstimatedDuration
	}
	if req.AudioID != nil {
		service.AudioID = *req.AudioID
	}
	if req.AudioEN != nil {
		service.AudioEN = *req.AudioEN
	}
	if req.NarrativeInstructionID != nil {
		service.NarrativeInstructionID = *req.NarrativeInstructionID
	}
	if req.NarrativeInstructionEN != nil {
		service.NarrativeInstructionEN = *req.NarrativeInstructionEN
	}
	if req.Status != nil {
		service.Status = *req.Status
	}
	if req.IsPharmacy != nil {
		service.IsPharmacy = *req.IsPharmacy
	}
	if req.IsPharmacyReception != nil {
		service.IsPharmacyReception = *req.IsPharmacyReception
	}
	service.UpdatedAt = time.Now().UnixMilli()
	if err := u.repo.Update(ctx, service); err != nil {
		return nil, err
	}
	u.tryAudit(ctx, "SERVICE_UPDATE", service.ID, map[string]any{"code": service.Code, "name": service.Name, "type": service.Type, "status": service.Status})
	return u.mapToResponse(service), nil
}

func (u *serviceUseCase) DeleteService(ctx context.Context, serviceID string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || serviceID == "" {
		return exception.ErrBadRequest
	}
	if err := u.repo.Delete(ctx, tenantID, serviceID); err != nil {
		return err
	}
	u.tryAudit(ctx, "SERVICE_DELETE", serviceID, nil)
	return nil
}

func (u *serviceUseCase) tryAudit(ctx context.Context, action, entityID string, values map[string]any) {
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
		Entity:         "service",
		EntityID:       entityID,
		NewValues:      values,
	})
}

func (u *serviceUseCase) mapToResponse(service *entity.Service) *model.ServiceResponse {
	return &model.ServiceResponse{
		ID:                       service.ID,
		TenantID:                 service.TenantID,
		Code:                     service.Code,
		Name:                     service.Name,
		Type:                     service.Type,
		DefaultEstimatedDuration: service.DefaultEstimatedDuration,
		AudioID:                  service.AudioID,
		AudioEN:                  service.AudioEN,
		NarrativeInstructionID:   service.NarrativeInstructionID,
		NarrativeInstructionEN:   service.NarrativeInstructionEN,
		Status:                   service.Status,
		IsPharmacy:               service.IsPharmacy,
		IsPharmacyReception:      service.IsPharmacyReception,
		CreatedAt:                service.CreatedAt,
		UpdatedAt:                service.UpdatedAt,
	}
}

func (u *serviceUseCase) UploadAudio(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return "", exception.ErrBadRequest
	}
	if u.storage == nil {
		return "", exception.ErrInternalServer
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return "", exception.ErrBadRequest
	}
	buf = buf[:n]

	detectedType := http.DetectContentType(buf)
	if strings.HasPrefix(detectedType, "text/plain") || detectedType == "application/octet-stream" {
		if contentType != "" {
			detectedType = contentType
		}
	}

	allowedTypes := map[string]string{
		"audio/mpeg":  ".mp3",
		"audio/mp3":   ".mp3",
		"audio/wav":   ".wav",
		"audio/x-wav": ".wav",
		"audio/wave":  ".wav",
		"audio/ogg":   ".ogg",
		"audio/x-m4a": ".m4a",
		"audio/m4a":   ".m4a",
		"audio/aac":   ".aac",
		"audio/mp4":   ".mp4",
		"video/ogg":   ".ogg",
	}

	ext, allowed := allowedTypes[detectedType]
	if !allowed {
		lowerName := strings.ToLower(filename)
		switch {
		case strings.HasSuffix(lowerName, ".mp3"):
			ext = ".mp3"
			detectedType = "audio/mpeg"
		case strings.HasSuffix(lowerName, ".wav"):
			ext = ".wav"
			detectedType = "audio/wav"
		case strings.HasSuffix(lowerName, ".ogg"):
			ext = ".ogg"
			detectedType = "audio/ogg"
		case strings.HasSuffix(lowerName, ".m4a"):
			ext = ".m4a"
			detectedType = "audio/x-m4a"
		case strings.HasSuffix(lowerName, ".aac"):
			ext = ".aac"
			detectedType = "audio/aac"
		default:
			return "", exception.ErrValidationError
		}
	}

	file = io.MultiReader(bytes.NewReader(buf), file)
	newFilename := fmt.Sprintf("services/audio/%s_%s%s", tenantID, uuid.New().String(), ext)

	url, err := u.storage.UploadFile(ctx, file, newFilename, detectedType)
	if err != nil {
		return "", exception.ErrInternalServer
	}

	if u.audit != nil {
		userID, _ := authcontext.UserIDFromContext(ctx)
		_ = u.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{
			UserID:   userID,
			Action:   "UPLOAD_SERVICE_AUDIO",
			Entity:   "Service",
			EntityID: tenantID,
			NewValues: map[string]interface{}{
				"url": url,
			},
		})
	}

	return url, nil
}
