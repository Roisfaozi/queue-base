package settings

import (
	settingsHttp "github.com/Roisfaozi/queue-base/internal/modules/settings/delivery/http"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SettingsModule struct {
	SettingsController    *settingsHttp.SettingsController
	QueueSettingsResolver *QueueSettingsResolver
}

func NewSettingsModule(db *gorm.DB, validate *validator.Validate, log *logrus.Logger) *SettingsModule {
	resolver := NewQueueSettingsResolver(db)
	ctrl := settingsHttp.NewSettingsController(validate, resolver, log, db)
	return &SettingsModule{SettingsController: ctrl, QueueSettingsResolver: resolver}
}
