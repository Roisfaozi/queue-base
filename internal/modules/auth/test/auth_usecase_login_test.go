package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestAuthUsecase_Login(t *testing.T) {
	ctx := context.Background()
	user, password := createTestUser("password123")

	tests := []struct {
		name     string
		category string
		req      model.LoginRequest
		setup    func(*testDependencies)
		wantErr  error
		wantNil  bool
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			req:      model.LoginRequest{Username: user.Username, Password: password, IPAddress: "127.0.0.1", UserAgent: "TestAgent"},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == user.ID && r.Action == "LOGIN" && r.Entity == "Auth" && r.IPAddress == "127.0.0.1"
				}), mock.Anything).Return(nil)
				deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			req:      model.LoginRequest{Username: "nonexistent", Password: "password123"},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, "nonexistent").Return(false, time.Duration(0), nil)
				simulateWithinTransaction(deps, usecase.ErrInvalidCredentials)
				deps.userRepo.On("FindByUsername", mock.Anything, "nonexistent").Return(nil, gorm.ErrRecordNotFound)
			},
			wantErr: usecase.ErrInvalidCredentials,
			wantNil: true,
		},
		{
			name:     "Negative_InvalidPassword",
			category: "negative",
			req:      model.LoginRequest{Username: user.Username, Password: "wrong-password"},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil)
				simulateWithinTransaction(deps, usecase.ErrInvalidCredentials)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
			},
			wantErr: usecase.ErrInvalidCredentials,
			wantNil: true,
		},
		{
			name:     "Negative_UserSuspended",
			category: "negative",
			req:      model.LoginRequest{Username: user.Username, Password: password},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				simulateWithinTransaction(deps, usecase.ErrAccountSuspended)
				suspendedUser := *user
				suspendedUser.Status = entity.UserStatusSuspended
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(&suspendedUser, nil)
			},
			wantErr: usecase.ErrAccountSuspended,
			wantNil: true,
		},
		{
			name:     "Negative_StoreTokenError",
			category: "negative",
			req:      model.LoginRequest{Username: user.Username, Password: password},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(errors.New("redis is down"))
			},
			wantErr: errors.New("failed to store session"), // Matches error.Contains
			wantNil: true,
		},
		{
			name:     "Negative_EnforcerError",
			category: "negative",
			req:      model.LoginRequest{Username: user.Username, Password: password},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, errors.New("casbin error"))
			},
			wantErr: errors.New("failed to get user roles"),
			wantNil: true,
		},
		{
			name:     "Positive_AuditError_ShouldNotBlock",
			category: "positive",
			req:      model.LoginRequest{Username: user.Username, Password: password},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{TestRole}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))
				deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()
			},
		},
		{
			name:     "Positive_NoRoles",
			category: "positive",
			req:      model.LoginRequest{Username: user.Username, Password: password},
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.orgRepo.On("FindUserOrganizations", mock.Anything, user.ID).Return([]*orgEntity.Organization{}, nil)
				deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == user.ID && r.Action == "LOGIN"
				}), mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)

			resp, refresh, err := authService.Login(ctx, tt.req)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if !errors.Is(err, tt.wantErr) {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
				assert.Empty(t, refresh)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, refresh)
				assert.NotNil(t, resp)
				if resp != nil {
					assert.NotEmpty(t, resp.AccessToken)
					assert.Equal(t, "Bearer", resp.TokenType)
					assert.Equal(t, user.ID, resp.User.ID)
					assert.Equal(t, user.Username, resp.User.Username)
				}
			}

			if tt.wantNil {
				assert.Nil(t, resp)
			}
			deps.tm.AssertExpectations(t)
			deps.userRepo.AssertExpectations(t)
			deps.authz.AssertExpectations(t)
			deps.tokenRepo.AssertExpectations(t)
			deps.publisher.AssertExpectations(t)
			deps.taskDistributor.AssertExpectations(t)
		})
	}
}

func TestAuthUsecase_Security_BruteForceProtection(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")
	wrongPassword := "wrongpass"
	maxAttempts := 1
	lockoutDuration := 15 * time.Minute

	t.Run("Negative_AccountLocked", func(t *testing.T) {
		_, baseDeps := setupTest(t)
		authService := usecase.NewAuthUsecase(
			maxAttempts, lockoutDuration, baseDeps.jwtManager, baseDeps.tokenRepo, baseDeps.userRepo,
			baseDeps.orgRepo, baseDeps.tm, baseDeps.log, baseDeps.publisher, baseDeps.authz,
			baseDeps.taskDistributor, baseDeps.ticketManager, nil,
		)

		baseDeps.tm.On("WithinTransaction", mock.Anything, mock.Anything).Return(func(c context.Context, fn func(context.Context) error) error {
			return fn(c)
		})
		baseDeps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
		baseDeps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)
		baseDeps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(1, nil)
		baseDeps.tokenRepo.On("LockAccount", mock.Anything, user.Username, lockoutDuration).Return(nil)
		baseDeps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		_, _, err := authService.Login(ctx, model.LoginRequest{Username: user.Username, Password: wrongPassword})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "too many failed attempts")
	})
}
