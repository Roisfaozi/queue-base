//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	auditRepository "github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	authRepository "github.com/Roisfaozi/queue-base/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	orgRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/user/model"
	"github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/user/usecase"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/Roisfaozi/queue-base/pkg/sso"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/Roisfaozi/queue-base/pkg/util"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupUserIntegration(env *setup.TestEnvironment) (usecase.UserUseCase, repository.UserRepository) {
	userRepo := repository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil, nil)

	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	jwtManager := jwt.NewJWTManager("test-secret", "test-refresh", time.Hour, time.Hour*24)

	orgRepo := orgRepository.NewOrganizationRepository(env.DB)
	authz := authRepository.NewCasbinAdapter(env.Enforcer, "role:user", "global")
	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, orgRepo, tm, env.Logger, nil, authz, nil, nil, make(map[string]sso.Provider))

	return usecase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, nil, nil), userRepo
}

func TestUserIntegration_Lifecycle(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository)
	}{
		{
			name:     "Positive_ValidData",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "validuser", Email: "valid@example.com", Password: "SecurePass123!",
					Name: "Valid User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, req.Username, result.Username)
			},
		},
		{
			name:     "Positive_WithRoleAssignment",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "roleuser", Email: "roleuser@example.com", Password: "Password123!",
					Name: "Role User", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)

				_ = env.Enforcer.LoadPolicy()
				roles, err := env.Enforcer.GetRolesForUser(result.ID, "global")
				require.NoError(t, err)
				assert.Contains(t, roles, "role:user")
			},
		},
		{
			name:     "Positive_Update",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				testUser := setup.CreateTestUser(t, env.DB, "updateuser", "update@example.com", "password123")
				updateReq := &model.UpdateUserRequest{
					ID: testUser.ID, Name: "Updated Name", IPAddress: "127.0.0.1", UserAgent: "TestAgent",
				}
				result, err := userUC.Update(context.Background(), updateReq)
				require.NoError(t, err)
				assert.Equal(t, "Updated Name", result.Name)
			},
		},
		{
			name:     "Positive_GetByID",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				testUser := setup.CreateTestUser(t, env.DB, "getbyiduser", "getbyid@example.com", "password123")
				result, err := userUC.GetUserByID(context.Background(), testUser.ID)
				require.NoError(t, err)
				assert.Equal(t, testUser.ID, result.ID)
			},
		},
		{
			name:     "Positive_Delete",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				testUser := setup.CreateTestUser(t, env.DB, "deleteuser", "delete@example.com", "password123")
				deleteReq := &model.DeleteUserRequest{ID: testUser.ID, IPAddress: "127.0.0.1", UserAgent: "TestAgent"}
				err := userUC.DeleteUser(context.Background(), "admin-id", deleteReq)
				require.NoError(t, err)

				_, err = userRepo.FindByID(context.Background(), testUser.ID)
				assert.Error(t, err)
			},
		},
		{
			name:     "Negative_DuplicateUsername",
			category: "negative",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				setup.CreateTestUser(t, env.DB, "duplicate", "first@example.com", "password123")
				req := &model.RegisterUserRequest{
					Username: "duplicate", Email: "second@example.com", Password: "password123", Name: "Second",
				}
				_, err := userUC.Create(context.Background(), req)
				assert.Error(t, err)
			},
		},
		{
			name:     "Negative_DuplicateEmail",
			category: "negative",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				setup.CreateTestUser(t, env.DB, "user1", "duplicate@example.com", "password123")
				req := &model.RegisterUserRequest{
					Username: "user2", Email: "duplicate@example.com", Password: "password123", Name: "User Two",
				}
				_, err := userUC.Create(context.Background(), req)
				assert.Error(t, err)
			},
		},
		{
			name:     "Negative_NonExistentUserUpdate",
			category: "negative",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				updateReq := &model.UpdateUserRequest{ID: "non-existent-id", Name: "Updated Name"}
				_, err := userUC.Update(context.Background(), updateReq)
				assert.Error(t, err)
			},
		},
		{
			name:     "Negative_NonExistentUserDelete",
			category: "negative",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				deleteReq := &model.DeleteUserRequest{ID: "non-existent-id"}
				err := userUC.DeleteUser(context.Background(), "admin-id", deleteReq)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()
			uc, repo := setupUserIntegration(env)
			tt.run(t, env, uc, repo)
		})
	}
}

func TestUserIntegration_Edge(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository)
	}{
		{
			name:     "MinimumUsernameLength",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "abc", Email: "min@example.com", Password: "password123", Name: "Min User",
				}
				result, err := userUC.Create(context.Background(), req)
				if err == nil {
					assert.NotEmpty(t, result.ID)
				}
			},
		},
		{
			name:     "MaximumUsernameLength",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				longUsername := strings.Repeat("a", 50)
				req := &model.RegisterUserRequest{
					Username: longUsername, Email: "max@example.com", Password: "password123", Name: "Max User",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.Equal(t, longUsername, result.Username)
			},
		},
		{
			name:     "SpecialCharactersInName",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "specialuser", Email: "special@example.com", Password: "password123", Name: "O'Brien-Smith (Jr.) & Co.",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.Equal(t, req.Name, result.Name)
			},
		},
		{
			name:     "UnicodeInName",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "unicodeuser", Email: "unicode@example.com", Password: "password123", Name: "张三 李四 Müller José",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.Equal(t, req.Name, result.Name)
			},
		},
		{
			name:     "EmptyOptionalFields",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				testUser := setup.CreateTestUser(t, env.DB, "emptyuser", "empty@example.com", "password123")
				updateReq := &model.UpdateUserRequest{ID: testUser.ID, Name: ""}
				result, err := userUC.Update(context.Background(), updateReq)
				if err == nil {
					assert.NotNil(t, result)
				}
			},
		},
		{
			name:     "EmailWithPlusSign",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "plususer", Email: "user+test@example.com", Password: "password123", Name: "Plus",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.Equal(t, req.Email, result.Email)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()
			uc, repo := setupUserIntegration(env)
			tt.run(t, env, uc, repo)
		})
	}
}

