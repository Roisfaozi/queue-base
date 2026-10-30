package usecase_test

import (
	"context"
	"errors"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	serviceMocks "github.com/Roisfaozi/queue-base/internal/modules/service/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/service/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServiceAuditHooks(t *testing.T) {
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	audit := serviceMocks.NewMockAuditLogger(t)
	repo := serviceMocks.NewMockServiceRepository(t)
	uc := usecase.NewServiceUseCase(repo, audit)

	repo.On("Create", mock.Anything, mock.Anything).Return(nil).Once()
	audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "SERVICE_CREATE" && req.Entity == "service"
	})).Return(nil).Once()

	_, err := uc.CreateService(ctx, &model.CreateServiceRequest{Code: "new", Name: "New"})
	require.NoError(t, err)

	existingService := &entity.Service{ID: "svc-1", TenantID: "tenant-1", Code: "REG", Name: "Registration", Status: entity.ServiceStatusActive}
	repo.On("FindByID", mock.Anything, "tenant-1", "svc-1").Return(existingService, nil).Once()
	repo.On("Update", mock.Anything, mock.Anything).Return(nil).Once()
	audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "SERVICE_UPDATE" && req.Entity == "service"
	})).Return(nil).Once()

	_, err = uc.UpdateService(ctx, "svc-1", &model.UpdateServiceRequest{})
	require.NoError(t, err)

	repo.On("Delete", mock.Anything, "tenant-1", "svc-1").Return(nil).Once()
	audit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		return req.Action == "SERVICE_DELETE" && req.Entity == "service"
	})).Return(nil).Once()

	err = uc.DeleteService(ctx, "svc-1")
	require.NoError(t, err)
}

func TestCreateService(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      model.CreateServiceRequest
		tenantID string
		setup    func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service)
		wantErr  error
		assert   func(t *testing.T, res *model.ServiceResponse, createdService *entity.Service)
	}{
		{
			name:     "Positive_UsesTenantContext",
			category: "positive",
			req:      model.CreateServiceRequest{Code: "reg", Name: "Registration"},
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(s *entity.Service) bool {
					*createdService = s
					return s.Code == "REG" && s.TenantID == "tenant-1"
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse, createdService *entity.Service) {
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
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(s *entity.Service) bool {
					*createdService = s
					return s.IsPharmacy && s.IsPharmacyReception
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse, createdService *entity.Service) {
				assert.True(t, res.IsPharmacy)
				assert.True(t, res.IsPharmacyReception)
				require.NotNil(t, createdService)
				assert.True(t, createdService.IsPharmacy)
				assert.True(t, createdService.IsPharmacyReception)
			},
		},
		{
			name:     "Positive_SanitizesCodeAndName",
			category: "positive",
			req:      model.CreateServiceRequest{Code: " reg ", Name: " Registration "},
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(s *entity.Service) bool {
					*createdService = s
					return s.Code == "REG" && s.Name == "Registration"
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse, createdService *entity.Service) {
				require.NotNil(t, createdService)
				assert.Equal(t, "REG", createdService.Code)
			},
		},
		{
			name:     "Positive_PersistsAudioAndNarrativeFields",
			category: "positive",
			req: model.CreateServiceRequest{
				Code:                   "aud",
				Name:                   "Audio Service",
				AudioID:                "audio-1",
				AudioEN:                "audio-en-1",
				NarrativeInstructionID: "narr-1",
				NarrativeInstructionEN: "narr-en-1",
			},
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service) {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(s *entity.Service) bool {
					*createdService = s
					return s.AudioID == "audio-1" && s.AudioEN == "audio-en-1" &&
						s.NarrativeInstructionID == "narr-1" && s.NarrativeInstructionEN == "narr-en-1"
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse, createdService *entity.Service) {
				require.NotNil(t, createdService)
				assert.Equal(t, "audio-1", createdService.AudioID)
				assert.Equal(t, "audio-en-1", createdService.AudioEN)
				assert.Equal(t, "narr-1", createdService.NarrativeInstructionID)
				assert.Equal(t, "narr-en-1", createdService.NarrativeInstructionEN)
				assert.Equal(t, "audio-1", res.AudioID)
				assert.Equal(t, "narr-en-1", res.NarrativeInstructionEN)
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
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service) {
			},
			wantErr: exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_EmptyTenant",
			category: "vulnerability",
			req:      model.CreateServiceRequest{Code: "reg", Name: "Registration"},
			tenantID: "",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, createdService **entity.Service) {
			},
			wantErr: exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := serviceMocks.NewMockServiceRepository(t)
			var createdService *entity.Service
			if tt.setup != nil {
				tt.setup(t, repo, &createdService)
			}
			uc := usecase.NewServiceUseCase(repo)

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
			if tt.assert != nil {
				tt.assert(t, res, createdService)
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
		setup     func(t *testing.T, repo *serviceMocks.MockServiceRepository)
		wantErr   error
		assert    func(t *testing.T, res *model.ServiceResponse)
	}{
		{
			name:      "Negative_RequiresTenantAndID_EmptyID",
			category:  "negative",
			serviceID: "",
			tenantID:  "tenant-1",
			setup:     func(t *testing.T, repo *serviceMocks.MockServiceRepository) {},
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			serviceID: "svc-1",
			tenantID:  "",
			setup:     func(t *testing.T, repo *serviceMocks.MockServiceRepository) {},
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Positive_Found",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository) {
				repo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(&entity.Service{ID: "svc-1", TenantID: "tenant-1", AudioID: "audio-1", NarrativeInstructionEN: "narr-en-1"}, nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse) {
				assert.Equal(t, "svc-1", res.ID)
				assert.Equal(t, "audio-1", res.AudioID)
			},
		},
		{
			name:      "Negative_NotFound",
			category:  "negative",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository) {
				repo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(nil, errors.New("db error")).Once()
			},
			wantErr: exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := serviceMocks.NewMockServiceRepository(t)
			tt.setup(t, repo)
			uc := usecase.NewServiceUseCase(repo)

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
			if tt.assert != nil {
				tt.assert(t, res)
			}
		})
	}
}

