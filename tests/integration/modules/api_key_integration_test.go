//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/api_key/model"
	"github.com/Roisfaozi/queue-base/internal/modules/api_key/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/api_key/usecase"
	orgRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApiKeyIntegration_Lifecycle(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, uc usecase.ApiKeyUseCase, env *setup.TestEnvironment, user *setup.TestUser, orgID string)
	}{
		{
			name:     "Create and Authenticate with Caching",
			category: "positive",
			run: func(t *testing.T, uc usecase.ApiKeyUseCase, env *setup.TestEnvironment, user *setup.TestUser, orgID string) {
				ctx := context.Background()
				createReq := &model.CreateApiKeyRequest{
					Name:   "Prod Key",
					Scopes: []string{"read", "write"},
				}

				created, err := uc.Create(ctx, user.ID, orgID, createReq)
				require.NoError(t, err)
				require.NotNil(t, created)
				require.NotEmpty(t, created.Key)

				identity, err := uc.Authenticate(ctx, created.Key)
				require.NoError(t, err)
				assert.Equal(t, user.ID, identity.UserID)
				assert.Equal(t, "api_tester", identity.Username)

				actualKey := created.Key[8:] 
				hash := setup.HashSHA256(actualKey)
				cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", hash)

				exists := env.Redis.Exists(ctx, cacheKey).Val()
				assert.Equal(t, int64(1), exists, "Key should be cached in Redis after first authentication")

				identity2, err := uc.Authenticate(ctx, created.Key)
				require.NoError(t, err)
				assert.Equal(t, identity.ApiKeyID, identity2.ApiKeyID)
			},
		},
		{
			name:     "Revoke should invalidate Cache",
			category: "negative",
			run: func(t *testing.T, uc usecase.ApiKeyUseCase, env *setup.TestEnvironment, user *setup.TestUser, orgID string) {
				ctx := context.Background()
				created, _ := uc.Create(ctx, user.ID, orgID, &model.CreateApiKeyRequest{Name: "To Revoke"})

				identity, _ := uc.Authenticate(ctx, created.Key)
				actualKey := created.Key[8:]
				hash := setup.HashSHA256(actualKey)
				cacheKey := fmt.Sprintf("nexusos:api_key:v1:%s", hash)

				require.Equal(t, int64(1), env.Redis.Exists(ctx, cacheKey).Val())

				err := uc.Revoke(ctx, orgID, identity.ApiKeyID)
				require.NoError(t, err)

				require.Eventually(t, func() bool {
					return env.Redis.Exists(ctx, cacheKey).Val() == 0
				}, 2*time.Second, 100*time.Millisecond, "Cache should be deleted immediately upon revocation")

				_, err = uc.Authenticate(ctx, created.Key)
				assert.Error(t, err, "Authentication should fail after revocation")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			apiKeyRepo := repository.NewApiKeyRepository(env.DB)
			orgRepo := orgRepository.NewOrganizationRepository(env.DB)
			userRepo := userRepository.NewUserRepository(env.DB, env.Logger)
			uc := usecase.NewApiKeyUseCase(apiKeyRepo, orgRepo, userRepo, env.Redis, env.Logger)

			user := setup.CreateTestUser(t, env.DB, "api_tester", "api@test.com", "Password123!")
			org := setup.CreateTestOrganization(t, env.DB, user.ID, "API Key Test Org", "api-key-test-org")

			tt.run(t, uc, env, user, org.ID)
		})
	}
}