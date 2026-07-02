package counter

import (
	counterHttp "github.com/Roisfaozi/queue-base/internal/modules/counter/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/usecase"
	branchRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	serviceRepository "github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type CounterModule struct {
	CounterController *counterHttp.CounterController
	CounterRepo       repository.CounterRepository
	CounterUseCase    usecase.CounterUseCase
}

func NewCounterModule(db *gorm.DB, validate *validator.Validate, branchRepo branchRepository.BranchRepository, branchServiceRepo serviceRepository.BranchServiceRepository, log *logrus.Logger, audit ...usecase.AuditLogger) *CounterModule {
	repo := repository.NewCounterRepository(db)
	uc := usecase.NewCounterUseCase(repo, branchRepo, branchServiceRepo, audit...)
	ctrl := counterHttp.NewCounterController(uc, validate, log)
	return &CounterModule{CounterController: ctrl, CounterRepo: repo, CounterUseCase: uc}
}
