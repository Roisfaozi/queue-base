package http

import (
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
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

// Register godoc
// @Summary      Register queue
// @Description  Creates one queue master row and first queue journey under tenant and branch scope.
// @Tags         queues
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        request body model.RegisterQueueRequest true "Register Queue Request"
// @Success      201  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      409  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /queues [post]
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

// GetAll godoc
// @Summary      List queues
// @Description  Returns queues under active tenant and branch scope.
// @Tags         queues
// @Security     BearerAuth
// @Produce      json
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        status query string false "Queue status"
// @Param        queue_date query string false "Queue date"
// @Param        service_id query string false "Service ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /queues [get]
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

// GetByID godoc
// @Summary      Get queue by ID
// @Description  Returns queue details under active tenant scope.
// @Tags         queues
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Queue ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /queues/{id} [get]
func (h *QueueController) GetByID(c *gin.Context) {
	res, err := h.useCase.GetQueueByID(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.HandleError(c, err, "failed to get queue")
		return
	}
	response.Success(c, res)
}

// Forward godoc
// @Summary      Forward queue
// @Description  Forwards queue by appending queue journey and preserving master queue row.
// @Tags         queues
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Queue ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        body body model.ForwardQueueRequest true "Forward Queue Request"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      422  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /queues/{id}/forward [post]
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
	if c.Param("id") == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing queue id")
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

// GetJourneysByBranchAndService godoc
// @Summary      List active journeys by branch and service
// @Description  Returns active queue journeys filtered by branch and service under active tenant scope.
// @Tags         queues
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Branch ID"
// @Param        service_id path string true "Service ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        queue_date query string false "Queue date"
// @Param        status query string false "Journey status"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /branches/{id}/services/{service_id}/queue-journeys [get]
func (h *QueueController) GetJourneysByBranchAndService(c *gin.Context) {
	var req model.QueueJourneyListRequest
	req.ServiceID = c.Param("service_id")
	branchID := c.Param("id")
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

// GetJourneysByBranchAndCounter godoc
// @Summary      List active journeys by branch and counter
// @Description  Returns active queue journeys filtered by branch and counter under active tenant scope.
// @Tags         queues
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Branch ID"
// @Param        counter_id path string true "Counter ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Param        queue_date query string false "Queue date"
// @Param        status query string false "Journey status"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /branches/{id}/counters/{counter_id}/queue-journeys [get]
func (h *QueueController) GetJourneysByBranchAndCounter(c *gin.Context) {
	var req model.QueueJourneyListRequest
	req.CounterID = c.Param("counter_id")
	branchID := c.Param("id")
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

// GetVisitJourneys godoc
// @Summary      Get visit journeys by queue
// @Description  Returns readable visit history for a queue under active tenant scope.
// @Tags         queues
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Queue ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      404  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /queues/{id}/visit-journeys [get]
func (h *QueueController) GetVisitJourneys(c *gin.Context) {
	queueID := c.Param("id")
	if queueID == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing queue id")
		return
	}
	res, err := h.useCase.GetVisitJourneys(c.Request.Context(), queueID)
	if err != nil {
		response.HandleError(c, err, "failed to get visit journeys")
		return
	}
	response.Success(c, res)
}

// GetQueueStats godoc
// @Summary      Get queue stats by branch
// @Description  Returns queue statistics for a branch under active tenant scope.
// @Tags         queues
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Branch ID"
// @Param        X-Organization-ID header string true "Tenant ID"
// @Success      200  {object}  response.SwaggerSuccessResponseWrapper
// @Failure      400  {object}  response.SwaggerErrorResponseWrapper
// @Failure      500  {object}  response.SwaggerErrorResponseWrapper
// @Router       /branches/{id}/queue-stats [get]
func (h *QueueController) GetQueueStats(c *gin.Context) {
	branchID := c.Param("id")
	ctx := database.SetBranchContext(c.Request.Context(), branchID)
	res, err := h.useCase.GetQueueStats(ctx)
	if err != nil {
		response.HandleError(c, err, "failed to get queue stats")
		return
	}
	response.Success(c, res)
}
