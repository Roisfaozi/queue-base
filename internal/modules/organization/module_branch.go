package organization

import (
	"github.com/Roisfaozi/queue-base/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BranchModule struct {
	BranchController *http.BranchController
	BranchRepo       repository.BranchRepository
	BranchUseCase    usecase.BranchUseCase
}

func NewBranchModule(db *gorm.DB, validate *validator.Validate,
	log *logrus.Logger,
) *BranchModule {
	repo := repository.NewBranchRepository(db)
	uc := usecase.NewBranchUseCase(repo)
	ctrl := http.NewBranchController(uc, validate, log)
	return &BranchModule{BranchController: ctrl, BranchRepo: repo, BranchUseCase: uc}
}
