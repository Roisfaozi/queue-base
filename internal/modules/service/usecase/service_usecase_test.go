package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
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
