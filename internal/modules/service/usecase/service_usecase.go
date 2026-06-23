package usecase

import (
	"context"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
)

type ServiceUseCase interface {
	CreateService(ctx context.Context, req *model.CreateServiceRequest) (*model.ServiceResponse, error)
	GetService(ctx context.Context, serviceID string) (*model.ServiceResponse, error)
	ListServices(ctx context.Context) ([]model.ServiceResponse, error)
	UpdateService(ctx context.Context, serviceID string, req *model.UpdateServiceRequest) (*model.ServiceResponse, error)
	DeleteService(ctx context.Context, serviceID string) error
}

type serviceUseCase struct {
	repo repository.ServiceRepository
}

func NewServiceUseCase(repo repository.ServiceRepository) ServiceUseCase {
	return &serviceUseCase{repo: repo}
}

func (u *serviceUseCase) CreateService(ctx context.Context, req *model.CreateServiceRequest) (*model.ServiceResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	now := time.Now().UnixMilli()
	service := &entity.Service{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		Code:      req.Code,
		Name:      req.Name,
		Status:    entity.ServiceStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := u.repo.Create(ctx, service); err != nil {
		return nil, err
	}
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
	if req.Status != nil {
		service.Status = *req.Status
	}
	service.UpdatedAt = time.Now().UnixMilli()
	if err := u.repo.Update(ctx, service); err != nil {
		return nil, err
	}
	return u.mapToResponse(service), nil
}

func (u *serviceUseCase) DeleteService(ctx context.Context, serviceID string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || serviceID == "" {
		return exception.ErrBadRequest
	}
	return u.repo.Delete(ctx, tenantID, serviceID)
}

func (u *serviceUseCase) mapToResponse(service *entity.Service) *model.ServiceResponse {
	return &model.ServiceResponse{
		ID:        service.ID,
		TenantID:  service.TenantID,
		Code:      service.Code,
		Name:      service.Name,
		Status:    service.Status,
		CreatedAt: service.CreatedAt,
		UpdatedAt: service.UpdatedAt,
	}
}
