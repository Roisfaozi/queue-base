package api_key

import (
	"github.com/Roisfaozi/queue-base/internal/modules/api_key/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/api_key/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/api_key/usecase"
	orgRepository "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type ApiKeyModule struct {
	Repo       repository.ApiKeyRepository
	UseCase    usecase.ApiKeyUseCase
	Controller *http.ApiKeyController
}

func NewApiKeyModule(db *gorm.DB, userRepo userRepository.UserRepository, redis *redis.Client, log *logrus.Logger, validator *validator.Validate) *ApiKeyModule {
	repo := repository.NewApiKeyRepository(db)
	orgRepo := orgRepository.NewOrganizationRepository(db)
	useCase := usecase.NewApiKeyUseCase(repo, orgRepo, userRepo, redis, log)
	controller := http.NewApiKeyController(useCase, log, validator)

	return &ApiKeyModule{
		Repo:       repo,
		UseCase:    useCase,
		Controller: controller,
	}
}
