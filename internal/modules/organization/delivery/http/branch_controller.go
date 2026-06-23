package http

import (
	"net/http"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type BranchController struct {
	useCase  usecase.BranchUseCase
	validate *validator.Validate
}

func NewBranchController(useCase usecase.BranchUseCase, validate *validator.Validate) *BranchController {
	return &BranchController{useCase: useCase, validate: validate}
}

// GetByID godoc
// @Summary      Get branch by ID
// @Description  Returns a branch only when it belongs to active tenant context.
// @Tags         branches
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Branch ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /branches/{id} [get]
func (h *BranchController) GetByID(c *gin.Context) {
	res, err := h.useCase.ResolveBranch(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get branch")
		return
	}
	response.Success(c, res)
}

func (h *BranchController) Create(c *gin.Context) {
	var req model.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.CreateBranch(c.Request.Context(), &req)
	if err != nil {
		response.HandleError(c, err, "failed to create branch")
		return
	}
	response.Created(c, res)
}

func (h *BranchController) GetAll(c *gin.Context) {
	res, err := h.useCase.ListBranches(c.Request.Context())
	if err != nil {
		response.HandleError(c, err, "failed to get branches")
		return
	}
	response.Success(c, res)
}

func (h *BranchController) Update(c *gin.Context) {
	var req model.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.UpdateBranch(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to update branch")
		return
	}
	response.Success(c, res)
}

func (h *BranchController) Delete(c *gin.Context) {
	if err := h.useCase.DeleteBranch(c.Request.Context(), c.Param("id")); err != nil {
		response.HandleError(c, err, "failed to delete branch")
		return
	}
	c.Status(http.StatusNoContent)
}
