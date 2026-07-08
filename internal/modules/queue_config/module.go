package queue_config

import (
	auditUsecase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	queueConfigHttp "github.com/Roisfaozi/queue-base/internal/modules/queue_config/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/queue_config/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/queue_config/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QueueConfigModule struct {
	QueueConfigController *queueConfigHttp.QueueConfigController
	QueueConfigResolver   *QueueConfigResolver
}

func NewQueueConfigModule(db *gorm.DB, validate *validator.Validate, log *logrus.Logger, audit ...auditUsecase.AuditUseCase) *QueueConfigModule {
	resolver := NewQueueConfigResolver(db)
	repo := repository.NewQueueConfigRepository(db)
	var auditUC auditUsecase.AuditUseCase
	if len(audit) > 0 {
		auditUC = audit[0]
	}
	uc := usecase.NewQueueConfigUseCase(repo, auditUC)
	ctrl := queueConfigHttp.NewQueueConfigController(validate, resolver, log, db, uc, auditUC)
	return &QueueConfigModule{QueueConfigController: ctrl, QueueConfigResolver: resolver}
}
