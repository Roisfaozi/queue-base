package scanner

import (
	"context"
	"testing"

	apiKeyModel "github.com/Roisfaozi/queue-base/internal/modules/api_key/model"
	apiKeyMocks "github.com/Roisfaozi/queue-base/internal/modules/api_key/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAPIKeyAuthenticator_Authenticate(t *testing.T) {
	tests := []struct {
		name     string
		category string
		tenantID string
		branchID string
		clientID string
		apiKey   string
		setup    func(t *testing.T) *apiKeyMocks.MockApiKeyUseCase
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			tenantID: "t-1",
			branchID: "b-1",
			clientID: "client-1",
			apiKey:   "sk_live_key",
			setup: func(t *testing.T) *apiKeyMocks.MockApiKeyUseCase {
				m := apiKeyMocks.NewMockApiKeyUseCase(t)
				m.On("Authenticate", mock.Anything, "sk_live_key").
					Return(&apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}, nil).Once()
				return m
			},
			wantErr: nil,
		},
		{
			name:     "Negative_WrongTenant",
			category: "negative",
			tenantID: "t-1",
			branchID: "b-1",
			clientID: "client-1",
			apiKey:   "sk_live_key",
			setup: func(t *testing.T) *apiKeyMocks.MockApiKeyUseCase {
				m := apiKeyMocks.NewMockApiKeyUseCase(t)
				m.On("Authenticate", mock.Anything, "sk_live_key").
					Return(&apiKeyModel.ApiKeyIdentity{OrganizationID: "other", UserID: "client-1"}, nil).Once()
				return m
			},
			wantErr: exception.ErrUnauthorized,
		},
		{
			name:     "Edge_EmptyClientIDAllowed",
			category: "edge",
			tenantID: "t-1",
			branchID: "b-1",
			clientID: "",
			apiKey:   "sk_live_key",
			setup: func(t *testing.T) *apiKeyMocks.MockApiKeyUseCase {
				m := apiKeyMocks.NewMockApiKeyUseCase(t)
				m.On("Authenticate", mock.Anything, "sk_live_key").
					Return(&apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}, nil).Once()
				return m
			},
			wantErr: nil,
		},
		{
			name:     "Security_WrongClientRejected",
			category: "vulnerability",
			tenantID: "t-1",
			branchID: "b-1",
			clientID: "client-2",
			apiKey:   "sk_live_key",
			setup: func(t *testing.T) *apiKeyMocks.MockApiKeyUseCase {
				m := apiKeyMocks.NewMockApiKeyUseCase(t)
				m.On("Authenticate", mock.Anything, "sk_live_key").
					Return(&apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}, nil).Once()
				return m
			},
			wantErr: exception.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAPIKeyAuthenticator(tt.setup(t))
			err := auth.Authenticate(context.Background(), tt.tenantID, tt.branchID, tt.clientID, tt.apiKey)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
