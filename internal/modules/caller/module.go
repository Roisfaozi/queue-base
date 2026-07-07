package caller

import (
	auditUsecase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	authUsecase "github.com/Roisfaozi/queue-base/internal/modules/auth/usecase"
	callerHttp "github.com/Roisfaozi/queue-base/internal/modules/caller/delivery/http"
	callerUsecase "github.com/Roisfaozi/queue-base/internal/modules/caller/usecase"
	queueUsecase "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CallerModule struct {
	CallerController *callerHttp.CallerController
}

func NewCallerModule(db *gorm.DB, qu queueUsecase.QueueUseCase, authUC authUsecase.AuthUseCase, auditUC auditUsecase.AuditUseCase, validate *validator.Validate, log *logrus.Logger) *CallerModule {
	uc := callerUsecase.NewCallerUseCase(db, qu, authUC, auditUC)
	ctrl := callerHttp.NewCallerController(uc, validate, log)
	return &CallerModule{CallerController: ctrl}
}
