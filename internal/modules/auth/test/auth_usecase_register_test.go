package test

import (
	"context"
	"errors"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	authEntity "github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/internal/worker/tasks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAuthUsecase_Register(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		category string
		req      *model.RegisterRequest
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			req: &model.RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
			},
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
				deps.userRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found")).Once()
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
				deps.authz.On("AssignDefaultRole", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Organization"), "owner").Return(nil)

				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
					return p.To == "new@example.com" && p.Subject == "Email Verification"
				}), mock.Anything).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "REGISTER" || r.Action == "ORGANIZATION_CREATED"
				}), mock.Anything).Return(nil)

				// After register, it calls Login automatically. We need to mock Login dependencies too.
				deps.tokenRepo.On("IsAccountLocked", mock.Anything, "newuser").Return(false, time.Duration(0), nil)

				userEntity, _ := createTestUser("password123")
				userEntity.Username = "newuser"
				userEntity.Email = "new@example.com"

				deps.userRepo.On("FindByUsername", mock.Anything, "newuser").Return(userEntity, nil).Once()
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, "newuser").Return(nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)

				deps.orgRepo.On("FindByOwnerID", mock.Anything, mock.AnythingOfType("string")).Return([]orgEntity.Organization{}, nil)
				deps.orgRepo.On("FindUserOrganizations", mock.Anything, mock.AnythingOfType("string")).Return([]*orgEntity.Organization{}, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]string{"role:user"}, nil)
				deps.authz.On("GetImplicitPermissionsForUser", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([][]string{}, nil)

				deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.AnythingOfType("model.UserInfo"), mock.AnythingOfType("[]string")).Return(nil)

				// Login also sends a LOGIN audit log
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "LOGIN"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Negative_EmailTaken",
			category: "negative",
			req: &model.RegisterRequest{
				Username: "newuser",
				Email:    "taken@example.com",
				Password: "password123",
				Name:     "New User",
			},
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found"))
				deps.userRepo.On("FindByEmail", mock.Anything, "taken@example.com").Return(&entity.User{}, nil)
			},
			wantErr: errors.New("email already exists"),
		},
		{
			name:     "Negative_UsernameTaken",
			category: "negative",
			req: &model.RegisterRequest{
				Username: "takenuser",
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
			},
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByUsername", mock.Anything, "takenuser").Return(&entity.User{}, nil)
			},
			wantErr: errors.New("username already exists"),
		},
		{
			name:     "Negative_TransactionError",
			category: "negative",
			req: &model.RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
			},
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found"))
				deps.userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
				dbErr := errors.New("db err")
				simulateWithinTransaction(deps, dbErr)
				deps.userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(dbErr)
				deps.authz.On("AssignDefaultRole", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Organization"), "owner").Return(nil)
			},
			wantErr: errors.New("db err"),
		},
		{
			name:     "Positive_AuditError_ShouldNotBlock",
			category: "positive",
			req: &model.RegisterRequest{
				Username: "newuser",
				Email:    "new@example.com",
				Password: "password123",
				Name:     "New User",
			},
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByEmail", mock.Anything, "new@example.com").Return(nil, errors.New("not found"))
				deps.userRepo.On("FindByUsername", mock.Anything, "newuser").Return(nil, errors.New("not found")).Once()
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
				deps.authz.On("AssignDefaultRole", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Organization"), "owner").Return(nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "REGISTER" || r.Action == "ORGANIZATION_CREATED"
				}), mock.Anything).Return(errors.New("audit error"))

				deps.tokenRepo.On("IsAccountLocked", mock.Anything, "newuser").Return(false, time.Duration(0), nil)
				userEntity, _ := createTestUser("password123")
				userEntity.Username = "newuser"
				userEntity.Email = "new@example.com"
				deps.userRepo.On("FindByUsername", mock.Anything, "newuser").Return(userEntity, nil).Once()
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, "newuser").Return(nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.orgRepo.On("FindByOwnerID", mock.Anything, mock.AnythingOfType("string")).Return([]orgEntity.Organization{}, nil)
				deps.orgRepo.On("FindUserOrganizations", mock.Anything, mock.AnythingOfType("string")).Return([]*orgEntity.Organization{}, nil)
				deps.authz.On("GetRolesForUser", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([]string{"role:user"}, nil)
				deps.authz.On("GetImplicitPermissionsForUser", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return([][]string{}, nil)
				deps.publisher.On("PublishUserLoggedIn", mock.Anything, mock.AnythingOfType("model.UserInfo"), mock.AnythingOfType("[]string")).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "LOGIN"
				}), mock.Anything).Return(errors.New("audit error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			_, _, err := authService.Register(ctx, *tt.req)
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

func TestAuthUsecase_VerifyEmail(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil
	token := "valid-verify-token"

	tests := []struct {
		name     string
		category string
		token    string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			token:    token,
			setup: func(deps *testDependencies) {
				freshUser, _ := createTestUser("password123")
				freshUser.EmailVerifiedAt = nil
				verifyToken := &authEntity.EmailVerificationToken{Email: freshUser.Email, Token: token, ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
				deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verifyToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, freshUser.Email).Return(freshUser, nil)
				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
				deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, freshUser.Email).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "EMAIL_VERIFIED"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Negative_InvalidToken",
			category: "negative",
			token:    "invalid",
			setup: func(deps *testDependencies) {
				deps.tokenRepo.On("FindVerificationToken", mock.Anything, "invalid").Return(nil, errors.New("not found"))
			},
			wantErr: usecase.ErrInvalidVerificationToken,
		},
		{
			name:     "Negative_ExpiredToken",
			category: "negative",
			token:    "expired",
			setup: func(deps *testDependencies) {
				verifyToken := &authEntity.EmailVerificationToken{Email: user.Email, Token: "expired", ExpiresAt: time.Now().Add(-1 * time.Hour).UnixMilli()}
				deps.tokenRepo.On("FindVerificationToken", mock.Anything, "expired").Return(verifyToken, nil)
				deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil)
			},
			wantErr: usecase.ErrInvalidVerificationToken,
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			token:    token,
			setup: func(deps *testDependencies) {
				verifyToken := &authEntity.EmailVerificationToken{Email: "missing@example.com", Token: token, ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
				deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verifyToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, "missing@example.com").Return(nil, errors.New("not found"))
			},
			wantErr: usecase.ErrInvalidVerificationToken,
		},
		{
			name:     "Negative_AlreadyVerified",
			category: "negative",
			token:    token,
			setup: func(deps *testDependencies) {
				verifyToken := &authEntity.EmailVerificationToken{Email: user.Email, Token: token, ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
				deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verifyToken, nil)
				verifiedUser := &entity.User{ID: user.ID, Email: user.Email, EmailVerifiedAt: new(int64)}
				deps.userRepo.On("FindByEmail", mock.Anything, user.Email).Return(verifiedUser, nil)
				deps.tokenRepo.On("DeleteVerificationTokenByEmail", mock.Anything, user.Email).Return(nil)
			},
			wantErr: usecase.ErrAlreadyVerified,
		},
		{
			name:     "Negative_TransactionError",
			category: "negative",
			token:    token,
			setup: func(deps *testDependencies) {
				freshUser, _ := createTestUser("password123")
				freshUser.EmailVerifiedAt = nil
				verifyToken := &authEntity.EmailVerificationToken{Email: freshUser.Email, Token: token, ExpiresAt: time.Now().Add(1 * time.Hour).UnixMilli()}
				deps.tokenRepo.On("FindVerificationToken", mock.Anything, token).Return(verifyToken, nil)
				deps.userRepo.On("FindByEmail", mock.Anything, freshUser.Email).Return(freshUser, nil)
				dbErr := errors.New("update error")
				simulateWithinTransaction(deps, dbErr)
				deps.userRepo.On("Update", mock.Anything, mock.AnythingOfType("*entity.User")).Return(dbErr)
			},
			wantErr: errors.New("update error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.VerifyEmail(ctx, tt.token)
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

func TestAuthUsecase_RequestVerification(t *testing.T) {
	ctx := context.Background()
	user, _ := createTestUser("password123")
	user.EmailVerifiedAt = nil

	tests := []struct {
		name     string
		category string
		userID   string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			userID:   user.ID,
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.MatchedBy(func(p *tasks.SendEmailPayload) bool {
					return p.To == user.Email
				}), mock.Anything).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "VERIFICATION_EMAIL_REQUESTED"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Negative_UserNotFound",
			category: "negative",
			userID:   "missing-id",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, "missing-id").Return(nil, errors.New("not found"))
			},
			wantErr: errors.New("not found"),
		},
		{
			name:     "Negative_AlreadyVerified",
			category: "negative",
			userID:   user.ID,
			setup: func(deps *testDependencies) {
				verifiedUser := &entity.User{ID: user.ID, Email: user.Email, EmailVerifiedAt: new(int64)}
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(verifiedUser, nil)
			},
			wantErr: usecase.ErrAlreadyVerified,
		},
		{
			name:     "Negative_SaveTokenError",
			category: "negative",
			userID:   user.ID,
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, user.ID).Return(user, nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(errors.New("db error"))
			},
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			err := authService.RequestVerification(ctx, tt.userID)
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAuthUsecase_ResendVerificationEmail(t *testing.T) {
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
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.MatchedBy(func(r auditModel.CreateAuditLogRequest) bool {
					return r.Action == "VERIFICATION_EMAIL_REQUESTED"
				}), mock.Anything).Return(nil)
			},
		},
		{
			name:     "Negative_DistributeError",
			category: "negative",
			setup: func(deps *testDependencies) {
				deps.userRepo.On("FindByID", mock.Anything, "test-user-id").Return(user, nil)
				deps.tokenRepo.On("SaveVerificationToken", mock.Anything, mock.AnythingOfType("*entity.EmailVerificationToken")).Return(nil)
				deps.taskDistributor.On("DistributeTaskSendEmail", mock.Anything, mock.Anything, mock.Anything).Return(errors.New("dist error"))
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
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
