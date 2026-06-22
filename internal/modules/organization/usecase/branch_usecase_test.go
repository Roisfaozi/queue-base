package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
)

type stubBranchRepo struct {
	branch *entity.Branch
	list   []*entity.Branch
	err    error
	seen   struct {
		tenantID string
		branchID string
	}
}

func (s *stubBranchRepo) Create(_ context.Context, branch *entity.Branch) error {
	s.branch = branch
	return s.err
}

func (s *stubBranchRepo) FindByID(_ context.Context, tenantID, branchID string) (*entity.Branch, error) {
	s.seen.tenantID = tenantID
	s.seen.branchID = branchID
	if s.err != nil {
		return nil, s.err
	}
	return s.branch, nil
}

func (s *stubBranchRepo) FindAll(_ context.Context, tenantID string) ([]*entity.Branch, error) {
	s.seen.tenantID = tenantID
	if s.err != nil {
		return nil, s.err
	}
	return s.list, nil
}

func (s *stubBranchRepo) Update(_ context.Context, branch *entity.Branch) error {
	s.branch = branch
	return s.err
}

func (s *stubBranchRepo) Delete(_ context.Context, tenantID, branchID string) error {
	s.seen.tenantID = tenantID
	s.seen.branchID = branchID
	return s.err
}

func TestResolveBranchUsesTenantScope(t *testing.T) {
	repo := &stubBranchRepo{branch: &entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive}}
	uc := NewBranchUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	res, err := uc.ResolveBranch(ctx, "branch-1")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.TenantID != "tenant-1" {
		t.Fatalf("expected tenant-1, got %s", res.TenantID)
	}
	if repo.seen.tenantID != "tenant-1" || repo.seen.branchID != "branch-1" {
		t.Fatalf("unexpected repo args: %+v", repo.seen)
	}
}

func TestResolveBranchRequiresTenantAndBranch(t *testing.T) {
	uc := NewBranchUseCase(&stubBranchRepo{})

	_, err := uc.ResolveBranch(context.Background(), "")
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestCreateBranchUsesTenantContext(t *testing.T) {
	repo := &stubBranchRepo{}
	uc := NewBranchUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	res, err := uc.CreateBranch(ctx, &model.CreateBranchRequest{Code: "main", Name: "Main Branch"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.TenantID != "tenant-1" {
		t.Fatalf("expected tenant-1, got %s", res.TenantID)
	}
}
