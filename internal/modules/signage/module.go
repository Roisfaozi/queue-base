package signage

import (
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	signageHttp "github.com/Roisfaozi/queue-base/internal/modules/signage/delivery/http"
	signageUsecase "github.com/Roisfaozi/queue-base/internal/modules/signage/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type SignageModule struct {
	SignageController *signageHttp.SignageController
}

func NewSignageModule(db *gorm.DB, qu queueUsecase.QueueUseCase, validate *validator.Validate, log *logrus.Logger) *SignageModule {
	uc := signageUsecase.NewSignageUseCase(db, qu, log)
	ctrl := signageHttp.NewSignageController(uc, validate, log)
	return &SignageModule{SignageController: ctrl}
}
