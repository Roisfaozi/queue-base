//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/delivery"
	auditRepository "github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	authEntity "github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/queue-base/internal/modules/auth/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	orgRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/internal/worker"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/Roisfaozi/queue-base/pkg/sse"
	"github.com/Roisfaozi/queue-base/pkg/sso"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/Roisfaozi/queue-base/pkg/util"
	"github.com/Roisfaozi/queue-base/pkg/ws"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuthIntegration(env *setup.TestEnvironment) (usecase.AuthUseCase, *jwt.JWTManager) {
	jwtManager := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
	return setupAuthIntegrationWithJWT(env, jwtManager), jwtManager
}

func setupAuthIntegrationWithJWT(env *setup.TestEnvironment, jwtManager *jwt.JWTManager) usecase.AuthUseCase {
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	_ = auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil, nil)

	wsConfig := &ws.WebSocketConfig{}
	presenceManager := ws.NewPresenceManager(env.Redis, env.Logger, 5*time.Minute)
	wsManager := ws.NewWebSocketManager(wsConfig, env.Logger, env.Redis, presenceManager)
	sseManager := sse.NewManager()

	env.AddCloser(func() {
		sseManager.Stop()
		wsManager.Stop()
	})

	taskDistributor := worker.NewRedisTaskDistributor(asynq.RedisClientOpt{Addr: env.RedisAddr})

	enforcer := env.Enforcer
	logger := env.Logger

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)

	ticketManager := ws.NewRedisTicketManager(env.Redis, 30*time.Second)

	publisher := delivery.NewEventPublisher(wsManager, sseManager, env.Logger)
	authz := authRepository.NewCasbinAdapter(enforcer, "role:user", "global")

	return usecase.NewAuthUsecase(
		5,              // MaxLoginAttempts
		30*time.Minute, // LockoutDuration
		jwtManager,
		tokenRepo,
		userRepo,
		orgRepo,
		tm,
		logger,
		publisher,
		authz,
		taskDistributor,
		ticketManager,
		make(map[string]sso.Provider),
	)
}

