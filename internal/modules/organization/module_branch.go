package organization

import (
	auditUseCase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/Roisfaozi/queue-base/pkg/storage"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type BranchModule struct {
	BranchController *http.BranchController
	BranchRepo       repository.BranchRepository
	BranchUseCase    usecase.BranchUseCase
	LogoController   *http.LogoController
}

// EnableLogoUpload wires the direct logo upload endpoints using the given storage provider.
func (m *BranchModule) EnableLogoUpload(provider storage.Provider, log *logrus.Logger) {
	m.LogoController = http.NewLogoController(usecase.NewLogoUseCase(provider, log), log)
}

func NewBranchModule(db *gorm.DB, validate *validator.Validate,
	log *logrus.Logger,
	auditUC ...auditUseCase.AuditUseCase,
) *BranchModule {
	repo := repository.NewBranchRepository(db)
	var auditLogger usecase.AuditLogger
	if len(auditUC) > 0 {
		auditLogger = auditUC[0]
	}
	uc := usecase.NewBranchUseCase(repo, auditLogger)
	ctrl := http.NewBranchController(uc, validate, log)
	return &BranchModule{BranchController: ctrl, BranchRepo: repo, BranchUseCase: uc}
}
