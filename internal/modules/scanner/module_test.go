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

func TestAPIKeyAuthenticator_Success(t *testing.T) {
	auth := NewAPIKeyAuthenticator(stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}})
	err := auth.Authenticate(context.Background(), "t-1", "b-1", "client-1", "sk_live_key")
	assert.NoError(t, err)
}

func TestAPIKeyAuthenticator_NegativeWrongTenant(t *testing.T) {
	auth := NewAPIKeyAuthenticator(stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "other", UserID: "client-1"}})
	err := auth.Authenticate(context.Background(), "t-1", "b-1", "client-1", "sk_live_key")
	assert.ErrorIs(t, err, exception.ErrUnauthorized)
}

func TestAPIKeyAuthenticator_EdgeEmptyClientIDAllowed(t *testing.T) {
	auth := NewAPIKeyAuthenticator(stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}})
	err := auth.Authenticate(context.Background(), "t-1", "b-1", "", "sk_live_key")
	assert.NoError(t, err)
}

func TestAPIKeyAuthenticator_SecurityWrongClientRejected(t *testing.T) {
	auth := NewAPIKeyAuthenticator(stubAPIKeyUseCase{identity: &apiKeyModel.ApiKeyIdentity{OrganizationID: "t-1", UserID: "client-1"}})
	err := auth.Authenticate(context.Background(), "t-1", "b-1", "client-2", "sk_live_key")
	assert.ErrorIs(t, err, exception.ErrUnauthorized)
}
