package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestCreateService(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      model.CreateServiceRequest
		tenantID string
		wantErr  error
		wantRes  func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo)
	}{
		{
			name:     "Positive_UsesTenantContext",
			category: "positive",
			req:      model.CreateServiceRequest{Code: "reg", Name: "Registration"},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo) {
				assert.Equal(t, "tenant-1", res.TenantID)
			},
		},
		{
			name:     "Positive_IncludesPharmacyFlags",
			category: "positive",
			req: model.CreateServiceRequest{
				Code:                "pha",
				Name:                "Pharmacy",
				IsPharmacy:          true,
				IsPharmacyReception: true,
			},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo) {
				assert.True(t, res.IsPharmacy)
				assert.True(t, res.IsPharmacyReception)
				require.NotNil(t, repo.service)
				assert.True(t, repo.service.IsPharmacy)
				assert.True(t, repo.service.IsPharmacyReception)
			},
		},
		{
			name:     "Positive_SanitizesCodeAndName",
			category: "positive",
			req:      model.CreateServiceRequest{Code: " reg ", Name: " Registration "},
			tenantID: "tenant-1",
			wantErr:  nil,
			wantRes: func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo) {
				require.NotNil(t, repo.service)
				assert.Equal(t, "REG", repo.service.Code)
			},
		},
		{
			name:     "Negative_RequiresTenantContextForPharmacyFlags",
			category: "negative",
			req: model.CreateServiceRequest{
				Code:       "pha",
				Name:       "Pharmacy",
				IsPharmacy: true,
			},
			tenantID: "", // empty tenant
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_EmptyTenant",
			category: "vulnerability",
			req:      model.CreateServiceRequest{Code: "reg", Name: "Registration"},
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubServiceRepo{}
			uc := NewServiceUseCase(repo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.CreateService(ctx, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res, repo)
			}
		})
	}
}

func TestGetService(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		serviceID string
		tenantID  string
		stubErr   error
		stubRes   *entity.Service
		wantErr   error
		wantRes   func(t *testing.T, res *model.ServiceResponse)
	}{
		{
			name:      "Negative_RequiresTenantAndID_EmptyID",
			category:  "negative",
			serviceID: "",
			tenantID:  "tenant-1",
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			serviceID: "svc-1",
			tenantID:  "",
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Positive_Found",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			stubRes:   &entity.Service{ID: "svc-1", TenantID: "tenant-1"},
			wantErr:   nil,
			wantRes: func(t *testing.T, res *model.ServiceResponse) {
				assert.Equal(t, "svc-1", res.ID)
			},
		},
		{
			name:      "Negative_NotFound",
			category:  "negative",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			stubErr:   errors.New("db error"),
			wantErr:   exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubServiceRepo{err: tt.stubErr, service: tt.stubRes}
			uc := NewServiceUseCase(repo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.GetService(ctx, tt.serviceID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

func TestListServices(t *testing.T) {
	tests := []struct {
		name     string
		category string
		tenantID string
		stubErr  error
		stubRes  []*entity.Service
		wantErr  error
		wantRes  func(t *testing.T, res []model.ServiceResponse)
	}{
		{
			name:     "Negative_RequiresTenant",
			category: "negative",
			tenantID: "",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Positive_Found",
			category: "positive",
			tenantID: "tenant-1",
			stubRes:  []*entity.Service{{ID: "svc-1"}, {ID: "svc-2"}},
			wantErr:  nil,
			wantRes: func(t *testing.T, res []model.ServiceResponse) {
				assert.Len(t, res, 2)
			},
		},
		{
			name:     "Negative_RepoError",
			category: "negative",
			tenantID: "tenant-1",
			stubErr:  errors.New("db error"),
			wantErr:  errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubServiceRepo{err: tt.stubErr, list: tt.stubRes}
			uc := NewServiceUseCase(repo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.ListServices(ctx)

			if tt.wantErr != nil {
				if errors.Is(tt.wantErr, exception.ErrBadRequest) || errors.Is(tt.wantErr, exception.ErrNotFound) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
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

func TestUpdateService(t *testing.T) {
	flagFalse := false
	codeNew := " new "

	tests := []struct {
		name      string
		category  string
		serviceID string
		req       model.UpdateServiceRequest
		tenantID  string
		stubErr   error
		stubRes   *entity.Service
		wantErr   error
		wantRes   func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo)
	}{
		{
			name:      "Negative_EmptyID",
			category:  "negative",
			serviceID: "",
			tenantID:  "tenant-1",
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			serviceID: "svc-1",
			tenantID:  "",
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Negative_NotFound",
			category:  "negative",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			stubErr:   errors.New("not found"),
			wantErr:   exception.ErrNotFound,
		},
		{
			name:      "Positive_CanDisablePharmacyReception",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			req:       model.UpdateServiceRequest{IsPharmacyReception: &flagFalse},
			stubRes: &entity.Service{
				ID:                  "svc-1",
				TenantID:            "tenant-1",
				Code:                "PHA",
				Name:                "Pharmacy",
				Status:              entity.ServiceStatusActive,
				IsPharmacy:          true,
				IsPharmacyReception: true,
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo) {
				assert.False(t, res.IsPharmacyReception)
				require.NotNil(t, repo.service)
				assert.False(t, repo.service.IsPharmacyReception)
			},
		},
		{
			name:      "Positive_SanitizesCode",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			req:       model.UpdateServiceRequest{Code: &codeNew},
			stubRes: &entity.Service{
				ID:       "svc-1",
				TenantID: "tenant-1",
				Code:     "OLD",
				Name:     "Old",
				Status:   entity.ServiceStatusActive,
			},
			wantErr: nil,
			wantRes: func(t *testing.T, res *model.ServiceResponse, repo *stubServiceRepo) {
				assert.Equal(t, "NEW", res.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubServiceRepo{err: tt.stubErr, service: tt.stubRes}
			uc := NewServiceUseCase(repo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			res, err := uc.UpdateService(ctx, tt.serviceID, &tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res, repo)
			}
		})
	}
}

func TestDeleteService(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		serviceID string
		tenantID  string
		stubErr   error
		wantErr   error
	}{
		{
			name:      "Negative_EmptyID",
			category:  "negative",
			serviceID: "",
			tenantID:  "tenant-1",
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			serviceID: "svc-1",
			tenantID:  "",
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Positive_DeleteSuccess",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			wantErr:   nil,
		},
		{
			name:      "Negative_RepoError",
			category:  "negative",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			stubErr:   errors.New("db error"),
			wantErr:   errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &stubServiceRepo{err: tt.stubErr}
			uc := NewServiceUseCase(repo)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			err := uc.DeleteService(ctx, tt.serviceID)

			if tt.wantErr != nil {
				if errors.Is(tt.wantErr, exception.ErrBadRequest) || errors.Is(tt.wantErr, exception.ErrNotFound) {
					assert.ErrorIs(t, err, tt.wantErr)
				} else {
					assert.EqualError(t, err, tt.wantErr.Error())
				}
				return
			}
			require.NoError(t, err)
		})
	}
}
