package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	organizationEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
)

type stubCounterRepo struct {
	counter *entity.Counter
	list    []*entity.Counter
	err     error
	seen    struct {
		tenantID  string
		counterID string
	}
}

func (s *stubCounterRepo) Create(_ context.Context, counter *entity.Counter) error {
	s.counter = counter
	return s.err
}

func (s *stubCounterRepo) FindByID(_ context.Context, tenantID, counterID string) (*entity.Counter, error) {
	s.seen.tenantID = tenantID
	s.seen.counterID = counterID
	if s.err != nil {
		return nil, s.err
	}
	return s.counter, nil
}

func (s *stubCounterRepo) FindAll(_ context.Context, tenantID string) ([]*entity.Counter, error) {
	s.seen.tenantID = tenantID
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *stubCounterRepo) Update(_ context.Context, counter *entity.Counter) error {
	s.counter = counter
	return s.err
}

func (s *stubCounterRepo) Delete(_ context.Context, tenantID, counterID string) error {
	s.seen.tenantID = tenantID
	s.seen.counterID = counterID
	return s.err
}

type stubCounterBranchRepo struct {
	branch *organizationEntity.Branch
	err    error
	seen   struct {
		tenantID string
		branchID string
	}
}

func (s *stubCounterBranchRepo) Create(_ context.Context, branch *organizationEntity.Branch) error {
	s.branch = branch
	return s.err
}

func (s *stubCounterBranchRepo) FindByID(_ context.Context, tenantID, branchID string) (*organizationEntity.Branch, error) {
	s.seen.tenantID = tenantID
	s.seen.branchID = branchID
	if s.err != nil {
		return nil, s.err
	}
	return s.branch, nil
}

func (s *stubCounterBranchRepo) FindAll(_ context.Context, tenantID string) ([]*organizationEntity.Branch, error) {
	s.seen.tenantID = tenantID
	if s.err != nil {
		return nil, s.err
	}
	return nil, nil
}

func (s *stubCounterBranchRepo) Update(_ context.Context, branch *organizationEntity.Branch) error {
	s.branch = branch
	return s.err
}

func (s *stubCounterBranchRepo) Delete(_ context.Context, tenantID, branchID string) error {
	s.seen.tenantID = tenantID
	s.seen.branchID = branchID
	return s.err
}

func TestCreateCounterUsesTenantScopedBranch(t *testing.T) {
	repo := &stubCounterRepo{}
	branchRepo := &stubCounterBranchRepo{branch: &organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1"}}
	uc := NewCounterUseCase(repo, branchRepo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	res, err := uc.CreateCounter(ctx, &model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: " a1 ", Name: " Front Desk "})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.TenantID != "tenant-1" {
		t.Fatalf("expected tenant-1, got %s", res.TenantID)
	}
	if branchRepo.seen.tenantID != "tenant-1" || branchRepo.seen.branchID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Fatalf("unexpected branch lookup: %+v", branchRepo.seen)
	}
	if repo.counter == nil || repo.counter.Code != "A1" || repo.counter.Name != "Front Desk" {
		t.Fatalf("counter not sanitized/stored: %+v", repo.counter)
	}
}

func TestCreateCounterRequiresTenant(t *testing.T) {
	uc := NewCounterUseCase(&stubCounterRepo{}, &stubCounterBranchRepo{})

	_, err := uc.CreateCounter(context.Background(), &model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"})
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestUpdateCounterSanitizesFields(t *testing.T) {
	code := " b2 "
	name := " Front Office "
	repo := &stubCounterRepo{counter: &entity.Counter{ID: "counter-1", TenantID: "tenant-1", BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk", Status: entity.CounterStatusActive}}
	uc := NewCounterUseCase(repo, &stubCounterBranchRepo{})
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	res, err := uc.UpdateCounter(ctx, "counter-1", &model.UpdateCounterRequest{Code: &code, Name: &name})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.Code != "B2" || res.Name != "Front Office" {
		t.Fatalf("expected sanitized response, got %+v", res)
	}
}

func TestCreateCounterRejectsCrossTenantBranch(t *testing.T) {
	repo := &stubCounterRepo{}
	branchRepo := &stubCounterBranchRepo{err: exception.ErrNotFound}
	uc := NewCounterUseCase(repo, branchRepo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	_, err := uc.CreateCounter(ctx, &model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"})
	if !errors.Is(err, exception.ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
	if repo.counter != nil {
		t.Fatalf("expected counter create blocked, got %+v", repo.counter)
	}
}
