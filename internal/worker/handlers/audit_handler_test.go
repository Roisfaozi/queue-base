package handlers_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/worker/handlers"
	"github.com/Roisfaozi/queue-base/internal/worker/tasks"
	"github.com/hibiken/asynq"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAuditHandlerTest() (*mocks.MockAuditUseCase, *handlers.AuditTaskHandler) {
	uc := new(mocks.MockAuditUseCase)
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	handler := handlers.NewAuditTaskHandler(logger, uc)
	return uc, handler
}

func TestAuditTaskHandler_ProcessTaskAuditLog(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		payload   []byte
		setupMock func(uc *mocks.MockAuditUseCase)
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "Success",
			category: "positive",
			payload: func() []byte {
				b, _ := json.Marshal(auditModel.CreateAuditLogRequest{UserID: "user123", Action: "CREATE"})
				return b
			}(),
			setupMock: func(uc *mocks.MockAuditUseCase) {
				uc.EXPECT().LogActivity(mock.Anything, auditModel.CreateAuditLogRequest{UserID: "user123", Action: "CREATE"}).Return(nil)
			},
			wantErr: false,
		},
		{
			name:      "InvalidPayload",
			category:  "negative",
			payload:   []byte("invalid json"),
			setupMock: nil,
			wantErr:   true,
			errMsg:    "failed to unmarshal",
		},
		{
			name:     "UseCaseError",
			category: "negative",
			payload: func() []byte {
				b, _ := json.Marshal(auditModel.CreateAuditLogRequest{UserID: "user123", Action: "CREATE"})
				return b
			}(),
			setupMock: func(uc *mocks.MockAuditUseCase) {
				uc.EXPECT().LogActivity(mock.Anything, auditModel.CreateAuditLogRequest{UserID: "user123", Action: "CREATE"}).Return(errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "failed to log audit activity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, handler := setupAuditHandlerTest()
			if tt.setupMock != nil {
				tt.setupMock(uc)
			}
			task := asynq.NewTask(tasks.TypeAuditLogCreate, tt.payload)
			err := handler.ProcessTaskAuditLog(context.Background(), task)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				uc.AssertExpectations(t)
			}
		})
	}
}

func TestAuditTaskHandler_ProcessTaskAuditLogExport(t *testing.T) {
	// Clean up exports directory after tests
	defer func() { _ = os.RemoveAll("exports") }()

	tests := []struct {
		name      string
		category  string
		payload   []byte
		setupMock func(uc *mocks.MockAuditUseCase)
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "Success",
			category: "positive",
			payload: func() []byte {
				b, _ := json.Marshal(auditModel.AuditLogExportPayload{
					UserID:         "user123",
					OrganizationID: "org123",
					FromDate:       "2023-01-01",
					ToDate:         "2023-01-31",
				})
				return b
			}(),
			setupMock: func(uc *mocks.MockAuditUseCase) {
				uc.EXPECT().ExportLogs(mock.Anything, "2023-01-01", "2023-01-31", mock.AnythingOfType("func([]model.AuditLogResponse) error")).
					RunAndReturn(func(ctx context.Context, from, to string, process func([]auditModel.AuditLogResponse) error) error {
						return process([]auditModel.AuditLogResponse{
							{
								ID:        "log1",
								UserID:    "user123",
								Action:    "LOGIN",
								Entity:    "User",
								EntityID:  "user123",
								CreatedAt: 1672531200,
							},
						})
					})
			},
			wantErr: false,
		},
		{
			name:      "InvalidPayload",
			category:  "negative",
			payload:   []byte("invalid json"),
			setupMock: nil,
			wantErr:   true,
			errMsg:    "failed to unmarshal",
		},
		{
			name:     "UseCaseError",
			category: "negative",
			payload: func() []byte {
				b, _ := json.Marshal(auditModel.AuditLogExportPayload{
					UserID:   "user123",
					FromDate: "2023-01-01",
					ToDate:   "2023-01-31",
				})
				return b
			}(),
			setupMock: func(uc *mocks.MockAuditUseCase) {
				uc.EXPECT().ExportLogs(mock.Anything, "2023-01-01", "2023-01-31", mock.AnythingOfType("func([]model.AuditLogResponse) error")).Return(errors.New("export failed"))
			},
			wantErr: true,
			errMsg:  "failed to export logs",
		},
		{
			name:     "ProcessCallbackError",
			category: "negative",
			payload: func() []byte {
				b, _ := json.Marshal(auditModel.AuditLogExportPayload{
					UserID: "user123",
				})
				return b
			}(),
			setupMock: func(uc *mocks.MockAuditUseCase) {
				uc.EXPECT().ExportLogs(mock.Anything, "", "", mock.AnythingOfType("func([]model.AuditLogResponse) error")).
					RunAndReturn(func(ctx context.Context, from, to string, process func([]auditModel.AuditLogResponse) error) error {
						return errors.New("write error")
					})
			},
			wantErr: true,
			errMsg:  "failed to export logs: write error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, handler := setupAuditHandlerTest()
			if tt.setupMock != nil {
				tt.setupMock(uc)
			}
			task := asynq.NewTask(tasks.TypeAuditLogExport, tt.payload)
			err := handler.ProcessTaskAuditLogExport(context.Background(), task)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				uc.AssertExpectations(t)
			}
		})
	}
}
