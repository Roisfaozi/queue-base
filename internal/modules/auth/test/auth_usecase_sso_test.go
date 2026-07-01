package test

import (
	"context"
	"errors"
	"testing"

	mock_auth "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/sso"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

func TestAuthUsecase_GetSSORedirectURL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		category string
		provider string
		state    string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			provider: "github",
			state:    "test-state",
			setup: func(deps *testDependencies) {
				ssoProvider := new(mock_auth.MockSSOProvider)
				ssoProvider.On("GetLoginURL", "test-state").Return("http://sso-login-url")
				deps.ssoProviders["github"] = ssoProvider
			},
		},
		{
			name:     "Negative_ProviderNotFound",
			category: "negative",
			provider: "unknown",
			state:    "test-state",
			setup: func(deps *testDependencies) {
			},
			wantErr: errors.New("unsupported SSO provider"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			url, err := authService.GetSSORedirectURL(ctx, tt.provider, tt.state)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, url)
			}
		})
	}
}

func TestAuthUsecase_HandleSSOCallback(t *testing.T) {
	ctx := context.Background()
	token := &oauth2.Token{AccessToken: "acc", RefreshToken: "ref"}
	userInfo := &sso.UserInfo{Email: "test@example.com", ProviderID: "12345", Name: "Test User"}

	tests := []struct {
		name     string
		category string
		provider string
		code     string
		setup    func(*testDependencies)
		wantErr  error
	}{
		{
			name:     "Negative_ProviderNotFound",
			category: "negative",
			provider: "unknown",
			code:     "code123",
			setup: func(deps *testDependencies) {
			},
			wantErr: errors.New("unsupported SSO provider"),
		},
		{
			name:     "Negative_ExchangeCodeError",
			category: "negative",
			provider: "github",
			code:     "code123",
			setup: func(deps *testDependencies) {
				ssoProvider := new(mock_auth.MockSSOProvider)
				ssoProvider.On("ExchangeCode", mock.Anything, "code123").Return(nil, errors.New("exchange error"))
				deps.ssoProviders["github"] = ssoProvider
			},
			wantErr: errors.New("exchange error"),
		},
		{
			name:     "Negative_GetUserInfoError",
			category: "negative",
			provider: "github",
			code:     "code123",
			setup: func(deps *testDependencies) {
				ssoProvider := new(mock_auth.MockSSOProvider)
				ssoProvider.On("ExchangeCode", mock.Anything, "code123").Return(token, nil)
				ssoProvider.On("GetUserInfo", mock.Anything, token).Return(nil, errors.New("user info err"))
				deps.ssoProviders["github"] = ssoProvider
			},
			wantErr: errors.New("user info err"),
		},
		{
			name:     "Positive_ExistingSSOIdentity",
			category: "positive",
			provider: "github",
			code:     "code123",
			setup: func(deps *testDependencies) {
				ssoProvider := new(mock_auth.MockSSOProvider)
				ssoProvider.On("ExchangeCode", mock.Anything, "code123").Return(token, nil)
				ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil)
				deps.ssoProviders["github"] = ssoProvider

				ssoIdentity := &userEntity.UserSSOIdentity{UserID: TestUserID, Provider: "github", ProviderID: "12345"}
				deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "12345").Return(ssoIdentity, nil)
				user := &userEntity.User{ID: TestUserID, Status: userEntity.UserStatusActive, Email: userInfo.Email}
				deps.userRepo.On("FindByID", mock.Anything, TestUserID).Return(user, nil)

				deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"role:user"}, nil)
				deps.authz.On("GetImplicitPermissionsForUser", mock.Anything, TestUserID, "").Return([][]string{}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "Positive_ExistingEmail_LinkIdentity",
			category: "positive",
			provider: "github",
			code:     "code123",
			setup: func(deps *testDependencies) {
				ssoProvider := new(mock_auth.MockSSOProvider)
				ssoProvider.On("ExchangeCode", mock.Anything, "code123").Return(token, nil)
				ssoProvider.On("GetUserInfo", mock.Anything, token).Return(userInfo, nil)
				deps.ssoProviders["github"] = ssoProvider

				deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "12345").Return(nil, errors.New("not found"))
				user := &userEntity.User{ID: TestUserID, Status: userEntity.UserStatusActive, Email: userInfo.Email}
				deps.userRepo.On("FindByEmail", mock.Anything, userInfo.Email).Return(user, nil)
				deps.userRepo.On("CreateSSOIdentity", mock.Anything, mock.AnythingOfType("*entity.UserSSOIdentity")).Return(nil)

				deps.authz.On("GetRolesForUser", mock.Anything, TestUserID, "").Return([]string{"role:user"}, nil)
				deps.authz.On("GetImplicitPermissionsForUser", mock.Anything, TestUserID, "").Return([][]string{}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		{
			name:     "Positive_NewUser_AutoProvision",
			category: "positive",
			provider: "github",
			code:     "code123",
			setup: func(deps *testDependencies) {
				newUserInfo := &sso.UserInfo{Email: "new@example.com", ProviderID: "12345", Name: "New User"}
				ssoProvider := new(mock_auth.MockSSOProvider)
				ssoProvider.On("ExchangeCode", mock.Anything, "code123").Return(token, nil)
				ssoProvider.On("GetUserInfo", mock.Anything, token).Return(newUserInfo, nil)
				deps.ssoProviders["github"] = ssoProvider

				deps.userRepo.On("FindBySSOIdentity", mock.Anything, "github", "12345").Return(nil, errors.New("not found"))
				deps.userRepo.On("FindByEmail", mock.Anything, newUserInfo.Email).Return(nil, errors.New("not found"))

				simulateWithinTransaction(deps, nil)
				deps.userRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.User")).Return(nil)
				deps.authz.On("AssignDefaultRole", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.orgRepo.On("Create", mock.Anything, mock.AnythingOfType("*entity.Organization"), "owner").Return(nil)
				deps.userRepo.On("CreateSSOIdentity", mock.Anything, mock.AnythingOfType("*entity.UserSSOIdentity")).Return(nil)

				deps.authz.On("GetRolesForUser", mock.Anything, mock.AnythingOfType("string"), "").Return([]string{"role:user"}, nil)
				deps.authz.On("GetImplicitPermissionsForUser", mock.Anything, mock.AnythingOfType("string"), "").Return([][]string{}, nil)
				deps.tokenRepo.On("StoreToken", mock.Anything, mock.AnythingOfType("*model.Auth")).Return(nil)
				deps.tokenRepo.On("ResetLoginAttempts", mock.Anything, mock.AnythingOfType("string")).Return(nil)
				deps.taskDistributor.On("DistributeTaskAuditLog", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			authService, deps := setupTest(t)
			tt.setup(deps)
			res, _, err := authService.HandleSSOCallback(ctx, tt.provider, tt.code)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, res)
			}
		})
	}
}
