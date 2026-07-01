package test

import (
	"context"
	"errors"
	"io"
	"testing"

	mocking "github.com/Roisfaozi/queue-base/internal/mocking"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	permissionMocks "github.com/Roisfaozi/queue-base/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type organizationTestDeps struct {
	OrgRepo    *mocks.MockOrganizationRepository
	MemberRepo *mocks.MockOrganizationMemberRepository
	OrgReader  *mocks.MockIOrganizationReader
	TM         *mocking.MockWithTransactionManager
	Enforcer   *permissionMocks.MockIEnforcer
}

func setupOrganizationTest() (*organizationTestDeps, usecase.OrganizationUseCase) {
	mockEnforcer := new(permissionMocks.MockIEnforcer)
	deps := &organizationTestDeps{
		OrgRepo:    new(mocks.MockOrganizationRepository),
		MemberRepo: new(mocks.MockOrganizationMemberRepository),
		OrgReader:  new(mocks.MockIOrganizationReader),
		TM:         new(mocking.MockWithTransactionManager),
		Enforcer:   mockEnforcer,
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.FatalLevel)

	mockEnforcer.On("WithContext", mock.Anything).Maybe().Return(mockEnforcer)
	mockEnforcer.On("LoadPolicy").Maybe().Return(nil)
	deps.OrgReader.On("InvalidateOrganizationCache", mock.Anything, mock.Anything).Maybe().Return(nil)

	uc := usecase.NewOrganizationUseCase(log, deps.TM, deps.OrgRepo, deps.MemberRepo, deps.OrgReader, deps.Enforcer)

	return deps, uc
}

