package http

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func buildLogoRequest(t *testing.T, field string, data []byte) *http.Request {
	t.Helper()
	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	if field != "" {
		part, err := w.CreateFormFile(field, "logo.png")
		assert.NoError(t, err)
		_, _ = part.Write(data)
	}
	assert.NoError(t, w.Close())
	req, _ := http.NewRequest(http.MethodPost, "/logo", body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestLogoController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		field      string
		size       int
		ucErr      error
		callsUC    bool
		wantStatus int
	}{
		{name: "Positive_Upload", field: "logo", size: 1024, callsUC: true, wantStatus: http.StatusOK},
		{name: "Negative_MissingFile", field: "", wantStatus: http.StatusBadRequest},
		{name: "Negative_WrongFieldName", field: "avatar", size: 10, wantStatus: http.StatusBadRequest},
		{name: "Negative_InvalidImageType", field: "logo", size: 10, callsUC: true, ucErr: exception.ErrValidationError, wantStatus: http.StatusUnprocessableEntity},
		{name: "Edge_ExactlyAtLimit", field: "logo", size: MaxDirectLogoSize, callsUC: true, wantStatus: http.StatusOK},
		{name: "Edge_OverLimit", field: "logo", size: MaxDirectLogoSize + 1, wantStatus: http.StatusBadRequest},
		{name: "Security_TenantMissing", field: "logo", size: 10, callsUC: true, ucErr: exception.ErrBadRequest, wantStatus: http.StatusBadRequest},
		{name: "Negative_StorageError", field: "logo", size: 10, callsUC: true, ucErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := mocks.NewMockLogoUseCase(t)
			if tt.callsUC {
				uc.On("UploadLogo", mock.Anything, model.LogoScopeBranch, mock.Anything).
					Return("https://cdn/logo.png", tt.ucErr).Once()
			}
			ctrl := NewLogoController(uc, logrus.New())

			r := gin.New()
			r.POST("/logo", ctrl.UploadBranchLogo)

			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, buildLogoRequest(t, tt.field, make([]byte, tt.size)))

			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
