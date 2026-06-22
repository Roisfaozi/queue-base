package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/response"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type QueueController struct {
	useCase  usecase.QueueUseCase
	validate *validator.Validate
}

func NewQueueController(useCase usecase.QueueUseCase, validate *validator.Validate) *QueueController {
	return &QueueController{useCase: useCase, validate: validate}
}

func (h *QueueController) Register(c *gin.Context) {
	var req model.RegisterQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.RegisterQueue(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to register queue")
		return
	}
	response.Created(c, res)
}

func (h *QueueController) Forward(c *gin.Context) {
	var req model.ForwardQueueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.ForwardQueue(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to forward queue")
		return
	}
	response.Success(c, res)
}
