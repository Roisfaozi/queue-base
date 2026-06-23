package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type CounterController struct {
	useCase  usecase.CounterUseCase
	validate *validator.Validate
}

func NewCounterController(useCase usecase.CounterUseCase, validate *validator.Validate) *CounterController {
	return &CounterController{useCase: useCase, validate: validate}
}

func (h *CounterController) Create(c *gin.Context) {
	var req model.CreateCounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateCounter(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create counter")
		return
	}
	response.Created(c, res)
}

func (h *CounterController) GetAll(c *gin.Context) {
	res, err := h.useCase.ListCounters(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get counters")
		return
	}
	response.Success(c, res)
}

func (h *CounterController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetCounter(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get counter")
		return
	}
	response.Success(c, res)
}

func (h *CounterController) Update(c *gin.Context) {
	var req model.UpdateCounterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.UpdateCounter(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to update counter")
		return
	}
	response.Success(c, res)
}

func (h *CounterController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteCounter(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete counter")
		return
	}
	c.Status(http.StatusNoContent)
}
