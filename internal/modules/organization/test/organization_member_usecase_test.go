package test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/queue-base/internal/mocking"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	permissionMocks "github.com/Roisfaozi/queue-base/internal/modules/permission/test/mocks"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	userMocks "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/worker/tasks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	wsPkg "github.com/Roisfaozi/queue-base/pkg/ws"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type memberTestDeps struct {
	MemberRepo      *mocks.MockOrganizationMemberRepository
	OrgRepo         *mocks.MockOrganizationRepository
	InvitationRepo  *mocks.MockInvitationRepository
	UserRepo        *userMocks.MockUserRepository
	TaskDistributor *mocking.MockTaskDistributor
	Enforcer        *permissionMocks.MockIEnforcer
	Presence        *mocks.MockPresenceReader
	OrgReader       *mocks.MockIOrganizationReader
	TM              *mocking.MockWithTransactionManager
}

func setupMemberTest() (*memberTestDeps, usecase.OrganizationMemberUseCase) {
	mockEnforcer := new(permissionMocks.MockIEnforcer)
	deps := &memberTestDeps{
		MemberRepo:      new(mocks.MockOrganizationMemberRepository),
		OrgRepo:         new(mocks.MockOrganizationRepository),
		InvitationRepo:  new(mocks.MockInvitationRepository),
		UserRepo:        new(userMocks.MockUserRepository),
		TaskDistributor: new(mocking.MockTaskDistributor),
		Enforcer:        mockEnforcer,
		Presence:        new(mocks.MockPresenceReader),
		OrgReader:       new(mocks.MockIOrganizationReader),
		TM:              new(mocking.MockWithTransactionManager),
	}

	log := logrus.New()
	log.SetOutput(io.Discard)
	log.SetLevel(logrus.FatalLevel)

	uc := usecase.NewOrganizationMemberUseCase(
		log,
		deps.TM,
		deps.MemberRepo,
		deps.OrgRepo,
		deps.InvitationRepo,
		deps.UserRepo,
		deps.TaskDistributor,
		deps.Enforcer,
		deps.Presence,
		deps.OrgReader,
		"http://localhost:3000",
	)

	deps.OrgReader.On("InvalidateMembershipCache", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(nil)

	return deps, uc
}

func TestOrganizationMemberUseCase(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_InviteMember_SuccessExistingUser",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com", RoleID: "role:member"}
				org := &entity.Organization{ID: orgID, Name: "Org 1", OwnerID: actorID}
				user := &userEntity.User{ID: "user-1", Email: req.Email}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
				deps.MemberRepo.On("AddMember", ctx, mock.MatchedBy(func(m *entity.OrganizationMember) bool {
					return m.UserID == user.ID && m.OrganizationID == orgID && m.Status == entity.MemberStatusInvited
				})).Return(nil)
				deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
				deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
				deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
					return p.To == req.Email
				})).Return(nil)

				res, err := uc.InviteMember(ctx, orgID, req)
				require.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, user.ID, res.UserID)
				assert.Equal(t, entity.MemberStatusInvited, res.Status)
				deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, user.ID)
			},
		},
		{
			name:     "Positive_InviteMember_SuccessShadowUser",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "shadow@example.com", RoleID: "role:member"}
				org := &entity.Organization{ID: orgID, OwnerID: actorID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))
				deps.UserRepo.On("Create", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
					return u.Email == req.Email && u.Status == "invited"
				})).Return(nil)

				deps.MemberRepo.On("CheckMembership", ctx, orgID, mock.AnythingOfType("string")).Return(false, nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, orgID, mock.AnythingOfType("string")).Return("", nil)
				deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
				deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
				deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(nil)
				deps.TaskDistributor.On("DistributeTaskSendEmail", ctx, mock.Anything).Return(nil)

				res, err := uc.InviteMember(ctx, orgID, req)
				require.NoError(t, err)
				assert.NotNil(t, res)
				deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, res.UserID)
			},
		},
		{
			name:     "Negative_InviteMember_AlreadyMember",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "member@example.com"}
				org := &entity.Organization{ID: orgID, OwnerID: actorID}
				user := &userEntity.User{ID: "user-1", Email: req.Email}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(true, nil)

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrConflict)
			},
		},
		{
			name:     "Negative_InviteMember_OrgNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, nil)

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_InviteMember_UserRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}
				org := &entity.Organization{ID: orgID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("db error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_InviteMember_ShadowUserCreateError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "shadow@example.com"}
				org := &entity.Organization{ID: orgID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(nil, errors.New("user not found"))
				deps.UserRepo.On("Create", ctx, mock.Anything).Return(errors.New("create error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_InviteMember_MembershipCheckError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}
				org := &entity.Organization{ID: orgID}
				user := &userEntity.User{ID: "user-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, errors.New("db error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_InviteMember_GetMemberStatusError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}
				org := &entity.Organization{ID: orgID}
				user := &userEntity.User{ID: "user-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", errors.New("db error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_InviteMember_AddMemberError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}
				org := &entity.Organization{ID: orgID}
				user := &userEntity.User{ID: "user-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
				deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(errors.New("db error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_InviteMember_InvitationCleanupError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}
				org := &entity.Organization{ID: orgID}
				user := &userEntity.User{ID: "user-1", Email: req.Email}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
				deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
				deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(errors.New("db error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_InviteMember_InvitationCreateError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				req := &model.InviteMemberRequest{Email: "user@example.com"}
				org := &entity.Organization{ID: orgID}
				user := &userEntity.User{ID: "user-1", Email: req.Email}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.UserRepo.On("FindByEmail", ctx, req.Email).Return(user, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, user.ID).Return(false, nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, orgID, user.ID).Return("", nil)
				deps.MemberRepo.On("AddMember", ctx, mock.Anything).Return(nil)
				deps.InvitationRepo.On("DeleteByEmailAndOrg", ctx, req.Email, orgID).Return(nil)
				deps.InvitationRepo.On("Create", ctx, mock.Anything).Return(errors.New("db error"))

				_, err := uc.InviteMember(ctx, orgID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Positive_GetMembers_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				org := &entity.Organization{ID: orgID, OwnerID: actorID}
				members := []*entity.OrganizationMember{
					{UserID: "u1", RoleID: "r1"},
					{UserID: "u2", RoleID: "r2"},
				}

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("FindMembers", ctx, orgID).Return(members, nil)

				res, err := uc.GetMembers(ctx, orgID)
				require.NoError(t, err)
				assert.Len(t, res, 2)
			},
		},
		{
			name:     "Positive_UpdateMember_SuccessUpdateRole",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				userID := "user-1"
				req := &model.UpdateMemberRequest{RoleID: "new-role"}
				org := &entity.Organization{ID: orgID, OwnerID: actorID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)

				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)

				deps.MemberRepo.On("FindMembers", ctx, orgID).Return([]*entity.OrganizationMember{{UserID: userID, RoleID: "new-role"}}, nil)

				res, err := uc.UpdateMember(ctx, orgID, userID, req)
				require.NoError(t, err)
				assert.Equal(t, "new-role", res.RoleID)
				deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, userID)
			},
		},
		{
			name:     "Negative_UpdateMember_NotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrNotFound)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, nil)

				_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{})
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_UpdateMember_CheckMembershipError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, errors.New("db error"))

				_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{})
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_UpdateMember_UpdateRoleError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				req := &model.UpdateMemberRequest{RoleID: "new-role"}
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(errors.New("db error"))

				_, err := uc.UpdateMember(ctx, orgID, userID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_UpdateMember_UpdateStatusError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				req := &model.UpdateMemberRequest{Status: "inactive"}
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.MemberRepo.On("UpdateMemberStatus", ctx, orgID, userID, "inactive").Return(errors.New("db error"))

				_, err := uc.UpdateMember(ctx, orgID, userID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_UpdateMember_EnforcerAddGroupingError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				req := &model.UpdateMemberRequest{RoleID: "new-role"}
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

				_, err := uc.UpdateMember(ctx, orgID, userID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_UpdateMember_FindMembersError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				req := &model.UpdateMemberRequest{RoleID: "new-role"}
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.MemberRepo.On("UpdateMemberRole", ctx, orgID, userID, "new-role").Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)
				deps.MemberRepo.On("FindMembers", ctx, orgID).Return(nil, errors.New("db error"))

				_, err := uc.UpdateMember(ctx, orgID, userID, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Security_UpdateMember_NonAdminCannotUpdateOwner",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := usecase.WithActorUserID(context.Background(), "member-1")
				orgID := "org-1"
				userID := "owner-1"
				org := &entity.Organization{ID: orgID, OwnerID: userID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrForbidden)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, "member-1").Return(true, nil)
				deps.MemberRepo.On("GetMemberRole", ctx, orgID, "member-1").Return("role:user", nil)

				_, err := uc.UpdateMember(ctx, orgID, userID, &model.UpdateMemberRequest{Status: entity.MemberStatusSuspended})
				require.ErrorIs(t, err, exception.ErrForbidden)
			},
		},
		{
			name:     "Positive_RemoveMember_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				actorID := "owner-1"
				ctx := usecase.WithActorUserID(context.Background(), actorID)
				orgID := "org-1"
				userID := "user-1"
				org := &entity.Organization{ID: orgID, OwnerID: actorID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(nil)

				err := uc.RemoveMember(ctx, orgID, userID)
				require.NoError(t, err)
				deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, orgID, userID)
			},
		},
		{
			name:     "Security_RemoveMember_RemoveOwnerForbidden",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "owner"
				org := &entity.Organization{ID: orgID, OwnerID: userID}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrForbidden)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)

				err := uc.RemoveMember(ctx, orgID, userID)
				require.ErrorIs(t, err, exception.ErrForbidden)
			},
		},
		{
			name:     "Negative_RemoveMember_CheckMembershipError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(false, errors.New("db error"))

				err := uc.RemoveMember(ctx, orgID, userID)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_RemoveMember_OrgFindError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.OrgRepo.On("FindByID", ctx, orgID).Return(nil, errors.New("db error"))

				err := uc.RemoveMember(ctx, orgID, userID)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_RemoveMember_RemoveMemberError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				userID := "user-1"
				org := &entity.Organization{ID: orgID, OwnerID: "other"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.MemberRepo.On("CheckMembership", ctx, orgID, userID).Return(true, nil)
				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("RemoveFilteredGroupingPolicy", mock.Anything, mock.Anything).Return(true, nil)
				deps.MemberRepo.On("RemoveMember", ctx, orgID, userID).Return(errors.New("db error"))

				err := uc.RemoveMember(ctx, orgID, userID)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Security_RemoveMember_ActorNotManagerForbidden",
			category: "security",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := usecase.WithActorUserID(context.Background(), "member-1")
				orgID := "org-1"
				userID := "user-1"
				org := &entity.Organization{ID: orgID, OwnerID: "owner-1"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrForbidden)

				deps.OrgRepo.On("FindByID", ctx, orgID).Return(org, nil)
				deps.MemberRepo.On("CheckMembership", ctx, orgID, "member-1").Return(true, nil)
				deps.MemberRepo.On("GetMemberRole", ctx, orgID, "member-1").Return("role:user", nil)

				err := uc.RemoveMember(ctx, orgID, userID)
				require.ErrorIs(t, err, exception.ErrForbidden)
			},
		},
		{
			name:     "Positive_GetPresence_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				orgID := "org-1"
				presenceUsers := []wsPkg.PresenceUser{{UserID: "u1"}}

				deps.Presence.On("GetOnlineUsers", ctx, orgID).Return(presenceUsers, nil)

				res, err := uc.GetPresence(ctx, orgID)
				require.NoError(t, err)
				assert.Len(t, res, 1)
			},
		},
		{
			name:     "Positive_AcceptInvitation_SuccessActivateNewUser",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass", Name: "Name"}
				inv := &entity.InvitationToken{ID: "inv-1", Email: "new@example.com", OrganizationID: "org-1", Role: "role:member", ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
				user := &userEntity.User{ID: "user-1", Email: "new@example.com", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.MatchedBy(func(u *userEntity.User) bool {
					return u.Status == "active" && u.Name == req.Name
				})).Return(nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
				deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)
				deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(nil)

				err := uc.AcceptInvitation(ctx, req)
				require.NoError(t, err)
				deps.OrgReader.AssertCalled(t, "InvalidateMembershipCache", mock.Anything, inv.OrganizationID, user.ID)
			},
		},
		{
			name:     "Negative_AcceptInvitation_InvalidToken",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "invalid"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrUnauthorized)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(nil, nil)

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrUnauthorized)
			},
		},
		{
			name:     "Negative_AcceptInvitation_Expired",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "expired"}
				inv := &entity.InvitationToken{
					ID:        "inv-1",
					ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
				}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrUnauthorized)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrUnauthorized)
			},
		},
		{
			name:     "Negative_AcceptInvitation_UserRepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(nil, errors.New("db error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_AcceptInvitation_UserNotFound",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrNotFound)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(nil, nil)

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_AcceptInvitation_ShadowUserMissingPassword",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: ""}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(ctx context.Context, fn func(context.Context) error) error {
					return fn(ctx)
				})

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_AcceptInvitation_UserUpdateError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.Anything).Return(errors.New("db error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_AcceptInvitation_GetMemberStatusError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return("", errors.New("db error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_AcceptInvitation_UpdateMemberStatusError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
				deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(errors.New("db error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_AcceptInvitation_AddMemberNoStatusError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return("", nil)
				deps.MemberRepo.On("AddMember", ctx, mock.MatchedBy(func(m *entity.OrganizationMember) bool {
					return m.Status == entity.MemberStatusActive
				})).Return(errors.New("db error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_AcceptInvitation_EnforcerError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1", Role: "role:member",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
				deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(false, errors.New("casbin error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
		{
			name:     "Negative_AcceptInvitation_InvitationDeleteError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupMemberTest()
				ctx := context.Background()
				req := &model.AcceptInvitationRequest{Token: "token", Password: "pass"}
				inv := &entity.InvitationToken{
					ID: "inv-1", Email: "user@example.com", OrganizationID: "org-1", Role: "role:member",
					ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli(),
				}
				user := &userEntity.User{ID: "user-1", Status: "invited"}

				deps.TM.On("WithinTransaction", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					fn := args.Get(1).(func(context.Context) error)
					_ = fn(ctx)
				}).Return(exception.ErrInternalServer)

				deps.InvitationRepo.On("FindByToken", ctx, req.Token).Return(inv, nil)
				deps.UserRepo.On("FindByEmail", ctx, inv.Email).Return(user, nil)
				deps.UserRepo.On("Update", ctx, mock.Anything).Return(nil)
				deps.MemberRepo.On("GetMemberStatus", ctx, inv.OrganizationID, user.ID).Return(entity.MemberStatusInvited, nil)
				deps.MemberRepo.On("UpdateMemberStatus", ctx, inv.OrganizationID, user.ID, entity.MemberStatusActive).Return(nil)
				deps.Enforcer.On("WithContext", mock.Anything).Return(deps.Enforcer)
				deps.Enforcer.On("AddGroupingPolicy", mock.Anything).Return(true, nil)
				deps.InvitationRepo.On("Delete", ctx, inv.ID).Return(errors.New("db error"))

				err := uc.AcceptInvitation(ctx, req)
				require.ErrorIs(t, err, exception.ErrInternalServer)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
