package usecase

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	organizationEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCreateCounter(t *testing.T) {
	tests := []struct {
		name     string
		category string
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
			name:     "Positive_CreatesCounterWithSanitization",
			category: "positive",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: " a1 ", Name: " Front Desk "},
			stubBranchRepo: struct {
				branch *organizationEntity.Branch
				err    error
			}{
				branch: &organizationEntity.Branch{ID: "550e8400-e29b-41d4-a716-446655440000", TenantID: "tenant-1"},
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
			category: "negative",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_RejectsCrossTenantBranch",
			category: "vulnerability",
			req:      model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Desk"},
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
			branchRepo := &stubCounterBranchRepo{
				branch: tt.stubBranchRepo.branch,
				err:    tt.stubBranchRepo.err,
			}
			uc := NewCounterUseCase(repo, branchRepo)

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
		category  string
		counterID string
		req       model.UpdateCounterRequest
		stubRepo  struct {
			counter *entity.Counter
			err     error
		}
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.CounterResponse)
	}{
		{
			name:      "Positive_SanitizesFieldsOnUpdate",
			category:  "positive",
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
			wantRes: func(t *testing.T, res *model.CounterResponse) {
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
			uc := NewCounterUseCase(repo, &stubCounterBranchRepo{})

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.UpdateCounter(ctx, tt.counterID, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				if tt.wantRes != nil {
					tt.wantRes(t, res)
				}
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}
