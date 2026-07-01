package handlers

import (
	"context"
	"errors"
	"strings"
	"testing"

	auditEntity "github.com/Roisfaozi/queue-base/internal/modules/audit/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/test/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestOutboxTaskHandler_ProcessAuditOutbox_Robustness(t *testing.T) {
	logger := logrus.New() // Note: assuming logrus is available in package scope or I'll import it

	tests := []struct {
		name      string
		category  string
		setupMock func(mockRepo *mocks.MockAuditRepository)
	}{
		{
			name:     "Success Path - Moves entry to main log and deletes from outbox",
			category: "positive",
			setupMock: func(mockRepo *mocks.MockAuditRepository) {
				ctx := mock.Anything
				orgID := "org-1"
				entries := []*auditEntity.AuditOutbox{
					{
						ID:             "outbox-1",
						OrganizationID: &orgID,
						UserID:         "user-1",
						Action:         "UPDATE",
						Entity:         "Profile",
						EntityID:       "user-1",
						CreatedAt:      123456789,
					},
				}
				mockRepo.On("FindPendingOutbox", ctx, 50).Return(entries, nil)
				mockRepo.On("Create", ctx, mock.MatchedBy(func(log *auditEntity.AuditLog) bool {
					return log.ID == "outbox-1" && log.UserID == "user-1" && log.EntityID == "user-1"
				})).Return(nil)
				mockRepo.On("DeleteOutbox", ctx, "outbox-1").Return(nil)
			},
		},
		{
			name:     "Failure Path - Update status to failed when main log creation fails",
			category: "negative",
			setupMock: func(mockRepo *mocks.MockAuditRepository) {
				ctx := mock.Anything
				entries := []*auditEntity.AuditOutbox{
					{ID: "outbox-2", UserID: "user-1", Action: "LOGIN"},
				}
				dbErr := errors.New("database connection lost")
				mockRepo.On("FindPendingOutbox", ctx, 50).Return(entries, nil)
				mockRepo.On("Create", ctx, mock.Anything).Return(dbErr)
				mockRepo.On("UpdateOutboxStatus", ctx, "outbox-2", auditEntity.OutboxStatusFailed, mock.MatchedBy(func(err string) bool {
					return strings.Contains(err, dbErr.Error())
				})).Return(nil)
			},
		},
		{
			name:     "Delete Failure - Marks moved entry completed to avoid duplicate replay",
			category: "negative",
			setupMock: func(mockRepo *mocks.MockAuditRepository) {
				ctx := mock.Anything
				entries := []*auditEntity.AuditOutbox{
					{ID: "outbox-3", UserID: "user-1", Action: "UPDATE", Entity: "Profile", EntityID: "user-1"},
				}
				deleteErr := errors.New("delete failed")
				mockRepo.On("FindPendingOutbox", ctx, 50).Return(entries, nil)
				mockRepo.On("Create", ctx, mock.AnythingOfType("*entity.AuditLog")).Return(nil)
				mockRepo.On("DeleteOutbox", ctx, "outbox-3").Return(deleteErr)
				mockRepo.On("UpdateOutboxStatus", ctx, "outbox-3", auditEntity.OutboxStatusCompleted, deleteErr.Error()).Return(nil)
			},
		},
		{
			name:     "Empty Outbox - Does nothing gracefully",
			category: "positive",
			setupMock: func(mockRepo *mocks.MockAuditRepository) {
				ctx := mock.Anything
				mockRepo.On("FindPendingOutbox", ctx, 50).Return([]*auditEntity.AuditOutbox{}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockAuditRepository)
			handler := NewOutboxTaskHandler(mockRepo, logger)
			tt.setupMock(mockRepo)

			err := handler.ProcessAuditOutbox(context.Background(), nil)

			assert.NoError(t, err)
			mockRepo.AssertExpectations(t)
		})
	}
}
