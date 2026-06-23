package project

import (
	"github.com/Roisfaozi/queue-base/internal/modules/project/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/project/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/project/usecase"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type ProjectModule struct {
	ProjectController *http.ProjectController
	ProjectRepo       repository.ProjectRepository
}

func NewProjectModule(db *gorm.DB, validate *validator.Validate) *ProjectModule {
	repo := repository.NewProjectRepository(db)
	uc := usecase.NewProjectUseCase(repo)
	ctrl := http.NewProjectController(uc, validate)

	return &ProjectModule{
		ProjectController: ctrl,
		ProjectRepo:       repo,
	}
}
