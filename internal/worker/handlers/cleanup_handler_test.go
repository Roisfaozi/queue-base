package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"testing"

	authMocks "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	userMocks "github.com/Roisfaozi/queue-base/internal/modules/user/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/worker/handlers"
	"github.com/Roisfaozi/queue-base/internal/worker/tasks"
	"github.com/Roisfaozi/queue-base/internal/worker/test/mocks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type cleanupTestDeps struct {
	AuthRepo  *authMocks.MockTokenRepository
	UserRepo  *userMocks.MockUserRepository
	AuditRepo *mocks.MockAuditRepository
	Handler   *handlers.CleanupTaskHandler
}

func setupCleanupHandlerTest() *cleanupTestDeps {
	authRepo := new(authMocks.MockTokenRepository)
	userRepo := new(userMocks.MockUserRepository)
	auditRepo := new(mocks.MockAuditRepository)
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	handler := handlers.NewCleanupTaskHandler(authRepo, userRepo, auditRepo, logger)

	return &cleanupTestDeps{
		AuthRepo:  authRepo,
		UserRepo:  userRepo,
		AuditRepo: auditRepo,
		Handler:   handler,
	}
}

func TestProcessCleanupExpiredTokens(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		setupMock func(deps *cleanupTestDeps)
		wantErr   bool
	}{
		{
			name:     "Success",
			category: "positive",
			setupMock: func(deps *cleanupTestDeps) {
				deps.AuthRepo.On("DeleteExpiredResetTokens", mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "Failure",
			category: "negative",
			setupMock: func(deps *cleanupTestDeps) {
				deps.AuthRepo.On("DeleteExpiredResetTokens", mock.Anything).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupCleanupHandlerTest()
			tt.setupMock(deps)
			task := asynq.NewTask(tasks.TypeCleanupExpiredTokens, nil)

			err := deps.Handler.ProcessCleanupExpiredTokens(context.Background(), task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			deps.AuthRepo.AssertExpectations(t)
		})
	}
}

func TestProcessCleanupSoftDeletedEntities(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		payload   []byte
		setupMock func(deps *cleanupTestDeps)
		wantErr   bool
	}{
		{
			name:     "Success",
			category: "positive",
			payload: func() []byte {
				p, _ := json.Marshal(tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 30})
				return p
			}(),
			setupMock: func(deps *cleanupTestDeps) {
				deps.UserRepo.On("HardDeleteSoftDeletedUsers", mock.Anything, 30).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "UnmarshalError",
			category:  "negative",
			payload:   []byte("invalid json"),
			setupMock: func(deps *cleanupTestDeps) {},
			wantErr:   true,
		},
		{
			name:     "RepoError",
			category: "negative",
			payload: func() []byte {
				p, _ := json.Marshal(tasks.CleanupSoftDeletedEntitiesPayload{RetentionDays: 30})
				return p
			}(),
			setupMock: func(deps *cleanupTestDeps) {
				deps.UserRepo.On("HardDeleteSoftDeletedUsers", mock.Anything, 30).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupCleanupHandlerTest()
			tt.setupMock(deps)
			task := asynq.NewTask(tasks.TypeCleanupSoftDeletedEntities, tt.payload)

			err := deps.Handler.ProcessCleanupSoftDeletedEntities(context.Background(), task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			deps.UserRepo.AssertExpectations(t)
		})
	}
}

func TestProcessPruneAuditLogs(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		payload   []byte
		setupMock func(deps *cleanupTestDeps)
		wantErr   bool
	}{
		{
			name:     "Success",
			category: "positive",
			payload: func() []byte {
				p, _ := json.Marshal(tasks.PruneAuditLogsPayload{RetentionDays: 180})
				return p
			}(),
			setupMock: func(deps *cleanupTestDeps) {
				deps.AuditRepo.On("DeleteLogsOlderThan", mock.Anything, mock.AnythingOfType("int64")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "UnmarshalError",
			category:  "negative",
			payload:   []byte("invalid json"),
			setupMock: func(deps *cleanupTestDeps) {},
			wantErr:   true,
		},
		{
			name:     "RepoError",
			category: "negative",
			payload: func() []byte {
				p, _ := json.Marshal(tasks.PruneAuditLogsPayload{RetentionDays: 180})
				return p
			}(),
			setupMock: func(deps *cleanupTestDeps) {
				deps.AuditRepo.On("DeleteLogsOlderThan", mock.Anything, mock.AnythingOfType("int64")).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := setupCleanupHandlerTest()
			tt.setupMock(deps)
			task := asynq.NewTask(tasks.TypePruneAuditLogs, tt.payload)

			err := deps.Handler.ProcessPruneAuditLogs(context.Background(), task)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			deps.AuditRepo.AssertExpectations(t)
		})
	}
}
