package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
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
	if repo.branch == nil || repo.branch.Code != "MAIN" || repo.branch.Name != "Main Branch" {
		t.Fatalf("expected sanitized branch, got %+v", repo.branch)
	}
}

func TestListBranchesUsesTenantScope(t *testing.T) {
	repo := &stubBranchRepo{list: []*entity.Branch{{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main"}}}
	uc := NewBranchUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	res, err := uc.ListBranches(ctx)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if len(res) != 1 || res[0].TenantID != "tenant-1" {
		t.Fatalf("unexpected list response: %+v", res)
	}
	if repo.seen.tenantID != "tenant-1" {
		t.Fatalf("expected tenant scope, got %+v", repo.seen)
	}
}

func TestUpdateBranchSanitizesFields(t *testing.T) {
	code := " sub "
	name := " Branch Office "
	repo := &stubBranchRepo{branch: &entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusActive}}
	uc := NewBranchUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	res, err := uc.UpdateBranch(ctx, "branch-1", &model.UpdateBranchRequest{Code: &code, Name: &name})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if res.Code != "SUB" || res.Name != "Branch Office" {
		t.Fatalf("expected sanitized response, got %+v", res)
	}
}

func TestDeleteBranchRequiresBranchID(t *testing.T) {
	uc := NewBranchUseCase(&stubBranchRepo{})
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	err := uc.DeleteBranch(ctx, "")
	if !errors.Is(err, exception.ErrBadRequest) {
		t.Fatalf("expected ErrBadRequest, got %v", err)
	}
}

func TestResolveBranchRejectsCrossTenantLookup(t *testing.T) {
	repo := &stubBranchRepo{err: exception.ErrNotFound}
	uc := NewBranchUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	_, err := uc.ResolveBranch(ctx, "branch-2")
	if !errors.Is(err, exception.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
	if repo.seen.tenantID != "tenant-1" || repo.seen.branchID != "branch-2" {
		t.Fatalf("unexpected repo args: %+v", repo.seen)
	}
}
