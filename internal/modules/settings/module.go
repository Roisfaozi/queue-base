package settings

import (
	settingsHttp "github.com/Roisfaozi/queue-base/internal/modules/settings/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SettingsModule struct {
	SettingsController    *settingsHttp.SettingsController
	SettingsRepo          repository.SettingsRepository
	SettingsUseCase       usecase.SettingsUseCase
	QueueSettingsResolver *QueueSettingsResolver
}

func NewSettingsModule(db *gorm.DB, validate *validator.Validate, log *logrus.Logger, audit ...usecase.AuditLogger) *SettingsModule {
	repo := repository.NewSettingsRepository(db)
	uc := usecase.NewSettingsUseCase(repo, audit...)
	resolver := NewQueueSettingsResolver(db, uc)
	ctrl := settingsHttp.NewSettingsControllerWithResolver(uc, validate, resolver, log)
	return &SettingsModule{SettingsController: ctrl, SettingsRepo: repo, SettingsUseCase: uc, QueueSettingsResolver: resolver}
}
