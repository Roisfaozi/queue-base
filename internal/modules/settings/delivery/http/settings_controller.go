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

func (h *SettingsController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetSetting(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get setting")
		return
	}
	response.Success(c, res)
}

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

func (h *SettingsController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteSetting(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete setting")
		return
	}
	c.Status(http.StatusNoContent)
}
