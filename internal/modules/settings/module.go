package settings

import (
	auditUsecase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	settingsHttp "github.com/Roisfaozi/queue-base/internal/modules/settings/delivery/http"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SettingsModule struct {
	SettingsController    *settingsHttp.SettingsController
	QueueSettingsResolver *QueueSettingsResolver
}

func NewSettingsModule(db *gorm.DB, validate *validator.Validate, log *logrus.Logger, audit ...auditUsecase.AuditUseCase) *SettingsModule {
	resolver := NewQueueSettingsResolver(db)
	var auditUC auditUsecase.AuditUseCase
	if len(audit) > 0 {
		auditUC = audit[0]
	}
	ctrl := settingsHttp.NewSettingsController(validate, resolver, log, db, auditUC)
	return &SettingsModule{SettingsController: ctrl, QueueSettingsResolver: resolver}
}
