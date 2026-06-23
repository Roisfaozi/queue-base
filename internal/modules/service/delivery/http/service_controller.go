package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
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

// Create godoc
// @Summary      Create service
// @Description  Creates a new service under active tenant scope.
// @Tags         services
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.CreateServiceRequest true "Create Service Request"
// @Success      201  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /services [post]
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

// GetAll godoc
// @Summary      List services
// @Description  Returns all services under active tenant scope.
// @Tags         services
// @Security     BearerAuth
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /services [get]
func (h *ServiceController) GetAll(c *gin.Context) {
	if database.GetTenantID(c.Request.Context()) == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing tenant context")
		return
	}
	res, err := h.useCase.ListServices(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get services")
		return
	}
	response.Success(c, res)
}

// GetByID godoc
// @Summary      Get service by ID
// @Description  Returns service details under active tenant scope.
// @Tags         services
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Service ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /services/{id} [get]
func (h *ServiceController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetService(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get service")
		return
	}
	response.Success(c, res)
}

// Update godoc
// @Summary      Update service
// @Description  Updates service fields under active tenant scope.
// @Tags         services
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Service ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.UpdateServiceRequest true "Update Service Request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /services/{id} [put]
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

// Delete godoc
// @Summary      Delete service
// @Description  Deletes service under active tenant scope.
// @Tags         services
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Service ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      204  {object}  nil
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /services/{id} [delete]
func (h *ServiceController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteService(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete service")
		return
	}
	c.Status(http.StatusNoContent)
}
