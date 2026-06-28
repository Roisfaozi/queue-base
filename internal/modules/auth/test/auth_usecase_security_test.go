package test

import (
	"context"
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	mocking "github.com/Roisfaozi/queue-base/internal/mocking"
	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	authEntity "github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/Roisfaozi/queue-base/pkg/sso"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	mock_auth "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	mock_org "github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	mock_user "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
)

func TestAuthUsecase_Security_LoginConcurrent(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "MultipleFailedAttempts_RaceCondition",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")
				wrongPassword := "wrongpassword"
				numConcurrent := 10

				var attemptCounter int32

				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

				deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					}).Return(usecase.ErrInvalidCredentials)

				deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).
					Run(func(args mock.Arguments) {
						atomic.AddInt32(&attemptCounter, 1)
					}).Return(3, nil)

				var wg sync.WaitGroup
				errChan := make(chan error, numConcurrent)

				for i := 0; i < numConcurrent; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						_, _, err := authService.Login(context.Background(), model.LoginRequest{
							Username: user.Username,
							Password: wrongPassword,
						})
						errChan <- err
					}()
				}
				wg.Wait()
				close(errChan)

				for err := range errChan {
					assert.ErrorIs(t, err, usecase.ErrInvalidCredentials)
				}
				assert.Equal(t, int32(numConcurrent), atomic.LoadInt32(&attemptCounter))
			},
		},
		{
			name:     "AccountLockAtThreshold",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")

				var lockCalled int32

				deps.tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil).Once()
				deps.userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil).Once()

				deps.tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					}).Return(usecase.ErrAccountLocked).Once()

				deps.tokenRepo.On("IncrementLoginAttempts", mock.Anything, user.Username).Return(5, nil).Once()
				deps.tokenRepo.On("LockAccount", mock.Anything, user.Username, mock.Anything).
					Run(func(args mock.Arguments) {
						atomic.AddInt32(&lockCalled, 1)
					}).Return(nil).Once()

				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
					return req.Action == "ACCOUNT_LOCKED"
				}), mock.Anything).Return(nil).Once()

				_, _, err := authService.Login(context.Background(), model.LoginRequest{
					Username: user.Username,
					Password: "wrongpassword",
				})
				assert.ErrorIs(t, err, usecase.ErrAccountLocked)
				assert.Equal(t, int32(1), atomic.LoadInt32(&lockCalled))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthUsecase_Security_TokenReplay(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "AccessToken_ReplayAfterRefresh",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")
				sessionID := "session-1"

				oldAccessToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-access-secret", 15*time.Minute)
				assert.NoError(t, err)

				newSession := &model.Auth{
					ID:           sessionID,
					UserID:       user.ID,
					AccessToken:  "new-access-token-after-refresh",
					RefreshToken: "new-refresh-token",
				}
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(newSession, nil)

				claims, err := authService.ValidateAccessToken(oldAccessToken)
				assert.ErrorIs(t, err, usecase.ErrTokenRevoked)
				assert.Nil(t, claims)
			},
		},
		{
			name:     "RefreshToken_ReplayAfterRefresh",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")
				sessionID := "session-1"

				oldRefreshToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", 24*time.Hour)
				assert.NoError(t, err)

				newSession := &model.Auth{
					ID:           sessionID,
					UserID:       user.ID,
					AccessToken:  "new-access-token",
					RefreshToken: "new-refresh-token-after-refresh",
				}
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(newSession, nil)

				claims, err := authService.ValidateRefreshToken(oldRefreshToken)
				assert.ErrorIs(t, err, usecase.ErrTokenRevoked)
				assert.Nil(t, claims)
			},
		},
		{
			name:     "VerifyEmail_SameTokenTwice",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")
				user.EmailVerifiedAt = nil
				verificationToken := "verification-token-123"

				tokenEntity := &authEntity.EmailVerificationToken{
					Email:     user.Email,
					Token:     verificationToken,
					ExpiresAt: time.Now().Add(15 * time.Minute).UnixMilli(),
				}

				deps.tokenRepo.On("FindVerificationToken", mock.Anything, verificationToken).Return(tokenEntity, nil).Once()
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil).Once()
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil).Once()
				deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil).Once()
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
					return req.Action == "EMAIL_VERIFIED" && req.Entity == "User"
				}), mock.Anything).Return(nil).Once()

				err := authService.VerifyEmail(context.Background(), verificationToken)
				assert.NoError(t, err)

				deps.tokenRepo.On("FindVerificationToken", mock.Anything, verificationToken).Return(nil, errors.New("not found")).Once()
				err = authService.VerifyEmail(context.Background(), verificationToken)
				assert.ErrorIs(t, err, usecase.ErrInvalidVerificationToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthUsecase_Security_ConcurrentTokenValidation(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "ConcurrentTokenValidation",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")
				sessionID := "concurrent-session"
				numConcurrent := 20

				accessToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-access-secret", 15*time.Minute)
				assert.NoError(t, err)

				validSession := &model.Auth{
					ID:           sessionID,
					UserID:       user.ID,
					AccessToken:  accessToken,
					RefreshToken: "refresh-token",
				}
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(validSession, nil)

				var wg sync.WaitGroup
				var successCount int32
				var failCount int32

				for i := 0; i < numConcurrent; i++ {
					wg.Add(1)
					go func() {
						defer wg.Done()
						claims, err := authService.ValidateAccessToken(accessToken)
						if err == nil && claims != nil {
							atomic.AddInt32(&successCount, 1)
						} else {
							atomic.AddInt32(&failCount, 1)
						}
					}()
				}
				wg.Wait()

				assert.Equal(t, int32(numConcurrent), successCount)
				assert.Equal(t, int32(0), failCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthUsecase_Security_SessionCleanup(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "RefreshToken_SessionCleanupFailure_StillSucceeds",
			category: "security",
			run: func(t *testing.T) {
				authService, deps := setupTest(t)
				user, _ := createTestUser("password123")
				sessionID := "session-to-refresh"

				refreshToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", 24*time.Hour)
				assert.NoError(t, err)

				savedSession := &model.Auth{
					ID:           sessionID,
					UserID:       user.ID,
					RefreshToken: refreshToken,
					AccessToken:  "old-access-token",
				}
				deps.tokenRepo.On("GetToken", mock.Anything, user.ID, sessionID).Return(savedSession, nil)
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, user.ID, "").Return([]string{"role:user"}, nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
					return req.Action == "LOGOUT" && req.Entity == "Auth"
				}), mock.Anything).Return(nil)
				deps.tokenRepo.On("DeleteToken", mock.Anything, user.ID, sessionID).Return(errors.New("redis connection lost"))
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

				tokenResp, newRefreshToken, err := authService.RefreshToken(context.Background(), refreshToken)
				assert.NoError(t, err)
				assert.NotNil(t, tokenResp)
				assert.NotEmpty(t, newRefreshToken)
			},
		},
		{
			name:     "RefreshToken_ExpiredOrphanedSession",
			category: "security",
			run: func(t *testing.T) {
				authService, _ := setupTest(t)
				user, _ := createTestUser("password123")
				sessionID := "orphaned-session"

				expiredToken, err := jwt.GenerateTestToken(user.ID, sessionID, "role:user", user.Username, "", "test-refresh-secret", -1*time.Hour)
				assert.NoError(t, err)

				_, _, err = authService.RefreshToken(context.Background(), expiredToken)
				assert.ErrorIs(t, err, usecase.ErrInvalidToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthUsecase_Security_NilEnforcer(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "NilEnforcer",
			category: "security",
			run: func(t *testing.T) {
				jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
				tokenRepo := new(mock_auth.MockTokenRepository)
				userRepo := new(mock_user.MockUserRepository)
				orgRepo := new(mock_org.MockOrganizationRepository)
				tm := new(mocking.MockWithTransactionManager)
				taskDistributor := new(mocking.MockTaskDistributor)
				log := logrus.New()
				log.SetOutput(io.Discard)

				mockPublisher := new(mock_auth.MockNotificationPublisher)
				mockPublisher.On("PublishUserLoggedIn", mock.Anything, mock.Anything, mock.Anything).Return()

				authService := usecase.NewAuthUsecase(
					5,
					30*time.Minute,
					jwtManager,
					tokenRepo,
					userRepo,
					orgRepo,
					tm,
					log,
					mockPublisher,
					nil, // nil enforcer
					taskDistributor,
					nil, // nil ticket manager
					make(map[string]sso.Provider),
				)

				user, _ := createTestUser("password123")
				tokenRepo.On("IsAccountLocked", mock.Anything, user.Username).Return(false, time.Duration(0), nil)
				tokenRepo.On("ResetLoginAttempts", mock.Anything, user.Username).Return(nil)
				tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				userRepo.On("FindByUsername", mock.Anything, user.Username).Return(user, nil)

				tm.On("WithinTransaction", mock.Anything, mock.AnythingOfType("func(context.Context) error")).
					Run(func(args mock.Arguments) {
						fn := args.Get(1).(func(context.Context) error)
						_ = fn(context.Background())
					}).Return(nil)

				taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
					return req.Action == "LOGIN" && req.Entity == "Auth"
				})).Return(nil)
				orgRepo.On("FindByOwnerID", mock.Anything, mock.AnythingOfType("string")).Return([]orgEntity.Organization{}, nil)
				orgRepo.On("FindUserOrganizations", mock.Anything, mock.AnythingOfType("string")).Return([]*orgEntity.Organization{}, nil)
				mockPublisher.On("PublishUserLoggedIn", mock.Anything, mock.AnythingOfType("model.UserInfo"), mock.AnythingOfType("[]string")).Return(nil)

				tokenResp, _, err := authService.Login(context.Background(), model.LoginRequest{
					Username: user.Username,
					Password: "password123",
				})
				assert.NoError(t, err)
				assert.NotNil(t, tokenResp)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
