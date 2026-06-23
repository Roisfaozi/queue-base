package usecase

import (
	"context"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/google/uuid"
)

type CounterUseCase interface {
	CreateCounter(ctx context.Context, req *model.CreateCounterRequest) (*model.CounterResponse, error)
	GetCounter(ctx context.Context, counterID string) (*model.CounterResponse, error)
	ListCounters(ctx context.Context) ([]model.CounterResponse, error)
	UpdateCounter(ctx context.Context, counterID string, req *model.UpdateCounterRequest) (*model.CounterResponse, error)
	DeleteCounter(ctx context.Context, counterID string) error
}

type counterUseCase struct {
	repo repository.CounterRepository
}

func NewCounterUseCase(repo repository.CounterRepository) CounterUseCase {
	return &counterUseCase{repo: repo}
}

func (u *counterUseCase) CreateCounter(ctx context.Context, req *model.CreateCounterRequest) (*model.CounterResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	now := time.Now().UnixMilli()
	counter := &entity.Counter{
		ID:        uuid.New().String(),
		TenantID:  tenantID,
		BranchID:  req.BranchID,
		Code:      req.Code,
		Name:      req.Name,
		Status:    entity.CounterStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := u.repo.Create(ctx, counter); err != nil {
		return nil, err
	}
	return u.mapToResponse(counter), nil
}

func (u *counterUseCase) GetCounter(ctx context.Context, counterID string) (*model.CounterResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || counterID == "" {
		return nil, exception.ErrBadRequest
	}
	counter, err := u.repo.FindByID(ctx, tenantID, counterID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	return u.mapToResponse(counter), nil
}

func (u *counterUseCase) ListCounters(ctx context.Context) ([]model.CounterResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return nil, exception.ErrBadRequest
	}
	counters, err := u.repo.FindAll(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	res := make([]model.CounterResponse, len(counters))
	for i, counter := range counters {
		res[i] = *u.mapToResponse(counter)
	}
	return res, nil
}

func (u *counterUseCase) UpdateCounter(ctx context.Context, counterID string, req *model.UpdateCounterRequest) (*model.CounterResponse, error) {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || counterID == "" {
		return nil, exception.ErrBadRequest
	}
	req.Sanitize()
	counter, err := u.repo.FindByID(ctx, tenantID, counterID)
	if err != nil {
		return nil, exception.ErrNotFound
	}
	if req.Code != nil {
		counter.Code = *req.Code
	}
	if req.Name != nil {
		counter.Name = *req.Name
	}
	if req.Status != nil {
		counter.Status = *req.Status
	}
	counter.UpdatedAt = time.Now().UnixMilli()
	if err := u.repo.Update(ctx, counter); err != nil {
		return nil, err
	}
	return u.mapToResponse(counter), nil
}

func (u *counterUseCase) DeleteCounter(ctx context.Context, counterID string) error {
	tenantID := database.GetTenantID(ctx)
	if tenantID == "" || counterID == "" {
		return exception.ErrBadRequest
	}
	return u.repo.Delete(ctx, tenantID, counterID)
}

func (u *counterUseCase) mapToResponse(counter *entity.Counter) *model.CounterResponse {
	return &model.CounterResponse{
		ID:        counter.ID,
		TenantID:  counter.TenantID,
		BranchID:  counter.BranchID,
		Code:      counter.Code,
		Name:      counter.Name,
		Status:    counter.Status,
		CreatedAt: counter.CreatedAt,
		UpdatedAt: counter.UpdatedAt,
	}
}
