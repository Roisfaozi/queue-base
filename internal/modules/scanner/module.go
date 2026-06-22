package scanner

import (
	"context"

	queueModulePkg "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue"
	scannerHttp "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/scanner/delivery/http"
	scannerUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/scanner/usecase"
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

func NewScannerModule(queueModule *queueModulePkg.QueueModule, validate *validator.Validate, authenticator ScannerAuthenticator) *ScannerModule {
	uc := scannerUsecase.NewScannerUseCase(queueModule.QueueUseCase, authenticator)
	ctrl := scannerHttp.NewScannerController(uc, validate)
	return &ScannerModule{ScannerController: ctrl, ScannerUseCase: uc}
}