func TestListServices(t *testing.T) {
	tests := []struct {
		name     string
		category string
		tenantID string
		setup    func(t *testing.T, repo *serviceMocks.MockServiceRepository)
		wantErr  error
		assert   func(t *testing.T, res []model.ServiceResponse)
	}{
		{
			name:     "Negative_RequiresTenant",
			category: "negative",
			tenantID: "",
			setup:    func(t *testing.T, repo *serviceMocks.MockServiceRepository) {},
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Positive_Found",
			category: "positive",
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository) {
				repo.On("FindAll", mock.Anything, "tenant-1").
					Return([]*entity.Service{{ID: "svc-1"}, {ID: "svc-2"}}, nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res []model.ServiceResponse) {
				assert.Len(t, res, 2)
			},
		},
		{
			name:     "Negative_RepoError",
			category: "negative",
			tenantID: "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository) {
				repo.On("FindAll", mock.Anything, "tenant-1").
					Return(nil, errors.New("db error")).Once()
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := serviceMocks.NewMockServiceRepository(t)
			tt.setup(t, repo)
			uc := usecase.NewServiceUseCase(repo)

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
			if tt.assert != nil {
				tt.assert(t, res)
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
		setup     func(t *testing.T, repo *serviceMocks.MockServiceRepository, updatedService **entity.Service)
		wantErr   error
		assert    func(t *testing.T, res *model.ServiceResponse, updatedService *entity.Service)
	}{
		{
			name:      "Negative_EmptyID",
			category:  "negative",
			serviceID: "",
			tenantID:  "tenant-1",
			setup:     func(t *testing.T, repo *serviceMocks.MockServiceRepository, updatedService **entity.Service) {},
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			serviceID: "svc-1",
			tenantID:  "",
			setup:     func(t *testing.T, repo *serviceMocks.MockServiceRepository, updatedService **entity.Service) {},
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Negative_NotFound",
			category:  "negative",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, updatedService **entity.Service) {
				repo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(nil, errors.New("not found")).Once()
			},
			wantErr: exception.ErrNotFound,
		},
		{
			name:      "Positive_CanDisablePharmacyReception",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			req:       model.UpdateServiceRequest{IsPharmacyReception: &flagFalse},
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, updatedService **entity.Service) {
				repo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(&entity.Service{
						ID:                  "svc-1",
						TenantID:            "tenant-1",
						Code:                "PHA",
						Name:                "Pharmacy",
						Status:              entity.ServiceStatusActive,
						IsPharmacy:          true,
						IsPharmacyReception: true,
					}, nil).Once()
				repo.On("Update", mock.Anything, mock.MatchedBy(func(s *entity.Service) bool {
					*updatedService = s
					return !s.IsPharmacyReception
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse, updatedService *entity.Service) {
				assert.False(t, res.IsPharmacyReception)
				require.NotNil(t, updatedService)
				assert.False(t, updatedService.IsPharmacyReception)
			},
		},
		{
			name:      "Positive_SanitizesCode",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			req:       model.UpdateServiceRequest{Code: &codeNew},
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository, updatedService **entity.Service) {
				repo.On("FindByID", mock.Anything, "tenant-1", "svc-1").
					Return(&entity.Service{
						ID:       "svc-1",
						TenantID: "tenant-1",
						Code:     "OLD",
						Name:     "Old",
						Status:   entity.ServiceStatusActive,
					}, nil).Once()
				repo.On("Update", mock.Anything, mock.MatchedBy(func(s *entity.Service) bool {
					*updatedService = s
					return s.Code == "NEW"
				})).Return(nil).Once()
			},
			wantErr: nil,
			assert: func(t *testing.T, res *model.ServiceResponse, updatedService *entity.Service) {
				assert.Equal(t, "NEW", res.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := serviceMocks.NewMockServiceRepository(t)
			var updatedService *entity.Service
			if tt.setup != nil {
				tt.setup(t, repo, &updatedService)
			}
			uc := usecase.NewServiceUseCase(repo)

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
			if tt.assert != nil {
				tt.assert(t, res, updatedService)
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
		setup     func(t *testing.T, repo *serviceMocks.MockServiceRepository)
		wantErr   error
	}{
		{
			name:      "Negative_EmptyID",
			category:  "negative",
			serviceID: "",
			tenantID:  "tenant-1",
			setup:     func(t *testing.T, repo *serviceMocks.MockServiceRepository) {},
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Vulnerability_EmptyTenant",
			category:  "vulnerability",
			serviceID: "svc-1",
			tenantID:  "",
			setup:     func(t *testing.T, repo *serviceMocks.MockServiceRepository) {},
			wantErr:   exception.ErrBadRequest,
		},
		{
			name:      "Positive_DeleteSuccess",
			category:  "positive",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository) {
				repo.On("Delete", mock.Anything, "tenant-1", "svc-1").Return(nil).Once()
			},
			wantErr: nil,
		},
		{
			name:      "Negative_RepoError",
			category:  "negative",
			serviceID: "svc-1",
			tenantID:  "tenant-1",
			setup: func(t *testing.T, repo *serviceMocks.MockServiceRepository) {
				repo.On("Delete", mock.Anything, "tenant-1", "svc-1").Return(errors.New("db error")).Once()
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := serviceMocks.NewMockServiceRepository(t)
			tt.setup(t, repo)
			uc := usecase.NewServiceUseCase(repo)

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
