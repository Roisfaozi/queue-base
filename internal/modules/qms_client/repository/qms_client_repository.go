package repository

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	"gorm.io/gorm"
)

type QmsClientRepository interface {
	FindByID(ctx context.Context, id string) (*entity.QMSClient, error)
	FindCredentialByClientID(ctx context.Context, clientID string) (*entity.QMSClientCredential, error)
}

type qmsClientRepository struct {
	db *gorm.DB
}

func NewQmsClientRepository(db *gorm.DB) QmsClientRepository {
	return &qmsClientRepository{db: db}
}

func (r *qmsClientRepository) FindByID(ctx context.Context, id string) (*entity.QMSClient, error) {
	var client entity.QMSClient
	if err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&client).Error; err != nil {
		return nil, err
	}
	return &client, nil
}

func (r *qmsClientRepository) FindCredentialByClientID(ctx context.Context, clientID string) (*entity.QMSClientCredential, error) {
	var cred entity.QMSClientCredential
	if err := r.db.WithContext(ctx).
		Joins("JOIN qms_clients ON qms_clients.id = qms_client_credentials.client_id AND qms_clients.is_active = ?", true).
		Where("qms_client_credentials.client_id = ?", clientID).
		First(&cred).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}
