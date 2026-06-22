package queue

import (
	queueHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/usecase"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type QueueModule struct {
	QueueController *queueHttp.QueueController
	QueueRepo       repository.QueueRepository
	QueueUseCase    usecase.QueueUseCase
}

func NewQueueModule(db *gorm.DB, validate *validator.Validate) *QueueModule {
	// FIXME: using nil repo for scaffold until real implementation is ready
	var repo repository.QueueRepository
	uc := usecase.NewQueueUseCase(repo)
	ctrl := queueHttp.NewQueueController(uc, validate)
	return &QueueModule{QueueController: ctrl, QueueRepo: repo, QueueUseCase: uc}
}
