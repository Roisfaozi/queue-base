package http

import (
	"github.com/Roisfaozi/queue-base/internal/modules/signage/usecase"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type SignageController struct {
	useCase  usecase.SignageUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewSignageController(uc usecase.SignageUseCase, v *validator.Validate, log *logrus.Logger) *SignageController {
	return &SignageController{useCase: uc, validate: v, log: log}
}

// Me godoc
// @Summary      Get signage client info
// @Description  Returns bound tenant, branch, and profile info for signage client.
// @Tags         signage
// @Accept       json
// @Produce      json
// @Param        X-Client-ID header string true "QMS Client ID"
// @Param        X-API-Key header string true "QMS Client API Key"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      401  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /signage/me [get]
func (h *SignageController) Me(c *gin.Context) {
	// TODO: implement context extraction
	res, err := h.useCase.GetMe(c.Request.Context(), "dummy-client-id")
	if err != nil {
		response.HandleError(c, err, "failed to get signage info")
		return
	}
	response.Success(c, res)
}

// CurrentCalls godoc
// @Summary      Get current calls for signage
// @Description  Returns active calling queues for the bound signage scope.
// @Tags         signage
// @Produce      json
// @Param        X-Client-ID header string true "QMS Client ID"
// @Param        X-API-Key header string true "QMS Client API Key"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Router       /signage/current-calls [get]
func (h *SignageController) CurrentCalls(c *gin.Context) {
	res, err := h.useCase.GetCurrentCalls(c.Request.Context(), "dummy-client-id")
	if err != nil {
		response.HandleError(c, err, "failed to get current calls")
		return
	}
	response.Success(c, res)
}

// Queues godoc
// @Summary      Get waiting queues for signage
// @Description  Returns active waiting queues for the bound signage scope.
// @Tags         signage
// @Produce      json
// @Param        X-Client-ID header string true "QMS Client ID"
// @Param        X-API-Key header string true "QMS Client API Key"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Router       /signage/queues [get]
func (h *SignageController) Queues(c *gin.Context) {
	res, err := h.useCase.GetQueues(c.Request.Context(), "dummy-client-id")
	if err != nil {
		response.HandleError(c, err, "failed to get waiting queues")
		return
	}
	response.Success(c, res)
}
