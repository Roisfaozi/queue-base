package http

import (
	"github.com/Roisfaozi/queue-base/internal/middleware"
	"github.com/Roisfaozi/queue-base/internal/modules/caller/model"
	"github.com/Roisfaozi/queue-base/internal/modules/caller/usecase"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type CallerController struct {
	useCase  usecase.CallerUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewCallerController(uc usecase.CallerUseCase, v *validator.Validate, log *logrus.Logger) *CallerController {
	return &CallerController{useCase: uc, validate: v, log: log}
}

// Action godoc
// @Summary      Caller action on queue journey
// @Description  Perform call/serve/complete/skip/cancel on a queue journey.
// @Tags         caller
// @Accept       json
// @Produce      json
// @Param        journey_id path string true "Queue Journey ID"
// @Param        X-Client-ID header string false "QMS Client ID"
// @Param        X-API-Key header string false "QMS Client API Key"
// @Param        request body model.CallerActionRequest true "Action request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /caller/queue-journeys/{journey_id}/action [post]
func (h *CallerController) Action(c *gin.Context) {
	var req model.CallerActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, nil, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	journeyID := c.Param("journey_id")
	if journeyID == "" {
		response.BadRequest(c, nil, "missing journey_id")
		return
	}
	clientID := middleware.GetQMSClientIDFromContext(c)
	res, err := h.useCase.ExecuteAction(c.Request.Context(), clientID, journeyID, req.Action)
	if err != nil {
		h.log.WithError(err).Error("caller action failed")
		response.HandleError(c, err, "caller action failed")
		return
	}
	response.Success(c, res)
}
