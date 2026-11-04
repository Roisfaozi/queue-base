package http

import (
	"errors"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// MaxDirectLogoSize is the largest logo accepted by the direct upload endpoints.
const MaxDirectLogoSize = 15 * 1024 * 1024

type LogoController struct {
	useCase usecase.LogoUseCase
	log     *logrus.Logger
}

func NewLogoController(useCase usecase.LogoUseCase, log *logrus.Logger) *LogoController {
	return &LogoController{useCase: useCase, log: log}
}

// UploadOrganizationLogo godoc
// @Summary      Upload organization logo
// @Description  Uploads an organization logo (max 15MB, JPEG/PNG/GIF/WEBP) and returns its URL.
// @Tags         organizations
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        logo formData file true "Logo image"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /tenant/logo [post]
func (h *LogoController) UploadOrganizationLogo(c *gin.Context) {
	h.upload(c, model.LogoScopeOrganization)
}

// UploadBranchLogo godoc
// @Summary      Upload branch logo
// @Description  Uploads a branch logo (max 15MB, JPEG/PNG/GIF/WEBP) and returns its URL.
// @Tags         branches
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        logo formData file true "Logo image"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /branches/logo [post]
func (h *LogoController) UploadBranchLogo(c *gin.Context) {
	h.upload(c, model.LogoScopeBranch)
}

func (h *LogoController) upload(c *gin.Context, scope model.LogoScope) {
	file, header, err := c.Request.FormFile("logo")
	if err != nil {
		h.log.WithError(err).Error("failed to get logo file from request")
		response.BadRequest(c, exception.ErrBadRequest, "logo file is required")
		return
	}
	defer func() {
		_ = file.Close()
	}()

	if header.Size > MaxDirectLogoSize {
		response.BadRequest(c, errors.New("file too large"), "logo size must be less than 15MB")
		return
	}

	url, err := h.useCase.UploadLogo(c.Request.Context(), scope, file)
	if err != nil {
		h.log.WithError(err).Error("failed to upload logo")
		response.HandleError(c, err, "failed to upload logo")
		return
	}

	response.Success(c, gin.H{"url": url})
}