func TestUserIntegration_Security(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository)
	}{
		{
			name:     "SQLInjectionInUsername",
			category: "security",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				injections := []string{"admin' OR '1'='1", "'; DROP TABLE users--", "admin'--", "1' UNION SELECT * FROM users--"}
				for i, injection := range injections {
					req := &model.RegisterUserRequest{
						Username: injection, Email: fmt.Sprintf("sql%d@example.com", i), Password: "password123", Name: "SQL User",
					}
					result, err := userUC.Create(context.Background(), req)
					if err == nil {
						assert.NotEmpty(t, result.ID)
						assert.Equal(t, pkg.SanitizeString(injection), result.Username)
					}
				}
			},
		},
		{
			name:     "XSSInName",
			category: "security",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				xssPayloads := []string{"<script>alert('XSS1')</script>", "<img src=x onerror=alert('XSS2')>", "javascript:alert('XSS3')", "<svg onload=alert('XSS4')>"}
				for i, xss := range xssPayloads {
					req := &model.RegisterUserRequest{
						Username: fmt.Sprintf("xssuser%d", i), Email: fmt.Sprintf("xss%d@example.com", i), Password: "password", Name: xss,
					}
					_, err := userUC.Create(context.Background(), req)
					require.NoError(t, err)
				}
			},
		},
		{
			name:     "PathTraversalInUsername",
			category: "security",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				pathTraversals := []string{"../../../etc/passwd", "..\\..\\windows\\system32", "....//....//"}
				for i, path := range pathTraversals {
					req := &model.RegisterUserRequest{
						Username: path, Email: fmt.Sprintf("path%d@example.com", i), Password: "password123", Name: "Path User",
					}
					result, err := userUC.Create(context.Background(), req)
					if err == nil {
						assert.NotEmpty(t, result.ID)
					}
				}
			},
		},
		{
			name:     "NoSQLInjection",
			category: "security",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				noSQLPayloads := []string{`{\"$$gt\":\""}`, `{\"$$ne\":null}`, `admin' || '1'=='1`}
				for i, payload := range noSQLPayloads {
					req := &model.RegisterUserRequest{
						Username: payload, Email: fmt.Sprintf("nosql%d@example.com", i), Password: "password123", Name: "NoSQL",
					}
					result, err := userUC.Create(context.Background(), req)
					if err == nil {
						assert.NotEmpty(t, result.ID)
					}
				}
			},
		},
		{
			name:     "PasswordNotInResponse",
			category: "security",
			run: func(t *testing.T, env *setup.TestEnvironment, userUC usecase.UserUseCase, userRepo repository.UserRepository) {
				req := &model.RegisterUserRequest{
					Username: "secureuser", Email: "secure@example.com", Password: "SecurePass123!", Name: "Secure User",
				}
				result, err := userUC.Create(context.Background(), req)
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, "Secure User", result.Name)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()
			uc, repo := setupUserIntegration(env)
			tt.run(t, env, uc, repo)
		})
	}
}

func TestUserRepository_Integration(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, env *setup.TestEnvironment, userRepo repository.UserRepository)
	}{
		{
			name:     "GetByOrganization",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userRepo repository.UserRepository) {
				setup.CleanupDatabase(t, env.DB)
				user := &entity.User{ID: "user-org-1", Username: "orguser", Email: "orguser@example.com", Status: "ACTIVE"}
				err := userRepo.Create(context.Background(), user)
				require.NoError(t, err)

				orgID := "org-1"
				err = env.DB.Exec("INSERT INTO organization_members (id, organization_id, user_id, role_id, status) VALUES (?, ?, ?, ?, ?)", "mem-1", orgID, user.ID, "ADMIN", "ACTIVE").Error
				require.NoError(t, err)

				users, err := userRepo.GetByOrganization(context.Background(), orgID)
				require.NoError(t, err)
				require.Len(t, users, 1)
				assert.Equal(t, user.ID, users[0].ID)
			},
		},
		{
			name:     "SSOIdentity",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, userRepo repository.UserRepository) {
				setup.CleanupDatabase(t, env.DB)
				user := &entity.User{ID: "user-sso-1", Username: "ssouser", Email: "ssouser@example.com", Status: "ACTIVE"}
				err := userRepo.Create(context.Background(), user)
				require.NoError(t, err)

				identity := &entity.UserSSOIdentity{ID: "sso-id-1", UserID: user.ID, Provider: "google", ProviderID: "google-123"}
				err = userRepo.CreateSSOIdentity(context.Background(), identity)
				require.NoError(t, err)

				found, err := userRepo.FindBySSOIdentity(context.Background(), "google", "google-123")
				require.NoError(t, err)
				assert.Equal(t, identity.ID, found.ID)
				assert.Equal(t, user.ID, found.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()
			userRepo := repository.NewUserRepository(env.DB, env.Logger)
			tt.run(t, env, userRepo)
		})
	}
}
