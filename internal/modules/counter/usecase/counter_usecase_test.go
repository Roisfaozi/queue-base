package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	organizationEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCounterAuditLogger struct {
	entries []auditModel.CreateAuditLogRequest
}

func (s *stubCounterAuditLogger) LogActivity(_ context.Context, req auditModel.CreateAuditLogRequest) error {
	s.entries = append(s.entries, req)
	return nil
}

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

func TestCreateCounter(t *testing.T) {
	tests := []struct {
		name     string
		req      model.CreateCounterRequest
		stubRepo struct {
			counter *entity.Counter
			err     error
		}
		stubBranchRepo struct {
			branch *organizationEntity.Branch
			err    error
		}
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.CounterResponse, repo *stubCounterRepo, branchRepo *stubCounterBranchRepo)
	}{
		{
			name: "Positive_CreatesCounterWithSanitization",
			req:  model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: " a1 ", Name: " Front Desk "},
			stubBranchRepo: struct {
				branch *organizationEntity.Branch
				err    error
			}{
				branch: &organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive},
				err:    nil,
			},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.CounterResponse, repo *stubCounterRepo, branchRepo *stubCounterBranchRepo) {
				assert.Equal(t, "tenant-1", res.TenantID)
				assert.Equal(t, "tenant-1", branchRepo.seen.tenantID)
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", branchRepo.seen.branchID)
				require.NotNil(t, repo.counter)
				assert.Equal(t, "A1", repo.counter.Code)
				assert.Equal(t, "Front Desk", repo.counter.Name)
			},
		},
		{
			name:     "Negative_RequiresTenant",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name: "Vulnerability_RejectsCrossTenantBranch",
			req:  model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
			stubBranchRepo: struct {
				branch *organizationEntity.Branch
				err    error
			}{
				branch: nil,
				err:    exception.ErrNotFound,
			},
			tenantID: "tenant-1",
			wantErr:  exception.ErrForbidden,
			wantRes: func(t *testing.T, res *model.CounterResponse, repo *stubCounterRepo, branchRepo *stubCounterBranchRepo) {
				assert.Nil(t, repo.counter)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubCounterRepo{
				counter: tt.stubRepo.counter,
				err:     tt.stubRepo.err,
			}
			branch := tt.stubBranchRepo.branch
			if branch == nil && tt.stubBranchRepo.err == nil {
				branch = &organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive}
			}
			branchRepo := &stubCounterBranchRepo{
				branch: branch,
				err:    tt.stubBranchRepo.err,
			}
			uc := NewCounterUseCase(repo, branchRepo, &stubCounterBranchServiceRepo{isActive: true})

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.CreateCounter(ctx, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.wantRes != nil {
					tt.wantRes(t, res, repo, branchRepo)
				}
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res, repo, branchRepo)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	code := " b2 "
	name := " Front Office "

	tests := []struct {
		name      string
		counterID string
		req       model.UpdateCounterRequest
		stubRepo  struct {
			counter *entity.Counter
			err     error
		}
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.CounterResponse, repo *stubCounterRepo)
	}{
		{
			name:      "Positive_SanitizesFieldsOnUpdate",
			counterID: "counter-1",
			req:       model.UpdateCounterRequest{Code: &code, Name: &name},
			stubRepo: struct {
				counter *entity.Counter
				err     error
			}{
				counter: &entity.Counter{ID: "counter-1", TenantID: "tenant-1", BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk", Status: entity.CounterStatusActive},
				err:     nil,
			},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.CounterResponse, repo *stubCounterRepo) {
				assert.Equal(t, "B2", res.Code)
				assert.Equal(t, "Front Office", res.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubCounterRepo{
				counter: tt.stubRepo.counter,
				err:     tt.stubRepo.err,
			}
			uc := NewCounterUseCase(repo, &stubCounterBranchRepo{branch: &organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive}}, &stubCounterBranchServiceRepo{isActive: true})

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.UpdateCounter(ctx, tt.counterID, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.wantRes != nil {
					tt.wantRes(t, res, repo)
				}
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res, repo)
			}
		})
	}
}

type stubCounterBranchServiceRepo struct {
	err      error
	isActive bool
}

func (s *stubCounterBranchServiceRepo) Create(_ context.Context, bs *serviceEntity.BranchService) error {
	return nil
}
func (s *stubCounterBranchServiceRepo) FindByID(_ context.Context, tenantID, branchID, id string) (*serviceEntity.BranchService, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &serviceEntity.BranchService{ID: id, TenantID: tenantID, BranchID: branchID, IsActive: s.isActive}, nil
}
func (s *stubCounterBranchServiceRepo) FindByService(_ context.Context, tenantID, branchID, serviceID string) (*serviceEntity.BranchService, error) {
	return nil, nil
}
func (s *stubCounterBranchServiceRepo) FindAll(_ context.Context, tenantID, branchID string) ([]*serviceEntity.BranchService, error) {
	return nil, nil
}
func (s *stubCounterBranchServiceRepo) Update(_ context.Context, bs *serviceEntity.BranchService) error {
	return nil
}
func (s *stubCounterBranchServiceRepo) Delete(_ context.Context, tenantID, branchID, id string) error {
	return nil
}

