//go:build integration
// +build integration

package scenarios

import (
	"context"
	"testing"
	"time"

	auditRepo "github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUC "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	authEntity "github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	authModel "github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	authRepo "github.com/Roisfaozi/queue-base/internal/modules/auth/repository"
	authUC "github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	orgRepo "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	userRepo "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/Roisfaozi/queue-base/pkg/sso"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/Roisfaozi/queue-base/pkg/util"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScenario_PasswordRecovery_Lifecycle(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_PasswordRecoveryFlow",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()
				setup.CleanupDatabase(t, env.DB)

				tm := tx.NewTransactionManager(env.DB, env.Logger)
				uRepo := userRepo.NewUserRepository(env.DB, env.Logger)
				tRepo := authRepo.NewTokenRepositoryRedis(env.Redis, env.Logger, env.DB, &util.RealClock{})
				aucRepo := auditRepo.NewAuditRepository(env.DB, env.Logger)

				_ = auditUC.NewAuditUseCase(aucRepo, env.Logger, nil, nil)
				jwtManager := jwt.NewJWTManager("secret", "refresh", 15*time.Minute, 24*time.Hour)

				oRepo := orgRepo.NewOrganizationRepository(env.DB)
				authz := authRepo.NewCasbinAdapter(env.Enforcer, "role:user", "global")
				authService := authUC.NewAuthUsecase(5, 30*time.Minute, jwtManager, tRepo, uRepo, oRepo, tm, env.Logger, nil, authz, nil, nil, make(map[string]sso.Provider))

				oldPassword := "OldPass123!"
				newPassword := "NewPass456!"
				user := setup.CreateTestUser(t, env.DB, "forgot_user", "forgot@test.com", oldPassword)

				err := authService.ForgotPassword(context.Background(), user.Email)
				require.NoError(t, err)

				var resetToken authEntity.PasswordResetToken
				err = env.DB.Where("email = ?", user.Email).First(&resetToken).Error
				require.NoError(t, err, "Reset token should be saved in DB")
				assert.NotEmpty(t, resetToken.Token)

				err = authService.ResetPassword(context.Background(), resetToken.Token, newPassword)
				require.NoError(t, err)

				var checkToken authEntity.PasswordResetToken
				err = env.DB.Where("token = ?", resetToken.Token).First(&checkToken).Error
				assert.Error(t, err, "Token should be deleted after use")

				_, _, err = authService.Login(context.Background(), authModel.LoginRequest{
					Username: user.Username, Password: oldPassword,
				})
				assert.Error(t, err, "Login with old password should fail")

				resp, _, err := authService.Login(context.Background(), authModel.LoginRequest{
					Username: user.Username, Password: newPassword,
				})
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.AccessToken)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
