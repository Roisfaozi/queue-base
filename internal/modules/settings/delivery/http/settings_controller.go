package http

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type QueueSettingResolver interface {
	Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error)
	ResolveDetailed(ctx context.Context, key string, branchID string, serviceID string, counterID string) (*model.ResolvedQueueSetting, error)
}

type SettingsController struct {
	queueResolver QueueSettingResolver
	validate      *validator.Validate
	log           *logrus.Logger
}

func (h *SettingsController) EffectiveQueueConfig(c *gin.Context) {
	var req model.EffectiveQueueConfigRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, exception.ErrBadRequest, "invalid query parameters")
		return
	}
	tenantID := database.GetTenantID(c.Request.Context())
	if tenantID == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing tenant context")
		return
	}
	if err := h.validate.Struct(req); err != nil {
		response.ValidationError(c, err, validation.FormatValidationErrors(err))
		return
	}
	res, err := h.resolveEffectiveQueueConfig(c.Request.Context(), tenantID, req)
	if err != nil {
		response.HandleError(c, err, "failed to resolve effective queue config")
		return
	}
	response.Success(c, res)
}

func (h *SettingsController) resolveEffectiveQueueConfig(ctx context.Context, tenantID string, req model.EffectiveQueueConfigRequest) (*model.EffectiveQueueConfigResponse, error) {
	queueResetTime, err := h.queueResolver.Resolve(ctx, "queue_reset_time", req.BranchID, req.ServiceID, req.CounterID)
	if err != nil {
		return nil, err
	}
	ticketPrefix, err := h.queueResolver.Resolve(ctx, "ticket_prefix", req.BranchID, req.ServiceID, req.CounterID)
	if err != nil {
		return nil, err
	}
	numberingStrategy, err := h.queueResolver.Resolve(ctx, "numbering_strategy", req.BranchID, req.ServiceID, req.CounterID)
	if err != nil {
		return nil, err
	}
	defaultEstimatedDuration, _ := h.queueResolver.Resolve(ctx, "default_estimated_duration", req.BranchID, req.ServiceID, req.CounterID)
	autoCallNext, _ := h.queueResolver.Resolve(ctx, "auto_call_next", req.BranchID, req.ServiceID, req.CounterID)
	queueResetTimeResolved, _ := h.queueResolver.ResolveDetailed(ctx, "queue_reset_time", req.BranchID, req.ServiceID, req.CounterID)
	ticketPrefixResolved, _ := h.queueResolver.ResolveDetailed(ctx, "ticket_prefix", req.BranchID, req.ServiceID, req.CounterID)
	numberingStrategyResolved, _ := h.queueResolver.ResolveDetailed(ctx, "numbering_strategy", req.BranchID, req.ServiceID, req.CounterID)
	defaultEstimatedDurationResolved, _ := h.queueResolver.ResolveDetailed(ctx, "default_estimated_duration", req.BranchID, req.ServiceID, req.CounterID)

	return &model.EffectiveQueueConfigResponse{
		TenantID:                          tenantID,
		BranchID:                          req.BranchID,
		ServiceID:                         req.ServiceID,
		CounterID:                         req.CounterID,
		QueueResetTime:                    queueResetTime,
		QueueResetTimeSource:              sourceOf(queueResetTimeResolved),
		QueueResetTimeInherited:           inheritedOf(queueResetTimeResolved),
		TicketPrefix:                      ticketPrefix,
		TicketPrefixSource:                sourceOf(ticketPrefixResolved),
		TicketPrefixInherited:             inheritedOf(ticketPrefixResolved),
		NumberingStrategy:                 numberingStrategy,
		NumberingStrategySource:           sourceOf(numberingStrategyResolved),
		NumberingStrategyInherited:        inheritedOf(numberingStrategyResolved),
		DefaultEstimatedDuration:          defaultEstimatedDuration,
		DefaultEstimatedDurationSource:    sourceOf(defaultEstimatedDurationResolved),
		DefaultEstimatedDurationInherited: inheritedOf(defaultEstimatedDurationResolved),
		AutoCallNext:                      parseBoolPtr(autoCallNext),
	}, nil
}

func NewSettingsController(validate *validator.Validate, resolver QueueSettingResolver, log *logrus.Logger) *SettingsController {
	return &SettingsController{queueResolver: resolver, validate: validate, log: log}
}

func sourceOf(resolved *model.ResolvedQueueSetting) string {
	if resolved == nil {
		return ""
	}
	return resolved.Source
}

func inheritedOf(resolved *model.ResolvedQueueSetting) bool {
	if resolved == nil {
		return false
	}
	return resolved.Inherited
}

func parseBoolPtr(value string) *bool {
	if value == "true" {
		v := true
		return &v
	}
	if value == "false" {
		v := false
		return &v
	}
	return nil
}
