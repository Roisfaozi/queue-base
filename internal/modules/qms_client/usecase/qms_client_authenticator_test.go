package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/usecase"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestQMSClientAuthenticator(t *testing.T) {
	hash, _ := pkg.HashPassword("secret123")

	now := time.Now().UnixMilli()
	past := now - 10000
	future := now + 10000

	tests := []struct {
		name        string
		category    string
		clientID    string
		apiKey      string
		setup       func(t *testing.T) *mocks.MockQmsClientRepository
		expectedErr error
		expectedID  *usecase.QMSClientIdentity
	}{
		{
			name:     "Negative_ClientNotFound",
			category: "negative",
			clientID: "c-1",
			apiKey:   "secret123",
			setup: func(t *testing.T) *mocks.MockQmsClientRepository {
				m := mocks.NewMockQmsClientRepository(t)
				m.On("FindByID", mock.Anything, "c-1").Return(nil, gorm.ErrRecordNotFound).Once()
				return m
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "Negative_CredentialNotFound",
			category: "negative",
			clientID: "c-1",
			apiKey:   "secret123",
			setup: func(t *testing.T) *mocks.MockQmsClientRepository {
				m := mocks.NewMockQmsClientRepository(t)
				m.On("FindByID", mock.Anything, "c-1").Return(&entity.QMSClient{ID: "c-1"}, nil).Once()
				m.On("FindCredentialByClientID", mock.Anything, "c-1").Return(nil, gorm.ErrRecordNotFound).Once()
				return m
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "Vulnerability_SecretMismatch",
			category: "vulnerability",
			clientID: "c-1",
			apiKey:   "wrongsecret",
			setup: func(t *testing.T) *mocks.MockQmsClientRepository {
				m := mocks.NewMockQmsClientRepository(t)
				m.On("FindByID", mock.Anything, "c-1").Return(&entity.QMSClient{ID: "c-1"}, nil).Once()
				m.On("FindCredentialByClientID", mock.Anything, "c-1").Return(&entity.QMSClientCredential{ClientSecretHash: hash}, nil).Once()
				return m
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "Edge_CredentialExpired",
			category: "edge",
			clientID: "c-1",
			apiKey:   "secret123",
			setup: func(t *testing.T) *mocks.MockQmsClientRepository {
				m := mocks.NewMockQmsClientRepository(t)
				m.On("FindByID", mock.Anything, "c-1").Return(&entity.QMSClient{ID: "c-1"}, nil).Once()
				m.On("FindCredentialByClientID", mock.Anything, "c-1").Return(&entity.QMSClientCredential{ClientSecretHash: hash, ExpiresAt: &past}, nil).Once()
				return m
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "Positive_ValidWithoutExpiry",
			category: "positive",
			clientID: "c-1",
			apiKey:   "secret123",
			setup: func(t *testing.T) *mocks.MockQmsClientRepository {
				m := mocks.NewMockQmsClientRepository(t)
				m.On("FindByID", mock.Anything, "c-1").Return(&entity.QMSClient{
					ID:         "c-1",
					TenantID:   "t-1",
					BranchID:   "b-1",
					ClientType: "signage",
				}, nil).Once()
				m.On("FindCredentialByClientID", mock.Anything, "c-1").Return(&entity.QMSClientCredential{ClientSecretHash: hash}, nil).Once()
				return m
			},
			expectedID: &usecase.QMSClientIdentity{
				ClientID:   "c-1",
				TenantID:   "t-1",
				BranchID:   "b-1",
				ClientType: "signage",
			},
		},
		{
			name:     "Positive_ValidWithFutureExpiry",
			category: "positive",
			clientID: "c-1",
			apiKey:   "secret123",
			setup: func(t *testing.T) *mocks.MockQmsClientRepository {
				m := mocks.NewMockQmsClientRepository(t)
				m.On("FindByID", mock.Anything, "c-1").Return(&entity.QMSClient{
					ID:         "c-1",
					TenantID:   "t-1",
					BranchID:   "b-1",
					ClientType: "signage",
				}, nil).Once()
				m.On("FindCredentialByClientID", mock.Anything, "c-1").Return(&entity.QMSClientCredential{ClientSecretHash: hash, ExpiresAt: &future}, nil).Once()
				return m
			},
			expectedID: &usecase.QMSClientIdentity{
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
			repo := tt.setup(t)
			auth := usecase.NewQMSClientAuthenticator(repo)
			res, err := auth.Authenticate(context.Background(), tt.clientID, tt.apiKey)

			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				assert.Nil(t, res)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.expectedID, res)
			repo.AssertExpectations(t)
		})
	}
}
