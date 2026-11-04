package usecase_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	storageMocks "github.com/Roisfaozi/queue-base/pkg/storage/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var pngHeader = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, 0, 0, 0, 0x0D, 'I', 'H', 'D', 'R'}

func TestLogoUseCase_UploadLogo(t *testing.T) {
	log := logrus.New()

	tests := []struct {
		name      string
		tenantID  string
		scope     model.LogoScope
		content   []byte
		uploadErr error
		wantErr   error
		wantURL   string
		wantPath  string
	}{
		{
			name: "Positive_PNGOrganizationLogo", tenantID: "tenant-1",
			scope: model.LogoScopeOrganization, content: pngHeader,
			wantURL: "https://cdn/logo.png", wantPath: "logos/tenant-1/organization_",
		},
		{
			name: "Positive_PNGBranchLogo", tenantID: "tenant-1",
			scope: model.LogoScopeBranch, content: pngHeader,
			wantURL: "https://cdn/logo.png", wantPath: "logos/tenant-1/branch_",
		},
		{
			name: "Negative_UnsupportedType", tenantID: "tenant-1",
			scope: model.LogoScopeBranch, content: []byte("plain text, not an image"),
			wantErr: exception.ErrValidationError,
		},
		{
			name: "Negative_InvalidScope", tenantID: "tenant-1",
			scope: model.LogoScope("other"), content: pngHeader,
			wantErr: exception.ErrBadRequest,
		},
		{
			name: "Negative_StorageFailure", tenantID: "tenant-1",
			scope: model.LogoScopeBranch, content: pngHeader,
			uploadErr: errors.New("s3 down"), wantErr: exception.ErrInternalServer,
			wantPath: "logos/tenant-1/branch_",
		},
		{
			name: "Edge_EmptyFile", tenantID: "tenant-1",
			scope: model.LogoScopeBranch, content: []byte{},
			wantErr: exception.ErrValidationError,
		},
		{
			name: "Edge_SVGRejectedForDirectUpload", tenantID: "tenant-1",
			scope: model.LogoScopeBranch, content: []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`),
			wantErr: exception.ErrValidationError,
		},
		{
			name: "Security_MissingTenantContext", tenantID: "",
			scope: model.LogoScopeOrganization, content: pngHeader,
			wantErr: exception.ErrBadRequest,
		},
		{
			name: "Security_DisguisedExecutableRejected", tenantID: "tenant-1",
			scope: model.LogoScopeBranch, content: append([]byte("MZ"), bytes.Repeat([]byte{0}, 600)...),
			wantErr: exception.ErrValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}

			store := new(storageMocks.MockProvider)
			if tt.wantPath != "" {
				store.On("UploadFile", mock.Anything, mock.Anything,
					mock.MatchedBy(func(name string) bool { return strings.HasPrefix(name, tt.wantPath) }),
					"image/png",
				).Return(tt.wantURL, tt.uploadErr).Once()
			}

			uc := usecase.NewLogoUseCase(store, log)
			url, err := uc.UploadLogo(ctx, tt.scope, bytes.NewReader(tt.content))

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Empty(t, url)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantURL, url)
			}
			store.AssertExpectations(t)
		})
	}
}
