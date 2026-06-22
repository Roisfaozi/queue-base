package service

import (
	serviceHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service/delivery/http"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service/repository"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service/usecase"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ServiceModule struct {
	ServiceController *serviceHttp.ServiceController
	ServiceRepo       repository.ServiceRepository
	ServiceUseCase    usecase.ServiceUseCase
}

func NewServiceModule(db *gorm.DB, validate *validator.Validate) *ServiceModule {
	repo := repository.NewServiceRepository(db)
	uc := usecase.NewServiceUseCase(repo)
	ctrl := serviceHttp.NewServiceController(uc, validate)
	return &ServiceModule{ServiceController: ctrl, ServiceRepo: repo, ServiceUseCase: uc}
}
