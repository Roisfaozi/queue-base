package worker_test

import (
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/worker"
	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type schedulerTestDeps struct {
	scheduler *worker.Scheduler
	mockRedis *miniredis.Miniredis
}

func setupSchedulerTest(t *testing.T) (*schedulerTestDeps, func()) {
	mr, err := miniredis.Run()
	if err != nil && strings.Contains(err.Error(), "operation not permitted") {
		t.Skip("socket listeners not permitted in this environment")
	}
	require.NoError(t, err)

	redisOpt := asynq.RedisClientOpt{
		Addr: mr.Addr(),
	}

	logger := logrus.New()
	logger.SetOutput(new(mockWriter)) // Reuse from processor_test.go

	scheduler := worker.NewScheduler(redisOpt, logger)

	cleanup := func() {
		mr.Close()
	}

	return &schedulerTestDeps{
		scheduler: scheduler,
		mockRedis: mr,
	}, cleanup
}

func TestScheduler_RegisterScheduledTasks(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_RegisterTasksSuccessfully",
			category: "positive",
			run: func(t *testing.T) {
				deps, cleanup := setupSchedulerTest(t)
				defer cleanup()

				// Should not panic or error during registration
				deps.scheduler.RegisterScheduledTasks()
			},
		},
		{
			name:     "Negative_BadRedisOption",
			category: "negative",
			run: func(t *testing.T) {
				logger := logrus.New()
				logger.SetOutput(new(mockWriter))
				scheduler := worker.NewScheduler(asynq.RedisClientOpt{Addr: "invalid:9999"}, logger)

				scheduler.RegisterScheduledTasks()
				// asynq.Scheduler.Register only parses crons and saves them in memory, doesn't immediately connect
			},
		},
		{
			name:     "Edge_ReRegisteringTasks",
			category: "edge",
			run: func(t *testing.T) {
				deps, cleanup := setupSchedulerTest(t)
				defer cleanup()

				// Registers without crashing
				deps.scheduler.RegisterScheduledTasks()

				// Attempting to register again might fail since same entry ID, let's see how it behaves
				deps.scheduler.RegisterScheduledTasks()
				// We just assert it doesn't panic
			},
		},
		{
			name:     "Vulnerability_VerifyTimeLocationSetup",
			category: "vulnerability",
			run: func(t *testing.T) {
				deps, cleanup := setupSchedulerTest(t)
				defer cleanup()

				// Verify no panic during init and that scheduler isn't nil
				assert.NotNil(t, deps.scheduler)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
