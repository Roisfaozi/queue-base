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

// Create godoc
// @Summary      Create counter
// @Description  Creates a new counter under active tenant scope.
// @Tags         counters
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.CreateCounterRequest true "Create Counter Request"
// @Success      201  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /counters [post]
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

// GetAll godoc
// @Summary      List counters
// @Description  Returns all counters under active tenant scope.
// @Tags         counters
// @Security     BearerAuth
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /counters [get]
func (h *CounterController) GetAll(c *gin.Context) {
	res, err := h.useCase.ListCounters(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get counters")
		return
	}
	response.Success(c, res)
}

// GetByID godoc
// @Summary      Get counter by ID
// @Description  Returns counter details under active tenant scope.
// @Tags         counters
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Counter ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /counters/{id} [get]
func (h *CounterController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetCounter(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get counter")
		return
	}
	response.Success(c, res)
}

// Update godoc
// @Summary      Update counter
// @Description  Updates counter fields under active tenant scope.
// @Tags         counters
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Counter ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.UpdateCounterRequest true "Update Counter Request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /counters/{id} [put]
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

// Delete godoc
// @Summary      Delete counter
// @Description  Deletes counter under active tenant scope.
// @Tags         counters
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Counter ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      204  {object}  nil
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /counters/{id} [delete]
func (h *CounterController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteCounter(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete counter")
		return
	}
	c.Status(http.StatusNoContent)
}
