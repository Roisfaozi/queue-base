package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	authEntity "github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	"github.com/Roisfaozi/queue-base/internal/worker/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthUsecase_ForgotPassword(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")

	tests := []struct {
		name     string
		category string
		email    string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			email:    user.Email,
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
					return p.To == user.Email && p.Subject == "Password Reset Request"
				}), mock.Anything).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == user.ID && r.Action == "FORGOT_PASSWORD_REQUEST"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Positive_UserNotFound_Security_EnumPrevention",
			category: "positive",
			email:    "notfound@example.com",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, "notfound@example.com").Return(nil, errors.New("user not found"))
			},
		},
		{
			name:     "Negative_RepositoryError",
			category: "negative",
			email:    user.Email,
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(errors.New("db save error"))
			},
		},
		{
			name:     "Positive_DistributeTaskError_ShouldNotBlock",
			category: "positive",
			email:    user.Email,
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("task error"))
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "Positive_AuditError_ShouldNotBlock",
			category: "positive",
			email:    user.Email,
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				deps.tokenRepo.On("Save", mock.Anything, mock.AnythingOfType("*entity.PasswordResetToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.ForgotPassword(ctx, tt.email)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUsecase_ResetPassword(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")
	token := "valid-token"

	tests := []struct {
		name        string
		category    string
		token       string
		newPassword string
		setup       func(*testDependencies)
		wantErr     error
	}{
		{
			name:        "Positive_Success",
			category:    "positive",
			token:       token,
			newPassword: "new-strong-password-123",
			setup: func(deps *testDependencies) {
				resetToken := &authEntity.PasswordResetToken{Email: user.Email, Token: token, ExpiresAt: time.Now().Add(1 * time.Hour)}
				deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				simulateWithinTransaction(deps, nil)
				deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)
				deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
				deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.UserID == user.ID && r.Action == "PASSWORD_RESET_SUCCESS"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:        "Negative_InvalidToken",
			category:    "negative",
			token:       "invalid-token",
			newPassword: "new",
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("FindByToken", mock.Anything, "invalid-token").Return(nil, errors.New("not found"))
			},
			wantErr: usecase.ErrInvalidResetToken,
		},
		{
			name:        "Negative_ExpiredToken",
			category:    "negative",
			token:       "expired",
			newPassword: "new",
			setup: func(deps *testDependencies) {
				resetToken := &authEntity.PasswordResetToken{Email: "test@example.com", Token: "expired", ExpiresAt: time.Now().Add(-1 * time.Hour)}
				deps.tokenRepo.On("FindByToken", mock.Anything, "expired").Return(resetToken, nil)
				deps.tokenRepo.On("DeleteByEmail", mock.Anything, resetToken.Email).Return(nil)
			},
			wantErr: usecase.ErrInvalidResetToken,
		},
		{
			name:        "Negative_UserDeleted",
			category:    "negative",
			token:       token,
			newPassword: "new",
			setup: func(deps *testDependencies) {
				resetToken := &authEntity.PasswordResetToken{Email: "deleted@example.com", Token: token, ExpiresAt: time.Now().Add(1 * time.Hour)}
				deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, resetToken.Email).Return(nil, errors.New("not found"))
			},
			wantErr: usecase.ErrInvalidResetToken,
		},
		{
			name:        "Negative_TransactionError",
			category:    "negative",
			token:       token,
			newPassword: "new-strong-password-123",
			setup: func(deps *testDependencies) {
				resetToken := &authEntity.PasswordResetToken{Email: user.Email, Token: token, ExpiresAt: time.Now().Add(1 * time.Hour)}
				deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				dbErr := errors.New("update failed")
				simulateWithinTransaction(deps, dbErr)
				deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)
				deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(dbErr)
			},
			wantErr: errors.New("update failed"),
		},
		{
			name:        "Positive_AuditError_ShouldNotBlock",
			category:    "positive",
			token:       token,
			newPassword: "new-strong-password-123",
			setup: func(deps *testDependencies) {
				resetToken := &authEntity.PasswordResetToken{Email: user.Email, Token: token, ExpiresAt: time.Now().Add(1 * time.Hour)}
				deps.tokenRepo.On("FindByToken", mock.Anything, token).Return(resetToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(user, nil)
				simulateWithinTransaction(deps, nil)
				deps.tokenRepo.On("RevokeAllSessions", mock.Anything, user.ID).Return(nil)
				deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
				deps.tokenRepo.On("DeleteByEmail", mock.Anything, user.Email).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("audit error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.ResetPassword(ctx, tt.token, tt.newPassword)
			if tt.wantErr != nil {
				assert.Error(t, err)
				if !errors.Is(err, tt.wantErr) {
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUsecase_RequestVerificationEmail(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil

	tests := []struct {
		name     string
		category string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, "test-user-id").Return(user, nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
					return p.To == user.Email
				}), mock.Anything).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, "test-user-id").Return(nil, errors.New("user not found"))
			},
			wantErr: errors.New("user not found"),
		},
		{
			name:     "Negative_AlreadyVerified",
			category: "negative",
			setup: func(deps *testDependencies) {
				verifiedUser, _ := createTestUser("pass")
				now := time.Now().UnixMilli()
				verifiedUser.EmailVerifiedAt = &now
				deps.userRepo.On("FindByID", mock.Anything, "test-user-id").Return(verifiedUser, nil)
			},
			wantErr: usecase.ErrAlreadyVerified,
		},
		{
			name:     "Negative_SaveTokenError",
			category: "negative",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, "test-user-id").Return(user, nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(errors.New("db err"))
			},
			wantErr: errors.New("db err"),
		},
		{
			name:     "Negative_DistributeError",
			category: "negative",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, "test-user-id").Return(user, nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("task err"))
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.RequestVerification(ctx, "test-user-id")
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
