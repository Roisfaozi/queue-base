package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthUsecase_GetTicket(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		category string
		context  model.UserSessionContext
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			context: model.UserSessionContext{
				UserID:    TestUserID,
				OrgID:     "org123",
				SessionID: "sess123",
				Role:      "role:user",
				Username:  TestUsername,
			},
			setup: func(deps *testDependencies) {
				user := &userEntity.User{ID: TestUserID, Status: userEntity.UserStatusActive}
				deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(user, nil)
				deps.ticketManager.On("CreateTicket", mock.Anything, TestUserID, "org123", "sess123", "role:user", TestUsername).Return("ticket123", nil)
			},
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			context: model.UserSessionContext{
				UserID: TestUserID,
			},
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(nil, errors.New("not found"))
			},
			wantErr: usecase.ErrInvalidCredentials,
		},
		{
			name:     "Negative_UserSuspended",
			category: "negative",
			context: model.UserSessionContext{
				UserID: TestUserID,
			},
			setup: func(deps *testDependencies) {
				user := &userEntity.User{ID: TestUserID, Status: userEntity.UserStatusSuspended}
				deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(user, nil)
			},
			wantErr: usecase.ErrAccountSuspended,
		},
		{
			name:     "Negative_TicketManagerError",
			category: "negative",
			context: model.UserSessionContext{
				UserID:    TestUserID,
				OrgID:     "org123",
				SessionID: "sess123",
				Role:      "role:user",
				Username:  TestUsername,
			},
			setup: func(deps *testDependencies) {
				user := &userEntity.User{ID: TestUserID, Status: userEntity.UserStatusActive}
				deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(user, nil)
				deps.ticketManager.On("CreateTicket", mock.Anything, TestUserID, "org123", "sess123", "role:user", TestUsername).Return("", errors.New("redis err"))
			},
			wantErr: errors.New("redis err"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			ticket, err := authService.GetTicket(ctx, tt.context)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Empty(t, ticket)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, ticket)
			}
		})
	}
}
