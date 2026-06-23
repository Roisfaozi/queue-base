package scanner

import (
	"context"

	apiKeyModel "github.com/Roisfaozi/queue-base/internal/modules/api_key/model"

	counterModulePkg "github.com/Roisfaozi/queue-base/internal/modules/counter"
	branchModulePkg "github.com/Roisfaozi/queue-base/internal/modules/organization"
	queueModulePkg "github.com/Roisfaozi/queue-base/internal/modules/queue"
	scannerHttp "github.com/Roisfaozi/queue-base/internal/modules/scanner/delivery/http"
	scannerUsecase "github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	serviceModulePkg "github.com/Roisfaozi/queue-base/internal/modules/service"
	settingsModulePkg "github.com/Roisfaozi/queue-base/internal/modules/settings"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/go-playground/validator/v10"
)

type ScannerAuthenticator interface {
	Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error
}

type apiKeyAuthenticator interface {
	Authenticate(ctx context.Context, key string) (*apiKeyModel.ApiKeyIdentity, error)
}

type ScannerModule struct {
	ScannerController *scannerHttp.ScannerController
	ScannerUseCase    scannerUsecase.ScannerUseCase
}

type NoopAuthenticator struct{}

func (NoopAuthenticator) Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error {
	return nil
}

type APIKeyAuthenticator struct {
	apiKeyUseCase apiKeyAuthenticator
}

func NewAPIKeyAuthenticator(apiKeyUseCase apiKeyAuthenticator) APIKeyAuthenticator {
	return APIKeyAuthenticator{apiKeyUseCase: apiKeyUseCase}
}

func (a APIKeyAuthenticator) Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error {
	if a.apiKeyUseCase == nil {
		return nil
	}
	identity, err := a.apiKeyUseCase.Authenticate(ctx, apiKey)
	if err != nil {
		return err
	}
	if identity.OrganizationID != tenantID {
		return exception.ErrUnauthorized
	}
	if clientID != "" && identity.UserID != clientID {
		return exception.ErrUnauthorized
	}
	return nil
}

func NewScannerModule(queueModule *queueModulePkg.QueueModule, branchModule *branchModulePkg.BranchModule, serviceModule *serviceModulePkg.ServiceModule, counterModule *counterModulePkg.CounterModule, settingsModule *settingsModulePkg.SettingsModule, validate *validator.Validate, authenticator ScannerAuthenticator) *ScannerModule {
	var resolver *settingsModulePkg.QueueSettingsResolver
	if settingsModule != nil {
		resolver = settingsModule.QueueSettingsResolver
	}
	relationValidator := scannerUsecase.NewRelationValidator(branchModule.BranchRepo, serviceModule.ServiceRepo, counterModule.CounterRepo, resolver)
	uc := scannerUsecase.NewScannerUseCase(queueModule.QueueUseCase, authenticator, relationValidator)
	ctrl := scannerHttp.NewScannerController(uc, validate)
	return &ScannerModule{ScannerController: ctrl, ScannerUseCase: uc}
}
