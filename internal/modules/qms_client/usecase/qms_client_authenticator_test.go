package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Fake repo for standalone unit testing without db scaffold
type stubRepo struct {
	client *entity.QMSClient
	cred   *entity.QMSClientCredential
	err    error
}

func (r *stubRepo) FindByID(ctx context.Context, id string) (*entity.QMSClient, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.client == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.client, nil
}

func (r *stubRepo) FindCredentialByClientID(ctx context.Context, clientID string) (*entity.QMSClientCredential, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.cred == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.cred, nil
}

func TestQMSClientAuthenticator(t *testing.T) {
	hash, _ := pkg.HashPassword("secret123")

	now := time.Now().UnixMilli()
	past := now - 10000
	future := now + 10000

	tests := []struct {
		name        string
		clientID    string
		apiKey      string
		repo        *stubRepo
		expectedErr error
		expectedID  *QMSClientIdentity
	}{
		{
			name:        "returns unauthorized when client not found",
			clientID:    "c-1",
			apiKey:      "secret123",
			repo:        &stubRepo{err: gorm.ErrRecordNotFound},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "returns unauthorized when credential not found",
			clientID: "c-1",
			apiKey:   "secret123",
			repo: &stubRepo{
				client: &entity.QMSClient{ID: "c-1"},
				cred:   nil,
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "returns unauthorized when secret mismatch",
			clientID: "c-1",
			apiKey:   "wrongsecret",
			repo: &stubRepo{
				client: &entity.QMSClient{ID: "c-1"},
				cred:   &entity.QMSClientCredential{ClientSecretHash: hash},
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "returns unauthorized when credential expired",
			clientID: "c-1",
			apiKey:   "secret123",
			repo: &stubRepo{
				client: &entity.QMSClient{ID: "c-1"},
				cred:   &entity.QMSClientCredential{ClientSecretHash: hash, ExpiresAt: &past},
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "returns full identity when valid and no expiry",
			clientID: "c-1",
			apiKey:   "secret123",
			repo: &stubRepo{
				client: &entity.QMSClient{
					ID:         "c-1",
					TenantID:   "t-1",
					BranchID:   "b-1",
					ClientType: "signage",
				},
				cred: &entity.QMSClientCredential{ClientSecretHash: hash},
			},
			expectedID: &QMSClientIdentity{
				ClientID:   "c-1",
				TenantID:   "t-1",
				BranchID:   "b-1",
				ClientType: "signage",
			},
		},
		{
			name:     "returns full identity when valid and expiry in future",
			clientID: "c-1",
			apiKey:   "secret123",
			repo: &stubRepo{
				client: &entity.QMSClient{
					ID:         "c-1",
					TenantID:   "t-1",
					BranchID:   "b-1",
					ClientType: "signage",
				},
				cred: &entity.QMSClientCredential{ClientSecretHash: hash, ExpiresAt: &future},
			},
			expectedID: &QMSClientIdentity{
				ClientID:   "c-1",
				TenantID:   "t-1",
				BranchID:   "b-1",
				ClientType: "signage",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			auth := NewQMSClientAuthenticator(tt.repo)
			res, err := auth.Authenticate(context.Background(), tt.clientID, tt.apiKey)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.expectedID, res)
		})
	}
}
