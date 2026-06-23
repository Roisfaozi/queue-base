package settings

import (
	settingsHttp "github.com/Roisfaozi/queue-base/internal/modules/settings/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/usecase"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type SettingsModule struct {
	SettingsController    *settingsHttp.SettingsController
	SettingsRepo          repository.SettingsRepository
	SettingsUseCase       usecase.SettingsUseCase
	QueueSettingsResolver *QueueSettingsResolver
}

func NewSettingsModule(db *gorm.DB, validate *validator.Validate) *SettingsModule {
	repo := repository.NewSettingsRepository(db)
	uc := usecase.NewSettingsUseCase(repo)
	ctrl := settingsHttp.NewSettingsController(uc, validate)
	resolver := NewQueueSettingsResolver(uc)
	return &SettingsModule{SettingsController: ctrl, SettingsRepo: repo, SettingsUseCase: uc, QueueSettingsResolver: resolver}
}