func TestCounterDeactivationEdgeCase(t *testing.T) {
	// ponytail: no cascade — deactivating parent does not auto-deactivate counter
	repo := &stubCounterRepo{}
	branchRepo := &stubCounterBranchRepo{
		branch: &organizationEntity.Branch{ID: "b-1", TenantID: "t-1", Status: organizationEntity.BranchStatusActive},
	}
	bsRepo := &stubCounterBranchServiceRepo{isActive: true}
	uc := NewCounterUseCase(repo, branchRepo, bsRepo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	req := &model.CreateCounterRequest{
		BranchID:        "b-1",
		BranchServiceID: "bs-1",
		Code:            "C1",
		Name:            "Counter 1",
		DisplayName:     "Counter 1",
	}
	res, err := uc.CreateCounter(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, entity.CounterStatusActive, res.Status)

	// Now deactivate branch — counter should still be active (no cascade)
	branchRepo.branch.Status = organizationEntity.BranchStatusInactive

	// Counter still accessible via GetCounter
	getRes, err := uc.GetCounter(ctx, res.ID)
	require.NoError(t, err)
	assert.Equal(t, entity.CounterStatusActive, getRes.Status, "counter remains active after branch deactivation (no cascade)")

	// New counter creation should be rejected
	req2 := &model.CreateCounterRequest{
		BranchID:        "b-1",
		BranchServiceID: "bs-1",
		Code:            "C2",
		Name:            "Counter 2",
		DisplayName:     "Counter 2",
	}
	_, err = uc.CreateCounter(ctx, req2)
	assert.ErrorIs(t, err, exception.ErrForbidden, "create on inactive branch should fail")
}

func TestCounterServiceRelationGuard(t *testing.T) {
	tests := []struct {
		name            string
		req             model.CreateCounterRequest
		stubBranchRepo  *stubCounterBranchRepo
		stubServiceRepo *stubCounterBranchServiceRepo
		wantError       error
	}{
		{
			name: "Vulnerability_RejectsInvalidBranchServiceRelation",
			req: model.CreateCounterRequest{
				BranchID:        "branch-1",
				BranchServiceID: "service-unknown",
				Code:            "A1",
			},
			stubBranchRepo: &stubCounterBranchRepo{
				branch: &organizationEntity.Branch{ID: "branch-1", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive},
			},
			stubServiceRepo: &stubCounterBranchServiceRepo{err: exception.ErrNotFound},
			wantError:       exception.ErrForbidden,
		},
		{
			name: "Vulnerability_RejectsInactiveBranchServiceRelation",
			req: model.CreateCounterRequest{
				BranchID:        "branch-1",
				BranchServiceID: "service-inactive",
				Code:            "A1",
			},
			stubBranchRepo: &stubCounterBranchRepo{
				branch: &organizationEntity.Branch{ID: "branch-1", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive},
			},
			stubServiceRepo: &stubCounterBranchServiceRepo{
				err: nil,
			},
			wantError: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubCounterRepo{}
			uc := NewCounterUseCase(repo, tt.stubBranchRepo, tt.stubServiceRepo)

			ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

			// Intercept stubServiceRepo to return inactive if needed
			if tt.name == "Vulnerability_RejectsInactiveBranchServiceRelation" {
				tt.stubServiceRepo.isActive = false
			} else if tt.stubServiceRepo != nil && tt.stubServiceRepo.err == nil {
				tt.stubServiceRepo.isActive = true
			}

			_, err := uc.CreateCounter(ctx, &tt.req)
			assert.ErrorIs(t, err, tt.wantError)
		})
	}
}

func TestCounterAuditHooks(t *testing.T) {
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	t.Run("Positive_LogsAuditActivitiesForCreateUpdateDelete", func(t *testing.T) {
		audit := &stubCounterAuditLogger{}
		repo := &stubCounterRepo{counter: &entity.Counter{ID: "counter-1", TenantID: "tenant-1", BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Status: entity.CounterStatusActive}}
		uc := NewCounterUseCase(repo, &stubCounterBranchRepo{branch: &organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1", Status: organizationEntity.BranchStatusActive}}, &stubCounterBranchServiceRepo{isActive: true}, audit)

		_, err := uc.CreateCounter(ctx, &model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Counter A"})
		require.NoError(t, err)
		_, err = uc.UpdateCounter(ctx, "counter-1", &model.UpdateCounterRequest{})
		require.NoError(t, err)
		err = uc.DeleteCounter(ctx, "counter-1")
		require.NoError(t, err)

		require.Len(t, audit.entries, 3)
		assert.Equal(t, "COUNTER_CREATE", audit.entries[0].Action)
		assert.Equal(t, "COUNTER_UPDATE", audit.entries[1].Action)
		assert.Equal(t, "COUNTER_DELETE", audit.entries[2].Action)
		assert.Equal(t, "counter", audit.entries[0].Entity)
	})
}
