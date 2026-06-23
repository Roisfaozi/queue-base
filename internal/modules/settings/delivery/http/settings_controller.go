package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type SettingsController struct {
	useCase  usecase.SettingsUseCase
	validate *validator.Validate
}

func NewSettingsController(useCase usecase.SettingsUseCase, validate *validator.Validate) *SettingsController {
	return &SettingsController{useCase: useCase, validate: validate}
}

// Create godoc
// @Summary      Create setting
// @Description  Creates tenant, branch, service, or counter scoped setting override.
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.CreateSettingRequest true "Create Setting Request"
// @Success      201  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /settings [post]
func (h *SettingsController) Create(c *gin.Context) {
	var req model.CreateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateSetting(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create setting")
		return
	}
	response.Created(c, res)
}

// GetByID godoc
// @Summary      Get setting by ID
// @Description  Returns setting details under active tenant scope.
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Setting ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /settings/{id} [get]
func (h *SettingsController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetSetting(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get setting")
		return
	}
	response.Success(c, res)
}

// Resolve godoc
// @Summary      Resolve effective setting
// @Description  Resolves setting value using tenant -> branch -> service -> counter inheritance.
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        Key query string true "Setting key"
// @Param        BranchID query string false "Branch ID"
// @Param        ServiceID query string false "Service ID"
// @Param        CounterID query string false "Counter ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /settings/resolve [get]
func (h *SettingsController) Resolve(c *gin.Context) {
	var req model.ResolveSettingRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid query parameters")
		return
	}
	if database.GetTenantID(c.Request.Context()) == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing tenant context")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.ResolveSetting(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to resolve setting")
		return
	}
	response.Success(c, res)
}

// Update godoc
// @Summary      Update setting
// @Description  Updates scoped setting override under active tenant scope.
// @Tags         settings
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Setting ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.UpdateSettingRequest true "Update Setting Request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /settings/{id} [put]
func (h *SettingsController) Update(c *gin.Context) {
	var req model.UpdateSettingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.UpdateSetting(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to update setting")
		return
	}
	response.Success(c, res)
}

// Delete godoc
// @Summary      Delete setting
// @Description  Deletes scoped setting override under active tenant scope.
// @Tags         settings
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Setting ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      204  {object}  nil
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /settings/{id} [delete]
func (h *SettingsController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteSetting(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete setting")
		return
	}
	c.Status(http.StatusNoContent)
}