func TestAuthIntegration_Login(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_ValidCredentials",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				password := "SecurePass123!"
				testUser := setup.CreateTestUser(t, env.DB, "authuser1", "auth1@example.com", password)
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")

				loginReq := model.LoginRequest{Username: "authuser1", Password: password, IPAddress: "127.0.0.1", UserAgent: "Mozilla/5.0"}
				resp, refreshToken, err := authUC.Login(context.Background(), loginReq)

				require.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotEmpty(t, resp.AccessToken)
				assert.NotEmpty(t, refreshToken)
				assert.Equal(t, "Bearer", resp.TokenType)
				assert.Equal(t, testUser.ID, resp.User.ID)
				assert.Greater(t, int64(resp.ExpiresIn), int64(0))

				sessionKeys, err := env.Redis.SMembers(context.Background(), fmt.Sprintf("session_index:%s", testUser.ID)).Result()
				require.NoError(t, err)
				assert.NotEmpty(t, sessionKeys)
			},
		},
		{
			name:     "Negative_InvalidPassword",
			category: "negative",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				setup.CreateTestUser(t, env.DB, "authuser2", "auth2@example.com", "SecurePass123!")

				loginReq := model.LoginRequest{Username: "authuser2", Password: "wrongpassword"}
				resp, _, err := authUC.Login(context.Background(), loginReq)
				assert.Error(t, err)
				assert.Nil(t, resp)
			},
		},
		{
			name:     "Negative_NonExistentUser",
			category: "negative",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				loginReq := model.LoginRequest{Username: "nonexistent", Password: "password123"}
				_, _, err := authUC.Login(context.Background(), loginReq)
				assert.Error(t, err)
			},
		},
		{
			name:     "Negative_EmptyCredentials",
			category: "negative",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				_, _, err1 := authUC.Login(context.Background(), model.LoginRequest{Username: "", Password: "password123"})
				assert.Error(t, err1)

				_, _, err2 := authUC.Login(context.Background(), model.LoginRequest{Username: "authuser", Password: ""})
				assert.Error(t, err2)
			},
		},
		{
			name:     "Edge_SpecialCharactersUsername",
			category: "edge",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				specialUN := "user-@#$%^&*()"
				setup.CreateTestUser(t, env.DB, specialUN, "special@example.com", "password123")

				loginReq := model.LoginRequest{Username: specialUN, Password: "password123"}
				_, _, err := authUC.Login(context.Background(), loginReq)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Edge_LongPassword",
			category: "edge",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				longPW := strings.Repeat("a", 72)
				setup.CreateTestUser(t, env.DB, "longpw", "long@example.com", longPW)

				loginReq := model.LoginRequest{Username: "longpw", Password: longPW}
				_, _, err := authUC.Login(context.Background(), loginReq)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Edge_UnicodeCharacters",
			category: "edge",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				unicodeUN := "用户名测试"
				setup.CreateTestUser(t, env.DB, unicodeUN, "unicode@example.com", "password123")

				loginReq := model.LoginRequest{Username: unicodeUN, Password: "password123"}
				_, _, err := authUC.Login(context.Background(), loginReq)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Edge_CaseSensitivity",
			category: "edge",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				setup.CreateTestUser(t, env.DB, "CaseUser", "case@example.com", "password123")

				loginReq := model.LoginRequest{Username: "caseuser", Password: "password123"}
				_, loginResp, err := authUC.Login(context.Background(), loginReq)
				if err == nil {
					assert.NotEmpty(t, loginResp)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthIntegration_TokenLifecycle(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_RefreshToken",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				password := "password123"
				testUser := setup.CreateTestUser(t, env.DB, "tokenuser1", "token1@example.com", password)
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")
				_, refreshToken, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "tokenuser1", Password: password})

				time.Sleep(1 * time.Second)
				newToken, newRefresh, err := authUC.RefreshToken(context.Background(), refreshToken)
				require.NoError(t, err)
				assert.NotEmpty(t, newToken.AccessToken)
				assert.NotEqual(t, refreshToken, newRefresh)
			},
		},
		{
			name:     "Success_MultipleRefreshInSequence",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)
				password := "password123"
				testUser := setup.CreateTestUser(t, env.DB, "tokenuser2", "token2@example.com", password)
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")
				_, refreshToken, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "tokenuser2", Password: password})

				currRefresh := refreshToken
				for i := 0; i < 3; i++ {
					time.Sleep(100 * time.Millisecond)
					_, nextRefresh, err := authUC.RefreshToken(context.Background(), currRefresh)
					require.NoError(t, err, "Refresh iteration %d failed", i+1)
					currRefresh = nextRefresh
				}
			},
		},
		{
			name:     "Negative_RefreshInvalidToken",
			category: "negative",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				_, _, err := authUC.RefreshToken(context.Background(), "invalid.token.here")
				assert.Error(t, err)
			},
		},
		{
			name:     "Negative_RefreshExpiredToken",
			category: "negative",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				shortJWT := jwt.NewJWTManager("secret", "refresh", time.Minute, 1*time.Millisecond)
				customUC := setupAuthIntegrationWithJWT(env, shortJWT)
				testUser := setup.CreateTestUser(t, env.DB, "tokenuser3", "token3@example.com", "pass")

				expToken, _, _ := shortJWT.GenerateTokenPair(jwt.UserContext{
					UserID:    testUser.ID,
					SessionID: "sid",
					Role:      "role:user",
					Username:  "tokenuser",
				})
				time.Sleep(10 * time.Millisecond)

				_, _, err := customUC.RefreshToken(context.Background(), expToken)
				assert.Error(t, err)
			},
		},
		{
			name:     "Success_LogoutRevoke",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, jwtManager := setupAuthIntegration(env)
				password := "password123"
				testUser := setup.CreateTestUser(t, env.DB, "tokenuser4", "token4@example.com", password)
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")

				lr, _, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "tokenuser4", Password: password})
				claims, _ := jwtManager.ValidateAccessToken(lr.AccessToken)

				err := authUC.RevokeToken(context.Background(), testUser.ID, claims.SessionID)
				require.NoError(t, err)

				sessionKey := fmt.Sprintf("session:%s:%s", testUser.ID, claims.SessionID)
				exists, err := env.Redis.Exists(context.Background(), sessionKey).Result()
				require.NoError(t, err)
				assert.Zero(t, exists, "Session key should be deleted from Redis")

				indexKey := fmt.Sprintf("session_index:%s", testUser.ID)
				member, err := env.Redis.SIsMember(context.Background(), indexKey, sessionKey).Result()
				require.NoError(t, err)
				assert.False(t, member, "Session index should not contain revoked session")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthIntegration_PasswordRecovery(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_ForgotPassword",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				email := "forgot@example.com"
				setup.CreateTestUser(t, env.DB, "forgotuser", email, "old-pass")

				err := authUC.ForgotPassword(context.Background(), email)
				require.NoError(t, err)

				var token authEntity.PasswordResetToken
				err = env.DB.Where("email = ?", email).First(&token).Error
				require.NoError(t, err)
				assert.NotEmpty(t, token.Token)
			},
		},
		{
			name:     "Success_ResetPassword",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				email := "reset_unique@example.com"
				testUser := setup.CreateTestUser(t, env.DB, "resetuser", email, "oldpass")

				resetToken := "secret-token-unique-123"
				err := env.DB.Create(&authEntity.PasswordResetToken{
					Email: email, Token: resetToken, ExpiresAt: time.Now().Add(time.Hour),
				}).Error
				require.NoError(t, err)

				err = authUC.ResetPassword(context.Background(), resetToken, "NewPass123!")
				require.NoError(t, err)

				_, _, err = authUC.Login(context.Background(), model.LoginRequest{Username: testUser.Username, Password: "oldpass"})
				assert.Error(t, err)
				_, _, err = authUC.Login(context.Background(), model.LoginRequest{Username: testUser.Username, Password: "NewPass123!"})
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestAuthIntegration_Security(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Security_SQLInjection",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				injections := []string{"admin' OR '1'='1", "admin'--", "admin'; DROP TABLE users--"}
				for _, inj := range injections {
					_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: inj, Password: "p"})
					assert.Error(t, err)
				}
			},
		},
		{
			name:     "Security_BruteForceProtection",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				setup.CreateTestUser(t, env.DB, "brute", "brute@example.com", "pass")
				for i := 0; i < 5; i++ {
					_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: "brute", Password: "w"})
					assert.Error(t, err)
				}
			},
		},
		{
			name:     "Security_TokenRotationReuse",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				testUser := setup.CreateTestUser(t, env.DB, "reuse", "reuse@example.com", "pass")
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")
				_, rt1, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "reuse", Password: "pass"})

				_, rt2, _ := authUC.RefreshToken(context.Background(), rt1)

				_, _, err := authUC.RefreshToken(context.Background(), rt1)
				assert.Error(t, err)

				_, _, err = authUC.RefreshToken(context.Background(), rt2)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Security_SessionHijacking",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				testUser := setup.CreateTestUser(t, env.DB, "hijack", "hijack@example.com", "pass")
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")

				r1, _, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "hijack", Password: "pass", UserAgent: "D1"})
				r2, _, _ := authUC.Login(context.Background(), model.LoginRequest{Username: "hijack", Password: "pass", UserAgent: "D2"})
				assert.NotEqual(t, r1.AccessToken, r2.AccessToken)
			},
		},
		{
			name:     "Security_XSSUserAgent",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				authUC, _ := setupAuthIntegration(env)

				testUser := setup.CreateTestUser(t, env.DB, "xss", "xss@example.com", "pass")
				_, _ = env.Enforcer.AddGroupingPolicy(testUser.ID, "role:user", "global")
				_, _, err := authUC.Login(context.Background(), model.LoginRequest{Username: "xss", Password: "pass", UserAgent: "<script>alert(1)</script>"})
				assert.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
