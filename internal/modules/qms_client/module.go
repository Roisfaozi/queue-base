package qms_client

import (
	qmsClientHttp "github.com/Roisfaozi/queue-base/internal/modules/qms_client/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/usecase"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QMSClientModule struct {
	Repo          repository.QmsClientRepository
	Authenticator usecase.QMSClientAuthenticator
	AdminUseCase  usecase.QMSClientAdminUseCase
	Controller    *qmsClientHttp.QMSClientController
}

func NewQMSClientModule(db *gorm.DB, validate *validator.Validate, log *logrus.Logger, audit ...usecase.AdminAuditLogger) *QMSClientModule {
	repo := repository.NewQmsClientRepository(db)
	auth := usecase.NewQMSClientAuthenticator(repo)
	adminUC := usecase.NewQMSClientAdminUseCase(db, audit...)
	ctrl := qmsClientHttp.NewQMSClientController(adminUC, validate, log)
	return &QMSClientModule{Repo: repo, Authenticator: auth, AdminUseCase: adminUC, Controller: ctrl}
}
