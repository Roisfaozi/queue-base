package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAuthUsecase_RefreshToken(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")
	oldRefreshToken, _ := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestRefreshSecret, 24*time.Hour)
	session := &model.Auth{ID: "session-1", UserID: user.ID, RefreshToken: oldRefreshToken}

	tests := []struct {
		name     string
		category string
		token    string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			token:    oldRefreshToken,
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
				deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, "session-1").Return(nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == user.ID && r.Action == "LOGOUT" && r.EntityID == "session-1"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			token:    oldRefreshToken,
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))
			},
			wantErr: errors.New("failed to get user roles"),
		},
		{
			name:     "Negative_InvalidToken",
			category: "negative",
			token:    "this.is.invalid",
			setup:    func(deps *testDependencies) {},
			wantErr:  usecase.ErrInvalidToken,
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			token:    oldRefreshToken,
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: gorm.ErrRecordNotFound,
		},
		{
			name:     "Negative_UserSuspended",
			category: "negative",
			token:    oldRefreshToken,
			setup: func(deps *testDependencies) {
				suspendedUser := *user
				suspendedUser.Status = entity.UserStatusBanned
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(session, nil)
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(&suspendedUser, nil)
			},
			wantErr: usecase.ErrAccountSuspended,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)

			resp, refresh, err := authService.RefreshToken(ctx, tt.token)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if !errors.Is(err, tt.wantErr) {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
				assert.Nil(t, resp)
				assert.Empty(t, refresh)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, refresh)
				assert.NotEqual(t, oldRefreshToken, refresh)
			}
		})
	}
}

func TestAuthUsecase_ValidateAccessToken(t *testing.T) {
	user, _ := createTestUser("password123")
	token, _ := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestAccessSecret, 15*time.Minute)
	expiredToken, _ := jwt.GenerateTestToken(user.ID, "session-1", TestRole, user.Username, "", TestAccessSecret, -1*time.Hour)

	tests := []struct {
		name     string
		category string
		token    string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			token:    token,
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(&model.Auth{ID: "session-1", UserID: user.ID, AccessToken: token}, nil)
			},
		},
		{
			name:     "Negative_Expired",
			category: "negative",
			token:    expiredToken,
			setup:    func(deps *testDependencies) {},
			wantErr:  usecase.ErrInvalidToken,
		},
		{
			name:     "Negative_TokenRevoked",
			category: "negative",
			token:    token,
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(nil, nil)
			},
			wantErr: usecase.ErrTokenRevoked,
		},
		{
			name:     "Negative_Mismatch",
			category: "negative",
			token:    token,
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, "session-1").Return(&model.Auth{ID: "session-1", UserID: user.ID, AccessToken: "different"}, nil)
			},
			wantErr: usecase.ErrTokenRevoked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			claims, err := authService.ValidateAccessToken(tt.token)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr))
				assert.Nil(t, claims)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, claims)
				assert.Equal(t, user.ID, claims.UserID)
			}
		})
	}
}

func TestAuthUsecase_RevokeToken(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		category string
		setup    func(*testDependencies)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("DeleteToken", mock.Anything, "u1", "s1").Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == "u1" && r.Action == "LOGOUT"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Positive_AuditError",
			category: "positive",
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("DeleteToken", mock.Anything, "u1", "s1").Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("err"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.RevokeToken(ctx, "u1", "s1")
			assert.NoError(t, err)
			deps.tokenRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUsecase_RevokeAllSessions(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		category string
		setup    func(*testDependencies)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("RevokeAllSessions", mock.Anything, "u1").Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == "u1" && r.Action == "REVOKE_ALL_SESSIONS"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Positive_AuditError",
			category: "positive",
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("RevokeAllSessions", mock.Anything, "u1").Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("err"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.RevokeAllSessions(ctx, "u1")
			assert.NoError(t, err)
			deps.tokenRepo.AssertExpectations(t)
		})
	}
}

func TestAuthUsecase_Verify_And_GetSessions(t *testing.T) {
	ctx := context.Background()

	t.Run("Verify_Success", func(t *testing.T) {
		authService, deps := setupTest(t)
		deps.tokenRepo.On("GetToken", mock.Anything, "u1", "s1").Return(&model.Auth{ID: "s1", UserID: "u1"}, nil)
		session, err := authService.Verify(ctx, "u1", "s1")
		assert.NoError(t, err)
		assert.Equal(t, "s1", session.ID)
	})

	t.Run("GetUserSessions_Success", func(t *testing.T) {
		authService, deps := setupTest(t)
		deps.tokenRepo.On("GetUserSessions", mock.Anything, "u1").Return([]*model.Auth{{ID: "s1", UserID: "u1"}}, nil)
		sessions, err := authService.GetUserSessions(ctx, "u1")
		assert.NoError(t, err)
		assert.Len(t, sessions, 1)
	})
}

func TestAuthUsecase_GenerateTokens(t *testing.T) {
	user, _ := createTestUser("pwd")

	tests := []struct {
		name     string
		category string
		method   string // access or refresh
		setup    func(*testDependencies)
		wantErr  bool
	}{
		{
			name:     "Positive_GenerateAccess_Success",
			category: "positive",
			method:   "access",
			setup: func(deps *testDependencies) {
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
			},
		},
		{
			name:     "Negative_GenerateAccess_EnforcerError",
			category: "negative",
			method:   "access",
			setup: func(deps *testDependencies) {
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("err"))
			},
			wantErr: true,
		},
		{
			name:     "Positive_GenerateRefresh_Success",
			category: "positive",
			method:   "refresh",
			setup: func(deps *testDependencies) {
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
			},
		},
		{
			name:     "Negative_GenerateRefresh_EnforcerError",
			category: "negative",
			method:   "refresh",
			setup: func(deps *testDependencies) {
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("err"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)

			var token string
			var err error
			if tt.method == "access" {
				token, err = authService.GenerateAccessToken(user)
			} else {
				token, err = authService.GenerateRefreshToken(user)
			}

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}
		})
	}
}
