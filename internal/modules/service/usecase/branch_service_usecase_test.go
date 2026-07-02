package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubBranchServiceAuditLogger struct {
	entries []auditModel.CreateAuditLogRequest
}

func (s *stubBranchServiceAuditLogger) LogActivity(_ context.Context, req auditModel.CreateAuditLogRequest) error {
	s.entries = append(s.entries, req)
	return nil
}

type stubBranchServiceRepo struct {
	branchService *entity.BranchService
	err           error
}

type stubBranchRepo struct {
	branch *branchEntity.Branch
	err    error
}

func (s *stubBranchRepo) Create(_ context.Context, branch *branchEntity.Branch) error {
	s.branch = branch
	return s.err
}
func (s *stubBranchRepo) FindByID(_ context.Context, tenantID, branchID string) (*branchEntity.Branch, error) {
	if s.branch != nil && s.branch.ID == branchID && s.branch.TenantID == tenantID {
		return s.branch, nil
	}
	return nil, s.err
}
func (s *stubBranchRepo) FindAll(_ context.Context, tenantID string) ([]*branchEntity.Branch, error) {
	return []*branchEntity.Branch{}, s.err
}
func (s *stubBranchRepo) Update(_ context.Context, branch *branchEntity.Branch) error {
	s.branch = branch
	return s.err
}
func (s *stubBranchRepo) Delete(_ context.Context, tenantID, branchID string) error { return s.err }

func (s *stubBranchServiceRepo) Create(_ context.Context, bs *entity.BranchService) error {
	s.branchService = bs
	return s.err
}

func (s *stubBranchServiceRepo) FindByID(_ context.Context, tenantID, branchID, id string) (*entity.BranchService, error) {
	if s.branchService != nil && s.branchService.ID == id && s.branchService.TenantID == tenantID && s.branchService.BranchID == branchID {
		return s.branchService, nil
	}
	return nil, s.err
}

func (s *stubBranchServiceRepo) FindByService(_ context.Context, tenantID, branchID, serviceID string) (*entity.BranchService, error) {
	if s.branchService != nil && s.branchService.ServiceID == serviceID && s.branchService.TenantID == tenantID && s.branchService.BranchID == branchID {
		return s.branchService, nil
	}
	return nil, s.err
}

func (s *stubBranchServiceRepo) FindAll(_ context.Context, tenantID, branchID string) ([]*entity.BranchService, error) {
	if s.branchService != nil && s.branchService.TenantID == tenantID && s.branchService.BranchID == branchID {
		return []*entity.BranchService{s.branchService}, nil
	}
	return []*entity.BranchService{}, s.err
}

func (s *stubBranchServiceRepo) Update(_ context.Context, bs *entity.BranchService) error {
	s.branchService = bs
	return s.err
}

func (s *stubBranchServiceRepo) Delete(_ context.Context, tenantID, branchID, id string) error {
	return s.err
}

func TestBranchServiceUseCase_AuditHooks(t *testing.T) {
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	audit := &stubBranchServiceAuditLogger{}
	branchRepo := &stubBranchRepo{branch: &branchEntity.Branch{ID: "branch-1", TenantID: "tenant-1"}}
	serviceRepo := &stubServiceRepo{service: &entity.Service{ID: "svc-1", TenantID: "tenant-1"}}
	repo := &stubBranchServiceRepo{
		branchService: &entity.BranchService{ID: "bs-1", TenantID: "tenant-1", BranchID: "branch-1", ServiceID: "svc-1", IsActive: true},
	}

	uc := NewBranchServiceUseCase(repo, serviceRepo, branchRepo, audit)

	// Action 1: Create
	reqCreate := &model.CreateBranchServiceRequest{ServiceID: "svc-1"}
	created, err := uc.CreateBranchService(ctx, "branch-1", reqCreate)
	require.NoError(t, err)

	// Action 2: Update
	reqUpdate := &model.UpdateBranchServiceRequest{SortOrder: new(int)}
	_, err = uc.UpdateBranchService(ctx, "branch-1", created.ID, reqUpdate)
	require.NoError(t, err)

	// Action 3: Delete
	err = uc.DeleteBranchService(ctx, "branch-1", created.ID)
	require.NoError(t, err)

	// Validate Audit
	require.Len(t, audit.entries, 3)
	assert.Equal(t, "BRANCH_SERVICE_CREATE", audit.entries[0].Action)
	assert.Equal(t, "BRANCH_SERVICE_UPDATE", audit.entries[1].Action)
	assert.Equal(t, "BRANCH_SERVICE_DELETE", audit.entries[2].Action)
	assert.Equal(t, "branch_service", audit.entries[0].Entity)
}
