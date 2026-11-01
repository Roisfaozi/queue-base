package tus

import (
	"context"

	"github.com/sirupsen/logrus"
)

// GenericHook handles general uploads without modifying domain-specific models like avatars.
type GenericHook struct {
	Logger *logrus.Logger
}

func (h *GenericHook) HandleUpload(ctx context.Context, event UploadEvent) error {
	if h.Logger != nil {
		h.Logger.WithFields(logrus.Fields{
			"upload_id": event.UploadID,
			"file_url":  event.FileURL,
			"metadata":  event.Metadata,
		}).Info("generic file upload completed successfully")
	}
	return nil
}
