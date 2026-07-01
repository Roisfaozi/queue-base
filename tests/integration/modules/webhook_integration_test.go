//go:build integration
// +build integration

package modules

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/webhook/model"
	"github.com/Roisfaozi/queue-base/internal/modules/webhook/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/webhook/usecase"
	"github.com/Roisfaozi/queue-base/internal/worker"
	"github.com/Roisfaozi/queue-base/internal/worker/handlers"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/hibiken/asynq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWebhookIntegration(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, uc usecase.WebhookUseCase, ctx context.Context)
	}{
		{
			name:     "Positive_TriggerAndDelivery",
			category: "integration",
			run: func(t *testing.T, uc usecase.WebhookUseCase, ctx context.Context) {
				received := make(chan bool, 1)
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					received <- true
					w.WriteHeader(http.StatusOK)
				}))
				defer mockServer.Close()

				wh, err := uc.Create(ctx, model.CreateWebhookRequest{
					Name: "Pos Test", OrganizationID: "org-1", URL: mockServer.URL, Events: []string{"user.created"}, Secret: "super-secret-key",
				})
				require.NoError(t, err)
				require.NotNil(t, wh)

				err = uc.Trigger(ctx, model.TriggerWebhookRequest{
					OrganizationID: "org-1", EventType: "user.created", Payload: map[string]string{"data": "ok"},
				})
				require.NoError(t, err)

				select {
				case <-received:
				case <-time.After(5 * time.Second):
					t.Fatal("Webhook not delivered")
				}

				var logs []any
				require.Eventually(t, func() bool {
					logs, _ = uc.FindLogs(ctx, wh.ID, "org-1", 1, 0)
					return len(logs) > 0
				}, 5*time.Second, 200*time.Millisecond)
				require.NotEmpty(t, logs)
			},
		},
		{
			name:     "Negative_TargetServerError",
			category: "integration",
			run: func(t *testing.T, uc usecase.WebhookUseCase, ctx context.Context) {
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				defer mockServer.Close()

				wh, err := uc.Create(ctx, model.CreateWebhookRequest{
					Name: "Neg Test", OrganizationID: "org-1", URL: mockServer.URL, Events: []string{"user.error"}, Secret: "super-secret-key",
				})
				require.NoError(t, err)
				require.NotNil(t, wh)

				_ = uc.Trigger(ctx, model.TriggerWebhookRequest{
					OrganizationID: "org-1", EventType: "user.error", Payload: map[string]string{"err": "fail"},
				})

				require.Eventually(t, func() bool {
					logs, _ := uc.FindLogs(ctx, wh.ID, "org-1", 1, 0)
					return len(logs) > 0
				}, 5*time.Second, 200*time.Millisecond)
			},
		},
		{
			name:     "Edge_InactiveWebhook",
			category: "edge",
			run: func(t *testing.T, uc usecase.WebhookUseCase, ctx context.Context) {
				callCount := int32(0)
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					atomic.AddInt32(&callCount, 1)
				}))
				defer mockServer.Close()

				wh, err := uc.Create(ctx, model.CreateWebhookRequest{
					Name: "Inactive Test", OrganizationID: "org-1", URL: mockServer.URL, Events: []string{"user.inactive"}, Secret: "super-secret-key",
				})
				require.NoError(t, err)

				active := false
				_, err = uc.Update(ctx, wh.ID, "org-1", model.UpdateWebhookRequest{IsActive: &active})
				require.NoError(t, err)

				_ = uc.Trigger(ctx, model.TriggerWebhookRequest{
					OrganizationID: "org-1", EventType: "user.inactive", Payload: map[string]string{"id": "1"},
				})

				time.Sleep(1 * time.Second)
				assert.Equal(t, int32(0), atomic.LoadInt32(&callCount))
			},
		},
		{
			name:     "Security_HMACSignature",
			category: "security",
			run: func(t *testing.T, uc usecase.WebhookUseCase, ctx context.Context) {
				secret := "secure-key-123"
				payload := map[string]string{"secure": "data"}
				payloadJSON, _ := json.Marshal(payload)

				sigChan := make(chan string, 1)
				mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					sigChan <- r.Header.Get("X-Webhook-Signature")
					w.WriteHeader(http.StatusOK)
				}))
				defer mockServer.Close()

				_, err := uc.Create(ctx, model.CreateWebhookRequest{
					Name: "Sec Test", OrganizationID: "org-sec", URL: mockServer.URL, Events: []string{"secure.event"}, Secret: secret,
				})
				require.NoError(t, err)

				_ = uc.Trigger(ctx, model.TriggerWebhookRequest{
					OrganizationID: "org-sec", EventType: "secure.event", Payload: payload,
				})

				select {
				case receivedSignature := <-sigChan:
					h := hmac.New(sha256.New, []byte(secret))
					h.Write(payloadJSON)
					expectedSignature := hex.EncodeToString(h.Sum(nil))
					assert.Equal(t, expectedSignature, receivedSignature)
				case <-time.After(5 * time.Second):
					t.Fatal("Timeout")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
			distributor := worker.NewRedisTaskDistributor(redisOpt)
			webhookRepo := repository.NewWebhookRepository(env.DB, env.Logger)
			webhookHandler := handlers.NewWebhookHandler(webhookRepo, env.Logger)
			cleanupHandler := handlers.NewCleanupTaskHandler(nil, nil, nil, env.Logger)
			processor := worker.NewRedisTaskProcessor(redisOpt, env.Logger, cleanupHandler, webhookHandler, nil, nil, worker.WorkerConfig{})
			env.StartWorker(processor)

			validate := validator.New()
			webhookUC := usecase.NewWebhookUseCase(webhookRepo, distributor, env.Logger, validate)
			ctx := context.Background()

			tt.run(t, webhookUC, ctx)
		})
	}
}
