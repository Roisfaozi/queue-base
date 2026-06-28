package scanner

import (
	"context"
	"testing"

	apiKeyModel "github.com/Roisfaozi/queue-base/internal/modules/api_key/model"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
)

type stubAPIKeyUseCase struct {
	identity *apiKeyModel.ApiKeyIdentity
	err      error
}

func (s stubAPIKeyUseCase) Authenticate(ctx context.Context, key string) (*apiKeyModel.ApiKeyIdentity, error) {
	return s.identity, s.err
}

func TestAPIKeyAuthenticator_Authenticate(t *testing.T) {
	tests := []struct {
		name     string
		category string
		tenantID string
		branchID string
		clientID string
		apiKey   string
		setup    func() stubAPIKeyUseCase
		wantErr  error
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			tenantID: "t-1",
			branchID: "b-1",
			clientID: "client-1",
			apiKey:   "sk_live_key",
			setup: func() stubAPIKeyUseCase {
				return stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}}
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
			setup: func() stubAPIKeyUseCase {
				return stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "other", UserID: "client-1"}}
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
			setup: func() stubAPIKeyUseCase {
				return stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}}
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
			setup: func() stubAPIKeyUseCase {
				return stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}}
			},
			wantErr: exception.ErrUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := NewAPIKeyAuthenticator(tt.setup())
			err := auth.Authenticate(context.Background(), tt.tenantID, tt.branchID, tt.clientID, tt.apiKey)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
