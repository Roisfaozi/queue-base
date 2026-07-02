package caller

import (
	callerHttp "github.com/Roisfaozi/queue-base/internal/modules/caller/delivery/http"
	callerUsecase "github.com/Roisfaozi/queue-base/internal/modules/caller/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CallerModule struct {
	CallerController *callerHttp.CallerController
}

func NewCallerModule(db *gorm.DB, qu usecase.QueueUseCase, validate *validator.Validate, log *logrus.Logger) *CallerModule {
	uc := callerUsecase.NewCallerUseCase(db, qu)
	ctrl := callerHttp.NewCallerController(uc, validate, log)
	return &CallerModule{CallerController: ctrl}
}
