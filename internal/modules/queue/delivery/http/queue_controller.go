package http

import (
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/usecase"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
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

func (h *QueueController) GetAll(c *gin.Context) {
	var req model.ListQueuesRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid query params")
		return
	}
	res, err := h.useCase.ListQueues(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err, "failed to get queues")
		return
	}
	response.Success(c, res)
}

func (h *QueueController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetQueueByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get queue")
		return
	}
	response.Success(c, res)
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

// Transition godoc
// @Summary      Transition queue state
// @Description  Moves queue state through call, serve, complete, skip, or cancel under tenant scope.
// @Tags         queues
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Queue ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        body body model.QueueTransitionRequest true "Transition request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /queues/{id}/transition [post]

func (h *QueueController) Transition(c *gin.Context) {
	var req model.QueueTransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid request body")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.useCase.TransitionQueue(c.Request.Context(), c.Param("id"), &req)
	if err != nil {
		response.HandleError(c, err, "failed to transition queue")
		return
	}
	response.Success(c, res)
}

func (h *QueueController) GetJourneysByBranchAndService(c *gin.Context) {
	var req model.QueueJourneyListRequest
	req.ServiceID = c.Param("service_id")
	branchID := c.Param("branch_id")
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid query params")
		return
	}
	ctx := database.SetBranchContext(c.Request.Context(), branchID)
	res, err := h.useCase.ListActiveJourneys(ctx, req)
	if err != nil {
		response.HandleError(c, err, "failed to get queue journeys")
		return
	}
	response.Success(c, res)
}

func (h *QueueController) GetJourneysByBranchAndCounter(c *gin.Context) {
	var req model.QueueJourneyListRequest
	req.CounterID = c.Param("counter_id")
	branchID := c.Param("branch_id")
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid query params")
		return
	}
	ctx := database.SetBranchContext(c.Request.Context(), branchID)
	res, err := h.useCase.ListActiveJourneys(ctx, req)
	if err != nil {
		response.HandleError(c, err, "failed to get queue journeys")
		return
	}
	response.Success(c, res)
}

func (h *QueueController) GetVisitJourneys(c *gin.Context) {
	res, err := h.useCase.GetVisitJourneys(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get visit journeys")
		return
	}
	response.Success(c, res)
}

func (h *QueueController) GetQueueStats(c *gin.Context) {
	branchID := c.Param("branch_id")
	ctx := database.SetBranchContext(c.Request.Context(), branchID)
	res, err := h.useCase.GetQueueStats(ctx)
	if err != nil {
		response.HandleError(c, err, "failed to get queue stats")
		return
	}
	response.Success(c, res)
}
