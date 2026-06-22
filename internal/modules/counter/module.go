package counter

import (
	counterHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/counter/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/counter/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/counter/usecase"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type CounterModule struct {
	CounterController *counterHttp.CounterController
	CounterRepo       repository.CounterRepository
	CounterUseCase    usecase.CounterUseCase
}

func NewCounterModule(db *gorm.DB, validate *validator.Validate) *CounterModule {
	repo := repository.NewCounterRepository(db)
	uc := usecase.NewCounterUseCase(repo)
	ctrl := counterHttp.NewCounterController(uc, validate)
	return &CounterModule{CounterController: ctrl, CounterRepo: repo, CounterUseCase: uc}
}
