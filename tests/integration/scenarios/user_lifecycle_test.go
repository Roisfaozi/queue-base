//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	auditRepository "github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	authModel "github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	authRepository "github.com/Roisfaozi/queue-base/internal/modules/auth/repository"
	authUseCase "github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	userModel "github.com/Roisfaozi/queue-base/internal/modules/user/model"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	userUseCase "github.com/Roisfaozi/queue-base/internal/modules/user/usecase"
	"github.com/Roisfaozi/queue-base/internal/worker"
	"github.com/Roisfaozi/queue-base/internal/worker/handlers"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/pkg/sso"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/Roisfaozi/queue-base/pkg/util"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserLifecycle_FullFlow(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, env *setup.TestEnvironment, deps *lifecycleDeps)
	}{
		{
			name:     "Register new user succeeds",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, deps *lifecycleDeps) {
				regReq := &userModel.RegisterUserRequest{
					Username: "lifecycle", Email: "lifecycle@example.com", Password: "password123", Name: "Life Cycle",
				}
				userResp, err := deps.userUC.Create(context.Background(), regReq)
				require.NoError(t, err)
				assert.NotEmpty(t, userResp.ID)
			},
		},
		{
			name:     "Login after registration",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, deps *lifecycleDeps) {
				loginReq := authModel.LoginRequest{Username: "lifecycle-login", Password: "password123"}
				_, _, err := deps.authUC.Login(context.Background(), loginReq)
				require.NoError(t, err)
			},
		},
		{
			name:     "Update user profile",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, deps *lifecycleDeps) {
				updateReq := &userModel.UpdateUserRequest{
					ID: deps.userID, Name: "Updated Life",
				}
				updateResp, err := deps.userUC.Update(context.Background(), updateReq)
				require.NoError(t, err)
				assert.Equal(t, "Updated Life", updateResp.Name)
			},
		},
		{
			name:     "Delete user soft-deletes",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, deps *lifecycleDeps) {
				deleteReq := &userModel.DeleteUserRequest{ID: deps.userID}
				err := deps.userUC.DeleteUser(context.Background(), deps.userID, deleteReq)
				require.NoError(t, err)

				err = handlers.NewOutboxTaskHandler(deps.auditRepo, env.Logger).ProcessAuditOutbox(context.Background(), nil)
				require.NoError(t, err)
			},
		},
		{
			name:     "Audit logs capture full lifecycle",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment, deps *lifecycleDeps) {
				var userLogs []auditModel.AuditLogResponse
				var actions map[string]bool

				require.Eventually(t, func() bool {
					logs, _, err := deps.auditUC.GetLogsDynamic(context.Background(), &querybuilder.DynamicFilter{
						Sort: &[]querybuilder.SortModel{{ColId: "CreatedAt", Sort: "asc"}},
					})
					if err != nil {
						return false
					}

					userLogs = nil
					actions = make(map[string]bool)
					for _, l := range logs {
						if l.UserID == deps.userID || l.EntityID == deps.userID {
							userLogs = append(userLogs, l)
							actions[l.Action] = true
						}
					}

					return len(userLogs) >= 4 && actions["CREATE"] && actions["LOGIN"] && actions["UPDATE"] && actions["DELETE"]
				}, 5*time.Second, 100*time.Millisecond, "Should have at least 4 audit entries covering all lifecycle actions")

				assert.True(t, actions["CREATE"], "CREATE log missing")
				assert.True(t, actions["LOGIN"], "LOGIN log missing")
				assert.True(t, actions["UPDATE"], "UPDATE log missing")
				assert.True(t, actions["DELETE"], "DELETE log missing")
			},
		},
	}

	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	deps := buildLifecycleDeps(t, env)

	ctx := context.Background()
	regReq := &userModel.RegisterUserRequest{
		Username: "lifecycle-login", Email: "lifecycle-login@example.com", Password: "password123", Name: "Life Cycle",
	}
	userResp, err := deps.userUC.Create(ctx, regReq)
	require.NoError(t, err)
	deps.userID = userResp.ID

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, env, deps)
		})
	}
}

type lifecycleDeps struct {
	userUC    userUseCase.UserUseCase
	authUC    authUseCase.AuthUseCase
	auditUC   auditUseCase.AuditUseCase
	auditRepo auditUseCase.AuditRepository
	userID    string
}

func buildLifecycleDeps(t *testing.T, env *setup.TestEnvironment) *lifecycleDeps {
	t.Helper()

	jwtManager := jwt.NewJWTManager("lifecycle-secret", "lifecycle-refresh", 15*time.Minute, 24*time.Hour)
	tokenRepo := authRepository.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
	userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
	tm := tx.NewTransactionManager(env.DB, env.Logger)
	auditRepo := auditRepository.NewAuditRepository(env.DB, env.Logger)
	auditUC := auditUseCase.NewAuditUseCase(auditRepo, env.Logger, nil, nil)

	oRepo := orgRepo.NewOrganizationRepository(env.DB)
	authz := authRepository.NewCasbinAdapter(env.Enforcer, "role:user", "global")
	redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
	taskDistributor := worker.NewRedisTaskDistributor(redisOpt)

	cleanupHandler := handlers.NewCleanupTaskHandler(tokenRepo, userRepo, auditRepo, env.Logger)
	workerCfg := worker.WorkerConfig{}
	processor := worker.NewRedisTaskProcessor(redisOpt, env.Logger, cleanupHandler, nil, auditUC, auditRepo, workerCfg)
	env.StartWorker(processor)

	authUC := authUseCase.NewAuthUsecase(5, 30*time.Minute, jwtManager, tokenRepo, userRepo, oRepo, tm, env.Logger, nil, authz, taskDistributor, nil, make(map[string]sso.Provider))
	userUC := userUseCase.NewUserUseCase(tm, env.Logger, userRepo, env.Enforcer, auditUC, authUC, nil, nil)

	return &lifecycleDeps{
		userUC:    userUC,
		authUC:    authUC,
		auditUC:   auditUC,
		auditRepo: auditRepo,
	}
}
