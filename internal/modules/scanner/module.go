package scanner

import (
	"context"

	counterModulePkg "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/counter"
	branchModulePkg "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization"
	queueModulePkg "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue"
	scannerHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/scanner/delivery/http"
	scannerUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/scanner/usecase"
	serviceModulePkg "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/service"
	"github.com/go-playground/validator/v10"
)

type ScannerAuthenticator interface {
	Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error
}

type ScannerModule struct {
	ScannerController *scannerHttp.ScannerController
	ScannerUseCase    scannerUsecase.ScannerUseCase
}

type NoopAuthenticator struct{}

func (NoopAuthenticator) Authenticate(ctx context.Context, tenantID, branchID, clientID, apiKey string) error {
	return nil
}

func NewScannerModule(queueModule *queueModulePkg.QueueModule, branchModule *branchModulePkg.BranchModule, serviceModule *serviceModulePkg.ServiceModule, counterModule *counterModulePkg.CounterModule, validate *validator.Validate, authenticator ScannerAuthenticator) *ScannerModule {
	relationValidator := scannerUsecase.NewRelationValidator(branchModule.BranchRepo, serviceModule.ServiceRepo, counterModule.CounterRepo)
	uc := scannerUsecase.NewScannerUseCase(queueModule.QueueUseCase, authenticator, relationValidator)
	ctrl := scannerHttp.NewScannerController(uc, validate)
	return &ScannerModule{ScannerController: ctrl, ScannerUseCase: uc}
}
