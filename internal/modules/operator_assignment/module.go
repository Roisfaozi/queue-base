package operator_assignment

import (
	assignmentHttp "github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Module struct {
	UseCase    usecase.OperatorAssignmentUseCase
	Controller *assignmentHttp.Controller
}

func NewModule(db *gorm.DB, validate *validator.Validate, log *logrus.Logger, audit ...usecase.AuditLogger) *Module {
	uc := usecase.NewOperatorAssignmentUseCase(db, log, audit...)
	return &Module{UseCase: uc, Controller: assignmentHttp.NewController(uc, validate, log)}
}
