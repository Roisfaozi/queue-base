package service

import (
	branchRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	serviceHttp "github.com/Roisfaozi/queue-base/internal/modules/service/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/service/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/service/usecase"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ServiceModule struct {
	ServiceController       *serviceHttp.ServiceController
	BranchServiceController *serviceHttp.BranchServiceController
	ServiceRepo             repository.ServiceRepository
	BranchServiceRepo       repository.BranchServiceRepository
	ServiceUseCase          usecase.ServiceUseCase
	BranchServiceUseCase    usecase.BranchServiceUseCase
}

func NewServiceModule(db *gorm.DB, validate *validator.Validate, branchRepo branchRepository.BranchRepository) *ServiceModule {
	repo := repository.NewServiceRepository(db)
	branchServiceRepo := repository.NewBranchServiceRepository(db)
	uc := usecase.NewServiceUseCase(repo)
	branchServiceUseCase := usecase.NewBranchServiceUseCase(branchServiceRepo, repo, branchRepo)
	ctrl := serviceHttp.NewServiceController(uc, validate)
	branchServiceCtrl := serviceHttp.NewBranchServiceController(branchServiceUseCase, validate)
	return &ServiceModule{ServiceController: ctrl, BranchServiceController: branchServiceCtrl, ServiceRepo: repo, BranchServiceRepo: branchServiceRepo, ServiceUseCase: uc, BranchServiceUseCase: branchServiceUseCase}
}
