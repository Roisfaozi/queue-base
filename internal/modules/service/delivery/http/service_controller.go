package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ServiceController struct {
	useCase  usecase.ServiceUseCase
	validate *validator.Validate
}

func NewServiceController(useCase usecase.ServiceUseCase, validate *validator.Validate) *ServiceController {
	return &ServiceController{useCase: useCase, validate: validate}
}

func (h *ServiceController) Create(c *gin.Context) {
	var req model.CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateService(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create service")
		return
	}
	response.Created(c, res)
}

func (h *ServiceController) GetAll(c *gin.Context) {
	res, err := h.useCase.ListServices(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get services")
		return
	}
	response.Success(c, res)
}

func (h *ServiceController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetService(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get service")
		return
	}
	response.Success(c, res)
}

func (h *ServiceController) Update(c *gin.Context) {
	var req model.UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.UpdateService(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to update service")
		return
	}
	response.Success(c, res)
}

func (h *ServiceController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteService(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete service")
		return
	}
	c.Status(http.StatusNoContent)
}
