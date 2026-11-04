package usecase

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/storage"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LogoUseCase handles direct (non-resumable) logo uploads for small files.
// Larger files are expected to go through the TUS upload endpoint.
type LogoUseCase interface {
	// UploadLogo validates the image and stores it under the active tenant, returning its URL.
	UploadLogo(ctx context.Context, scope model.LogoScope, file io.Reader) (string, error)
}

type logoUseCase struct {
	storage storage.Provider
	log     *logrus.Logger
}

func NewLogoUseCase(storage storage.Provider, log *logrus.Logger) LogoUseCase {
	return &logoUseCase{storage: storage, log: log}
}

var logoAllowedTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func (u *logoUseCase) UploadLogo(ctx context.Context, scope model.LogoScope, file io.Reader) (string, error) {
	if scope != model.LogoScopeOrganization && scope != model.LogoScopeBranch {
		return "", exception.ErrBadRequest
	}

	tenantID := database.GetTenantID(ctx)
	if tenantID == "" {
		return "", exception.ErrBadRequest
	}
	if u.storage == nil {
		return "", exception.ErrInternalServer
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(file, buf)
	if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
		u.log.WithError(err).Warn("failed to read logo for type detection")
		return "", exception.ErrBadRequest
	}
	buf = buf[:n]

	detectedType := http.DetectContentType(buf)
	ext, ok := logoAllowedTypes[detectedType]
	if !ok {
		u.log.Warnf("blocked logo upload of type: %s", detectedType)
		return "", exception.ErrValidationError
	}

	file = io.MultiReader(bytes.NewReader(buf), file)
	filename := fmt.Sprintf("logos/%s/%s_%s%s", tenantID, scope, uuid.NewString(), ext)

	url, err := u.storage.UploadFile(ctx, file, filename, detectedType)
	if err != nil {
		u.log.WithError(err).Error("failed to upload logo")
		return "", exception.ErrInternalServer
	}

	u.log.Infof("logo uploaded tenant=%s scope=%s", tenantID, scope)
	return url, nil
}
