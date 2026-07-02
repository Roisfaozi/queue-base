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

type BranchServiceController struct {
	useCase  usecase.BranchServiceUseCase
	validate *validator.Validate
}

func NewBranchServiceController(useCase usecase.BranchServiceUseCase, validate *validator.Validate) *BranchServiceController {
	return &BranchServiceController{useCase: useCase, validate: validate}
}

func (h *BranchServiceController) Create(c *gin.Context) {
	var req model.CreateBranchServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateBranchService(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create branch service")
		return
	}
	response.Created(c, res)
}

func (h *BranchServiceController) GetAll(c *gin.Context) {
	res, err := h.useCase.ListBranchServices(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get branch services")
		return
	}
	response.Success(c, res)
}

func (h *BranchServiceController) Update(c *gin.Context) {
	var req model.UpdateBranchServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.UpdateBranchService(c.Request.Context(), c.Param("id"), c.Param("branch_service_id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to update branch service")
		return
	}
	response.Success(c, res)
}

func (h *BranchServiceController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteBranchService(c.Request.Context(), c.Param("id"), c.Param("branch_service_id")); err != nil {
		response.HandleError(c, err, "failed to delete branch service")
		return
	}
	c.Status(http.StatusNoContent)
}
