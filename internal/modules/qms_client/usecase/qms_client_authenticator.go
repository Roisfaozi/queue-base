package usecase

import (
	"context"
	"errors"

	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/repository"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"gorm.io/gorm"
)

type QMSClientIdentity struct {
	ClientID        string
	TenantID        string
	BranchID        string
	ClientType      string
	BranchServiceID string
	CounterID       string
}

type QMSClientAuthenticator interface {
	Authenticate(ctx context.Context, clientID, apiKey string) (*QMSClientIdentity, error)
}

type qmsClientAuthenticator struct {
	repo repository.QmsClientRepository
}

func NewQMSClientAuthenticator(repo repository.QmsClientRepository) QMSClientAuthenticator {
	return &qmsClientAuthenticator{repo: repo}
}

func (a *qmsClientAuthenticator) Authenticate(ctx context.Context, clientID, apiKey string) (*QMSClientIdentity, error) {
	client, err := a.repo.FindByID(ctx, clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUnauthorized
		}
		return nil, exception.ErrUnauthorized
	}

	cred, err := a.repo.FindCredentialByClientID(ctx, clientID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, exception.ErrUnauthorized
		}
		return nil, exception.ErrUnauthorized
	}

	if !pkg.CheckPasswordHash(apiKey, cred.ClientSecretHash) {
		return nil, exception.ErrUnauthorized
	}

	return &QMSClientIdentity{
		ClientID: client.ID,
		TenantID: client.TenantID,
		BranchID: client.BranchID,
		BranchServiceID: func() string {
			if client.BranchServiceID == nil {
				return ""
			}
			return *client.BranchServiceID
		}(),
		CounterID: func() string {
			if client.CounterID == nil {
				return ""
			}
			return *client.CounterID
		}(),
		ClientType: string(client.ClientType),
	}, nil
}
