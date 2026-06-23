package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
)

type stubServiceRepo struct {
	service *entity.Service
	list    []*entity.Service
	err     error
}

func (s *stubServiceRepo) Create(_ context.Context, service *entity.Service) error {
	s.service = service
	return s.err
}

func (s *stubServiceRepo) FindByID(_ context.Context, _, _ string) (*entity.Service, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.service, nil
}

func (s *stubServiceRepo) FindAll(_ context.Context, _ string) ([]*entity.Service, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *stubServiceRepo) Update(_ context.Context, service *entity.Service) error {
	s.service = service
	return s.err
}

func (s *stubServiceRepo) Delete(_ context.Context, _, _ string) error { return s.err }

func TestCreateServiceUsesTenantContext(t *testing.T) {
	repo := &stubServiceRepo{}
	uc := NewServiceUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	res, err := uc.CreateService(ctx, &model.CreateServiceRequest{Code: "reg", Name: "Registration"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.TenantID != "tenant-1" {
		t.Fatalf("expected tenant-1, got %s", res.TenantID)
	}
}

func TestGetServiceRequiresTenantAndID(t *testing.T) {
	uc := NewServiceUseCase(&stubServiceRepo{})
	_, err := uc.GetService(context.Background(), "")
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestCreateServiceIncludesPharmacyFlags(t *testing.T) {
	repo := &stubServiceRepo{}
	uc := NewServiceUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	res, err := uc.CreateService(ctx, &model.CreateServiceRequest{
		Code:                "pha",
		Name:                "Pharmacy",
		IsPharmacy:          true,
		IsPharmacyReception: true,
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !res.IsPharmacy {
		t.Fatalf("expected service response pharmacy flag true")
	}
	if !res.IsPharmacyReception {
		t.Fatalf("expected service response pharmacy reception flag true")
	}
	if repo.service == nil || !repo.service.IsPharmacy || !repo.service.IsPharmacyReception {
		t.Fatalf("expected persisted service pharmacy flags true")
	}
}

func TestUpdateServiceCanDisablePharmacyReception(t *testing.T) {
	repo := &stubServiceRepo{service: &entity.Service{
		ID:                  "svc-1",
		TenantID:            "tenant-1",
		Code:                "PHA",
		Name:                "Pharmacy",
		Status:              entity.ServiceStatusActive,
		IsPharmacy:          true,
		IsPharmacyReception: true,
	}}
	uc := NewServiceUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	flag := false

	res, err := uc.UpdateService(ctx, "svc-1", &model.UpdateServiceRequest{IsPharmacyReception: &flag})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.IsPharmacyReception {
		t.Fatalf("expected pharmacy reception false after update")
	}
	if repo.service == nil || repo.service.IsPharmacyReception {
		t.Fatalf("expected persisted pharmacy reception false after update")
	}
}

func TestCreateServiceRequiresTenantContextForPharmacyFlags(t *testing.T) {
	uc := NewServiceUseCase(&stubServiceRepo{})
	_, err := uc.CreateService(context.Background(), &model.CreateServiceRequest{
		Code:       "pha",
		Name:       "Pharmacy",
		IsPharmacy: true,
	})
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestListServicesRequiresTenant(t *testing.T) {
	uc := NewServiceUseCase(&stubServiceRepo{})
	_, err := uc.ListServices(context.Background())
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestDeleteServiceRequiresTenantAndID(t *testing.T) {
	uc := NewServiceUseCase(&stubServiceRepo{})
	err := uc.DeleteService(context.Background(), "svc-1")
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest no tenant, got %v", err)
	}
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	err = uc.DeleteService(ctx, "")
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest empty id, got %v", err)
	}
}

func TestUpdateServiceNotFoundReturnsError(t *testing.T) {
	repo := &stubServiceRepo{err: errors.New("not found")}
	uc := NewServiceUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	_, err := uc.UpdateService(ctx, "svc-1", &model.UpdateServiceRequest{})
	if !errors.Is(err, exception.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestCreateServiceSanitizesCodeAndName(t *testing.T) {
	repo := &stubServiceRepo{}
	uc := NewServiceUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	_, err := uc.CreateService(ctx, &model.CreateServiceRequest{Code: " reg ", Name: " Registration "})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if repo.service == nil || repo.service.Code != "REG" {
		t.Fatalf("expected sanitized code REG, got %v", repo.service)
	}
}

func TestUpdateServiceSanitizesCode(t *testing.T) {
	existing := &entity.Service{ID: "svc-1", TenantID: "tenant-1", Code: "OLD", Name: "Old", Status: entity.ServiceStatusActive}
	repo := &stubServiceRepo{service: existing}
	uc := NewServiceUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	code := " new "
	res, err := uc.UpdateService(ctx, "svc-1", &model.UpdateServiceRequest{Code: &code})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.Code != "NEW" {
		t.Fatalf("expected sanitized code NEW, got %s", res.Code)
	}
}