func TestOrganizationUseCase(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_CreateOrganization_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				userID := "user-123"
				req := &model.CreateOrganizationRequest{Name: "Acme Corp", Slug: "acme-corp"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
				deps.OrgRepo.On("Create", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
					return org.Name == req.Name && org.Slug == req.Slug && org.OwnerID == userID
				}), usecase.DefaultOwnerRoleID).Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:admin", "global"}).Return([][]string{
					{"role:admin", "global", "/api/v1/projects", "POST"},
				}, nil)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:user", "global"}).Return([][]string{
					{"role:user", "global", "/api/v1/projects", "GET"},
				}, nil)
				deps.Enforcer.On("AddPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 4 && params[0] == "role:admin" && params[2] == "/api/v1/projects" && params[3] == "POST"
				})).Return(true, nil)
				deps.Enforcer.On("AddPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 4 && params[0] == "role:user" && params[2] == "/api/v1/projects" && params[3] == "GET"
				})).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 3 && params[0] == usecase.DefaultOwnerRoleID && params[1] == "role:admin"
				})).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 3 && params[1] == usecase.DefaultOwnerRoleID
				})).Return(true, nil)

				res, err := uc.CreateOrganization(ctx, userID, req)

				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, req.Name, res.Name)
				deps.OrgRepo.AssertExpectations(t)
				deps.Enforcer.AssertExpectations(t)
			},
		},
		{
			name:     "Negative_CreateOrganization_SlugExists",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				userID := "user-123"
				req := &model.CreateOrganizationRequest{Name: "Acme Corp", Slug: "exists"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrConflict)

				deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(true, nil)

				res, err := uc.CreateOrganization(ctx, userID, req)

				assert.ErrorIs(t, err, exception.ErrConflict)
				assert.Nil(t, res)
			},
		},
		{
			name:     "Negative_CreateOrganization_SlugCheckFails",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, errors.New("db error"))

				_, err := uc.CreateOrganization(ctx, "u1", req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_CreateOrganization_CreateFails",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
				deps.OrgRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(errors.New("create error"))

				_, err := uc.CreateOrganization(ctx, "u1", req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_CreateOrganization_EnforcerFails",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
				deps.OrgRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:admin", "global"}).Return([][]string{
					{"role:admin", "global", "/api/v1/projects", "POST"},
				}, nil)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:user", "global"}).Return([][]string{
					{"role:user", "global", "/api/v1/projects", "GET"},
				}, nil)
				deps.Enforcer.On("AddPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 4 && params[0] == "role:admin" && params[2] == "/api/v1/projects" && params[3] == "POST"
				})).Return(true, nil)
				deps.Enforcer.On("AddPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 4 && params[0] == "role:user" && params[2] == "/api/v1/projects" && params[3] == "GET"
				})).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 3 && params[0] == usecase.DefaultOwnerRoleID && params[1] == "role:admin"
				})).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 3 && params[0] == "u1" && params[1] == usecase.DefaultOwnerRoleID
				})).Return(false, errors.New("casbin error"))

				_, err := uc.CreateOrganization(ctx, "u1", req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_CreateOrganization_BootstrapPolicyFails",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				req := &model.CreateOrganizationRequest{Name: "Acme", Slug: "acme"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("SlugExists", ctx, req.Slug).Return(false, nil)
				deps.OrgRepo.On("Create", ctx, mock.Anything, mock.Anything).Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:admin", "global"}).Return(nil, errors.New("casbin read error"))

				_, err := uc.CreateOrganization(ctx, "u1", req)
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},

		{
			name:     "Positive_GetOrganization_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				org := &entity.Organization{ID: "org-1", Name: "Org 1"}

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)

				res, err := uc.GetOrganization(ctx, "org-1")
				assert.NoError(t, err)
				assert.Equal(t, org.ID, res.ID)
			},
		},
		{
			name:     "Negative_GetOrganization_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)

				_, err := uc.GetOrganization(ctx, "org-1")
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_GetOrganization_RepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

				_, err := uc.GetOrganization(ctx, "org-1")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Positive_GetOrganizationBySlug_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				org := &entity.Organization{ID: "org-1", Slug: "slug-1"}

				deps.OrgRepo.On("FindBySlug", ctx, "slug-1").Return(org, nil)

				res, err := uc.GetOrganizationBySlug(ctx, "slug-1")
				assert.NoError(t, err)
				assert.Equal(t, org.Slug, res.Slug)
			},
		},
		{
			name:     "Negative_GetOrganizationBySlug_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.OrgRepo.On("FindBySlug", ctx, "slug-1").Return(nil, nil)

				_, err := uc.GetOrganizationBySlug(ctx, "slug-1")
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_GetOrganizationBySlug_RepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.OrgRepo.On("FindBySlug", ctx, "slug-1").Return(nil, errors.New("db error"))

				_, err := uc.GetOrganizationBySlug(ctx, "slug-1")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Positive_UpdateOrganization_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "owner-1")
				orgID := "org-1"
				req := &model.UpdateOrganizationRequest{Name: "New Name"}
				existingOrg := &entity.Organization{ID: orgID, Name: "Old Name", OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
				deps.OrgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
					return org.Name == "New Name"
				})).Return(nil)

				res, err := uc.UpdateOrganization(ctx, orgID, req)
				assert.NoError(t, err)
				assert.Equal(t, "New Name", res.Name)
			},
		},
		{
			name:     "Positive_UpdateOrganization_SettingsUpdate",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "owner-1")
				orgID := "org-1"
				settings := map[string]interface{}{"theme": "dark"}
				req := &model.UpdateOrganizationRequest{Settings: settings}
				existingOrg := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
				deps.OrgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
					return org.Settings["theme"] == "dark"
				})).Return(nil)

				res, err := uc.UpdateOrganization(ctx, orgID, req)
				assert.NoError(t, err)
				assert.Equal(t, "dark", res.Settings["theme"])
			},
		},
		{
			name:     "Positive_UpdateOrganization_NoChange",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "owner-1")
				orgID := "org-1"
				req := &model.UpdateOrganizationRequest{}
				existingOrg := &entity.Organization{ID: orgID, Name: "Name", OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
				deps.OrgRepo.On("Update", ctx, mock.Anything).Return(nil)

				res, err := uc.UpdateOrganization(ctx, orgID, req)
				assert.NoError(t, err)
				assert.Equal(t, "Name", res.Name)
			},
		},
		{
			name:     "Negative_UpdateOrganization_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrNotFound)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)

				_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{})
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_UpdateOrganization_FindError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

				_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{})
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_UpdateOrganization_UpdateError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "owner-1")
				existingOrg := &entity.Organization{ID: "org-1", OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(existingOrg, nil)
				deps.OrgRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

				_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{Name: "New"})
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Security_UpdateOrganization_ForbiddenForNonManagerActor",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "user-2")
				existingOrg := &entity.Organization{ID: "org-1", OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrForbidden)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(existingOrg, nil)
				deps.MemberRepo.On("CheckMembership", ctx, "org-1", "user-2").Return(true, nil)
				deps.MemberRepo.On("GetMemberRole", ctx, "org-1", "user-2").Return("role:user", nil)

				_, err := uc.UpdateOrganization(ctx, "org-1", &model.UpdateOrganizationRequest{Name: "Blocked"})
				assert.ErrorIs(t, err, exception.ErrForbidden)
			},
		},

		{
			name:     "Positive_GetUserOrganizations_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				orgs := []*entity.Organization{{ID: "org-1"}}

				deps.OrgRepo.On("FindUserOrganizations", ctx, "user-1").Return(orgs, nil)

				res, err := uc.GetUserOrganizations(ctx, "user-1")
				assert.NoError(t, err)
				assert.Equal(t, 1, res.Total)
			},
		},
		{
			name:     "Positive_GetUserOrganizations_Empty",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.OrgRepo.On("FindUserOrganizations", ctx, "user-1").Return([]*entity.Organization{}, nil)

				res, err := uc.GetUserOrganizations(ctx, "user-1")
				assert.NoError(t, err)
				assert.Equal(t, 0, res.Total)
			},
		},
		{
			name:     "Negative_GetUserOrganizations_RepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.OrgRepo.On("FindUserOrganizations", ctx, "user-1").Return(nil, errors.New("db error"))

				_, err := uc.GetUserOrganizations(ctx, "user-1")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Positive_DeleteOrganization_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "owner-1"
				org := &entity.Organization{ID: orgID, OwnerID: userID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.OrgRepo.On("Delete", ctx, orgID).Return(nil)

				err := uc.DeleteOrganization(ctx, orgID, userID)
				assert.NoError(t, err)
				deps.OrgReader.AssertCalled(t, "InvalidateOrganizationCache", mock.Anything, orgID)
			},
		},
		{
			name:     "Negative_DeleteOrganization_NotOwner",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				orgID := "org-1"
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrForbidden)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)

				err := uc.DeleteOrganization(ctx, orgID, "other-user")
				assert.ErrorIs(t, err, exception.ErrForbidden)
			},
		},
		{
			name:     "Negative_DeleteOrganization_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrNotFound)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, nil)

				err := uc.DeleteOrganization(ctx, "org-1", "user-1")
				assert.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_DeleteOrganization_DeleteError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()
				org := &entity.Organization{ID: "org-1", OwnerID: "user-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(org, nil)
				deps.OrgRepo.On("Delete", ctx, "org-1").Return(errors.New("db error"))

				err := uc.DeleteOrganization(ctx, "org-1", "user-1")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_DeleteOrganization_FindError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, "org-1").Return(nil, errors.New("db error"))

				err := uc.DeleteOrganization(ctx, "org-1", "user-1")
				assert.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},

		{
			name:     "Security_CreateOrganization_XSS",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := context.Background()

				xssPayload := "<script>alert('xss')</script>"
				request := &model.CreateOrganizationRequest{
					Name: "Acme " + xssPayload,
					Slug: "acme-xss",
				}
				userID := "user-123"

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("SlugExists", ctx, "acme-xss").Return(false, nil)
				deps.OrgRepo.On("Create", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
					return org.Name == request.Name
				}), usecase.DefaultOwnerRoleID).Return(nil)

				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:admin", "global"}).Return([][]string{
					{"role:admin", "global", "/api/v1/projects", "POST"},
				}, nil)
				deps.Enforcer.On("GetFilteredPolicy", 0, []string{"role:user", "global"}).Return([][]string{
					{"role:user", "global", "/api/v1/projects", "GET"},
				}, nil)
				deps.Enforcer.On("AddPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 4 && params[0] == "role:admin" && params[2] == "/api/v1/projects" && params[3] == "POST"
				})).Return(true, nil)
				deps.Enforcer.On("AddPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 4 && params[0] == "role:user" && params[2] == "/api/v1/projects" && params[3] == "GET"
				})).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 3 && params[0] == usecase.DefaultOwnerRoleID && params[1] == "role:admin"
				})).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.MatchedBy(func(params []interface{}) bool {
					return len(params) == 3 && params[1] == usecase.DefaultOwnerRoleID
				})).Return(true, nil)

				response, err := uc.CreateOrganization(ctx, userID, request)

				assert.NoError(t, err)
				assert.Equal(t, request.Name, response.Name)
			},
		},
		{
			name:     "Positive_UpdateOrganization_Settings",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "owner-1")

				orgID := "org-123"
				settings := map[string]interface{}{
					"theme":       "dark",
					"mfa_enabled": true,
					"max_users":   100,
				}

				existingOrg := &entity.Organization{
					ID:       orgID,
					Name:     "Acme Corp",
					OwnerID:  "owner-1",
					Settings: nil,
				}

				request := &model.UpdateOrganizationRequest{
					Settings: settings,
				}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
				deps.OrgRepo.On("Update", ctx, mock.MatchedBy(func(org *entity.Organization) bool {
					return org.Settings["theme"] == "dark" && org.Settings["mfa_enabled"] == true
				})).Return(nil)

				response, err := uc.UpdateOrganization(ctx, orgID, request)

				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, "dark", response.Settings["theme"])
			},
		},
		{
			name:     "Positive_UpdateOrganization_NoChange",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupOrganizationTest()
				ctx := usecase.WithActorUserID(context.Background(), "owner-1")

				orgID := "org-123"
				existingOrg := &entity.Organization{
					ID:      orgID,
					Name:    "Acme Corp",
					OwnerID: "owner-1",
				}

				request := &model.UpdateOrganizationRequest{}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(nil)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(existingOrg, nil)
				deps.OrgRepo.On("Update", ctx, mock.Anything).Return(nil)

				response, err := uc.UpdateOrganization(ctx, orgID, request)

				assert.NoError(t, err)
				assert.Equal(t, "Acme Corp", response.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
