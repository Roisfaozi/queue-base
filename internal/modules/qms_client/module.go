package qms_client

import (
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/usecase"
	"gorm.io/gorm"
)

type QMSClientModule struct {
	Repo          repository.QmsClientRepository
	Authenticator usecase.QMSClientAuthenticator
}

func NewQMSClientModule(db *gorm.DB) *QMSClientModule {
	repo := repository.NewQmsClientRepository(db)
	auth := usecase.NewQMSClientAuthenticator(repo)
	return &QMSClientModule{Repo: repo, Authenticator: auth}
}
