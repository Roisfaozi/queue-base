package usecase

import (
	"context"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/signage/model"
	"gorm.io/gorm"
)

type SignageUseCase interface {
	GetMe(ctx context.Context, clientID string) (*model.SignageMeResponse, error)
	GetCurrentCalls(ctx context.Context, clientID string) ([]model.SignageCurrentCallResponse, error)
	GetQueues(ctx context.Context, clientID string) ([]queueModel.QueueResponse, error)
}

type signageUseCase struct {
	db      *gorm.DB
	queueUC queueUsecase.QueueUseCase
}

func NewSignageUseCase(db *gorm.DB, qu queueUsecase.QueueUseCase) SignageUseCase {
	return &signageUseCase{db: db, queueUC: qu}
}

func (u *signageUseCase) GetMe(ctx context.Context, clientID string) (*model.SignageMeResponse, error) {
	// MVP mock implementation until qms_clients auth middleware is fully active
	return &model.SignageMeResponse{
		ClientID:   clientID,
		ClientType: "signage",
	}, nil
}

func (u *signageUseCase) GetCurrentCalls(ctx context.Context, clientID string) ([]model.SignageCurrentCallResponse, error) {
	return []model.SignageCurrentCallResponse{}, nil
}

func (u *signageUseCase) GetQueues(ctx context.Context, clientID string) ([]queueModel.QueueResponse, error) {
	return []queueModel.QueueResponse{}, nil
}
