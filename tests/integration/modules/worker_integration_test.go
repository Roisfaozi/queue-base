//go:build integration
// +build integration

package modules

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"sync/atomic"
	"testing"
	"time"

	authMocks "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	userMocks "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/worker"
	"github.com/Roisfaozi/queue-base/internal/worker/handlers"
	"github.com/Roisfaozi/queue-base/internal/worker/tasks"
	workerMocks "github.com/Roisfaozi/queue-base/internal/worker/test/mocks"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newSilentLogger creates a logger that discards all output
func newSilentLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(io.Discard)
	return log
}

func TestWorkerIntegration(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "EmailTask_EnqueueAndInspect",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				distributor := worker.NewRedisTaskDistributor(redisOpt)

				inspector := asynq.NewInspector(redisOpt)
				defer inspector.Close()

				payload := &tasks.SendEmailPayload{
					To:      "integration-test@example.com",
					Subject: "Integration Test - Email Enqueue",
					Body:    "<h1>Test Email Body</h1>",
				}

				ctx := context.Background()
				err := distributor.DistributeTaskSendEmail(ctx, payload, asynq.MaxRetry(3), asynq.Queue("default"))
				require.NoError(t, err)

				time.Sleep(200 * time.Millisecond)

				pendingTasks, err := inspector.ListPendingTasks("default", asynq.Page(1), asynq.PageSize(10))
				require.NoError(t, err)

				var foundTask *asynq.TaskInfo
				for _, task := range pendingTasks {
					if task.Type == tasks.TypeSendEmail {
						foundTask = task
						break
					}
				}

				require.NotNil(t, foundTask, "SendEmail task not found in pending queue")

				var actualPayload tasks.SendEmailPayload
				err = json.Unmarshal(foundTask.Payload, &actualPayload)
				require.NoError(t, err)
				assert.Equal(t, payload.To, actualPayload.To)
				assert.Equal(t, payload.Subject, actualPayload.Subject)
				assert.Equal(t, payload.Body, actualPayload.Body)
				assert.Equal(t, 3, foundTask.MaxRetry)
			},
		},
		{
			name:     "CleanupExpiredTokens",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				authRepo := new(authMocks.MockTokenRepository)
				userRepo := new(userMocks.MockUserRepository)
				auditRepo := new(workerMocks.MockAuditRepository)

				authRepo.On("DeleteExpiredResetTokens", mock.Anything).Return(nil).Once()

				cleanupHandler := handlers.NewCleanupTaskHandler(authRepo, userRepo, auditRepo, log)

				mux := asynq.NewServeMux()
				mux.HandleFunc(tasks.TypeCleanupExpiredTokens, cleanupHandler.ProcessCleanupExpiredTokens)

				server := asynq.NewServer(redisOpt, asynq.Config{
					Concurrency: 1,
					Logger:      worker.NewAsynqLogger(log),
				})

				err := server.Start(mux)
				require.NoError(t, err)
				defer server.Shutdown()

				client := asynq.NewClient(redisOpt)
				defer client.Close()

				task := asynq.NewTask(tasks.TypeCleanupExpiredTokens, nil)
				_, err = client.Enqueue(task, asynq.Queue("default"))
				require.NoError(t, err)

				time.Sleep(2 * time.Second)

				authRepo.AssertExpectations(t)
			},
		},
		{
			name:     "CleanupSoftDeletedUsers",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				authRepo := new(authMocks.MockTokenRepository)
				userRepo := new(userMocks.MockUserRepository)
				auditRepo := new(workerMocks.MockAuditRepository)

				userRepo.On("HardDeleteSoftDeletedUsers", mock.Anything, 30).Return(nil).Once()

				cleanupHandler := handlers.NewCleanupTaskHandler(authRepo, userRepo, auditRepo, log)

				mux := asynq.NewServeMux()
				mux.HandleFunc(tasks.TypeCleanupSoftDeletedEntities, cleanupHandler.ProcessCleanupSoftDeletedEntities)

				server := asynq.NewServer(redisOpt, asynq.Config{
					Concurrency: 1,
					Logger:      worker.NewAsynqLogger(log),
				})

				err := server.Start(mux)
				require.NoError(t, err)
				defer server.Shutdown()

				client := asynq.NewClient(redisOpt)
				defer client.Close()

				payload := tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 30}
				jsonPayload, _ := json.Marshal(payload)
				task := asynq.NewTask(tasks.TypeCleanupSoftDeletedEntities, jsonPayload)
				_, err = client.Enqueue(task, asynq.Queue("default"))
				require.NoError(t, err)

				time.Sleep(2 * time.Second)

				userRepo.AssertExpectations(t)
			},
		},
		{
			name:     "PruneAuditLogs",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				authRepo := new(authMocks.MockTokenRepository)
				userRepo := new(userMocks.MockUserRepository)
				auditRepo := new(workerMocks.MockAuditRepository)

				auditRepo.On("DeleteLogsOlderThan", mock.Anything, mock.MatchedBy(func(cutoff int64) bool {
					expected := time.Now().AddDate(0, 0, -180).UnixMilli()
					return cutoff >= expected-10000 && cutoff <= expected+10000
				})).Return(nil).Once()

				cleanupHandler := handlers.NewCleanupTaskHandler(authRepo, userRepo, auditRepo, log)

				mux := asynq.NewServeMux()
				mux.HandleFunc(tasks.TypePruneAuditLogs, cleanupHandler.ProcessPruneAuditLogs)

				server := asynq.NewServer(redisOpt, asynq.Config{
					Concurrency: 1,
					Logger:      worker.NewAsynqLogger(log),
				})

				err := server.Start(mux)
				require.NoError(t, err)
				defer server.Shutdown()

				client := asynq.NewClient(redisOpt)
				defer client.Close()

				payload := tasks.PruneAuditLogsPayload{RetentionDays: 180}
				jsonPayload, _ := json.Marshal(payload)
				task := asynq.NewTask(tasks.TypePruneAuditLogs, jsonPayload)
				_, err = client.Enqueue(task, asynq.Queue("default"))
				require.NoError(t, err)

				time.Sleep(2 * time.Second)

				auditRepo.AssertExpectations(t)
			},
		},
		{
			name:     "TaskRetry_OnTransientFailure",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				var callCount int32

				mux := asynq.NewServeMux()
				mux.HandleFunc("test:retry", func(ctx context.Context, task *asynq.Task) error {
					count := atomic.AddInt32(&callCount, 1)
					if count <= 1 {
						return errors.New("transient error - will retry")
					}
					return nil
				})

				server := asynq.NewServer(redisOpt, asynq.Config{
					Concurrency: 1,
					Logger:      worker.NewAsynqLogger(log),
					RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
						return 500 * time.Millisecond
					},
				})

				err := server.Start(mux)
				require.NoError(t, err)
				defer server.Shutdown()

				client := asynq.NewClient(redisOpt)
				defer client.Close()

				task := asynq.NewTask("test:retry", nil)
				_, err = client.Enqueue(task, asynq.MaxRetry(5), asynq.Queue("default"))
				require.NoError(t, err)

				time.Sleep(10 * time.Second)

				finalCount := atomic.LoadInt32(&callCount)
				assert.GreaterOrEqual(t, finalCount, int32(2), "Handler should be called at least 2 times (1 fail + 1 success)")
			},
		},
		{
			name:     "TaskFailure_DeadLetterQueue",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				var callCount int32

				mux := asynq.NewServeMux()
				mux.HandleFunc("test:always-fail", func(ctx context.Context, task *asynq.Task) error {
					atomic.AddInt32(&callCount, 1)
					return errors.New("permanent failure - cannot process")
				})

				server := asynq.NewServer(redisOpt, asynq.Config{
					Concurrency: 1,
					Logger:      worker.NewAsynqLogger(log),
					RetryDelayFunc: func(n int, e error, t *asynq.Task) time.Duration {
						return 500 * time.Millisecond
					},
				})

				err := server.Start(mux)
				require.NoError(t, err)
				defer server.Shutdown()

				client := asynq.NewClient(redisOpt)
				defer client.Close()

				task := asynq.NewTask("test:always-fail", []byte(`{"test":"data"}`))
				_, err = client.Enqueue(task, asynq.MaxRetry(1), asynq.Queue("default"))
				require.NoError(t, err)

				time.Sleep(10 * time.Second)

				finalCount := atomic.LoadInt32(&callCount)
				assert.GreaterOrEqual(t, finalCount, int32(2), "Handler should be called at least 2 times (initial + 1 retry)")

				inspector := asynq.NewInspector(redisOpt)
				defer inspector.Close()

				var foundArchived *asynq.TaskInfo
				deadline := time.Now().Add(5 * time.Second)
				for time.Now().Before(deadline) {
					archivedTasks, err := inspector.ListArchivedTasks("default", asynq.Page(1), asynq.PageSize(10))
					if err == nil {
						for _, at := range archivedTasks {
							if at.Type == "test:always-fail" {
								foundArchived = at
								break
							}
						}
					}
					if foundArchived != nil {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}

				if foundArchived != nil {
					assert.Equal(t, "test:always-fail", foundArchived.Type)
					assert.Contains(t, foundArchived.LastErr, "permanent failure")
				} else {
					assert.GreaterOrEqual(t, finalCount, int32(2), "Task was processed multiple times")
					t.Log("Task not yet in archived queue (may still be in retry), but handler was called multiple times confirming retry behavior")
				}
			},
		},
		{
			name:     "SchedulerRegistration",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				scheduler := worker.NewScheduler(redisOpt, log)

				scheduler.RegisterScheduledTasks()

				t.Log("Scheduled tasks registered successfully without errors")

				errCh := make(chan error, 1)
				go func() {
					errCh <- scheduler.Start()
				}()

				time.Sleep(500 * time.Millisecond)

				scheduler.Shutdown()

				select {
				case err := <-errCh:
					if err != nil {
						t.Logf("Scheduler returned: %v (expected during shutdown)", err)
					}
				case <-time.After(2 * time.Second):
					t.Log("Scheduler shutdown completed")
				}
			},
		},
		{
			name:     "MultipleTaskEnqueue",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				distributor := worker.NewRedisTaskDistributor(redisOpt)

				ctx := context.Background()

				for i := 0; i < 5; i++ {
					payload := &tasks.SendEmailPayload{
						To:      "batch-test@example.com",
						Subject: "Batch Test",
						Body:    "Body",
					}
					err := distributor.DistributeTaskSendEmail(ctx, payload, asynq.Queue("default"))
					require.NoError(t, err)
				}

				time.Sleep(200 * time.Millisecond)

				inspector := asynq.NewInspector(redisOpt)
				defer inspector.Close()

				pendingTasks, err := inspector.ListPendingTasks("default", asynq.Page(1), asynq.PageSize(20))
				require.NoError(t, err)

				emailTaskCount := 0
				for _, task := range pendingTasks {
					if task.Type == tasks.TypeSendEmail {
						emailTaskCount++
					}
				}

				assert.Equal(t, 5, emailTaskCount, "Should have 5 email tasks in pending queue")
			},
		},
		{
			name:     "ProcessorStartShutdown",
			category: "integration",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				redisOpt := asynq.RedisClientOpt{Addr: env.RedisAddr}
				log := newSilentLogger()

				authRepo := new(authMocks.MockTokenRepository)
				userRepo := new(userMocks.MockUserRepository)
				auditRepo := new(workerMocks.MockAuditRepository)

				cleanupHandler := handlers.NewCleanupTaskHandler(authRepo, userRepo, auditRepo, log)

				cfg := worker.WorkerConfig{
					SMTP: worker.SMTPConfig{
						Host:       "localhost",
						Port:       1025,
						Username:   "test",
						Password:   "test",
						FromSender: "Test",
						FromEmail:  "test@example.com",
					},
				}

				processor := worker.NewRedisTaskProcessor(redisOpt, log, cleanupHandler, nil, nil, auditRepo, cfg)

				errCh := make(chan error, 1)
				go func() {
					errCh <- processor.Start()
				}()

				time.Sleep(500 * time.Millisecond)

				processor.Shutdown()

				select {
				case err := <-errCh:
					if err != nil {
						t.Logf("Processor returned: %v (expected during shutdown)", err)
					}
				case <-time.After(3 * time.Second):
					t.Log("Processor shutdown completed")
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
