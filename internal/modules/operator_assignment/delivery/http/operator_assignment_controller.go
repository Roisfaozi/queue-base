package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/model"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type Controller struct {
	useCase  usecase.OperatorAssignmentUseCase
	validate *validator.Validate
	log      *logrus.Logger
}

func NewController(useCase usecase.OperatorAssignmentUseCase, validate *validator.Validate, log *logrus.Logger) *Controller {
	return &Controller{useCase: useCase, validate: validate, log: log}
}

func (h *Controller) Create(c *gin.Context) {
	var req model.OperatorAssignmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.Create(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create operator assignment")
		return
	}
	response.Created(c, res)
}

func (h *Controller) GetAll(c *gin.Context) {
	res, err := h.useCase.GetAll(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get operator assignments")
		return
	}
	response.Success(c, res)
}

func (h *Controller) Delete(c *gin.Context) {
	if err := h.useCase.Delete(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to unassign operator")
		return
	}
	c.Status(http.StatusNoContent)
}
