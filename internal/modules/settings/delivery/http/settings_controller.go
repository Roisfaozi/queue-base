package http

import (
	"context"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QueueSettingResolver interface {
	Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error)
	ResolveDetailed(ctx context.Context, key string, branchID string, serviceID string, counterID string) (*model.ResolvedQueueSetting, error)
}

type AuditLogger interface {
	LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error
}

type SettingsController struct {
	queueResolver QueueSettingResolver
	validate      *validator.Validate
	log           *logrus.Logger
	db            *gorm.DB
	audit         AuditLogger
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

func (h *SettingsController) ResetBranchQueueSetting(c *gin.Context) {
	h.resetQueueSetting(c, "branch_queue_settings", "branch_id", branchParam(c), c.Param("field"))
}

func (h *SettingsController) ResetBranchServiceQueueSetting(c *gin.Context) {
	h.resetQueueSetting(c, "branch_service_queue_settings", "branch_service_id", c.Param("branch_service_id"), c.Param("field"))
}

func (h *SettingsController) ResetCounterQueueSetting(c *gin.Context) {
	h.resetQueueSetting(c, "counter_queue_settings", "counter_id", c.Param("counter_id"), c.Param("field"))
}

func (h *SettingsController) resetQueueSetting(c *gin.Context, table, idColumn, id, field string) {
	if h.db == nil || id == "" {
		response.BadRequest(c, exception.ErrBadRequest, "invalid reset request")
		return
	}
	tenantID := database.GetTenantID(c.Request.Context())
	if tenantID == "" {
		response.BadRequest(c, exception.ErrBadRequest, "missing tenant context")
		return
	}
	column, ok := resettableQueueSettingColumn(table, field)
	if !ok {
		response.BadRequest(c, exception.ErrBadRequest, "invalid queue setting field")
		return
	}
	query := h.db.WithContext(c.Request.Context()).Table(table).Where("tenant_id = ? AND "+idColumn+" = ?", tenantID, id)
	if table == "branch_service_queue_settings" {
		query = query.Where("branch_id = ?", branchParam(c))
	}
	if err := query.UpdateColumn(column, nil).Error; err != nil {
		response.HandleError(c, err, "failed to reset queue setting")
		return
	}
	h.tryAudit(c.Request.Context(), "SETTING_RESET", table+":"+id+":"+field, map[string]string{"table": table, "field": field})
	c.Status(204)
}

func (h *SettingsController) EffectiveBranchConfig(c *gin.Context) {
	h.effectiveQueueConfigFor(c, model.EffectiveQueueConfigRequest{BranchID: branchParam(c)})
}

func (h *SettingsController) EffectiveBranchServiceConfig(c *gin.Context) {
	h.effectiveQueueConfigFor(c, model.EffectiveQueueConfigRequest{BranchID: branchParam(c), ServiceID: c.Param("service_id")})
}

func (h *SettingsController) EffectiveCounterConfig(c *gin.Context) {
	h.effectiveQueueConfigFor(c, model.EffectiveQueueConfigRequest{BranchID: branchParam(c), CounterID: c.Param("counter_id")})
}

func branchParam(c *gin.Context) string {
	if v := c.Param("branch_id"); v != "" {
		return v
	}
	return c.Param("id")
}

func (h *SettingsController) effectiveQueueConfigFor(c *gin.Context, req model.EffectiveQueueConfigRequest) {
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
	effectiveLogoAssetID := h.resolveEffectiveLogoAssetID(ctx, tenantID, req.BranchID)
	queueResetTimeResolved, _ := h.queueResolver.ResolveDetailed(ctx, "queue_reset_time", req.BranchID, req.ServiceID, req.CounterID)
	ticketPrefixResolved, _ := h.queueResolver.ResolveDetailed(ctx, "ticket_prefix", req.BranchID, req.ServiceID, req.CounterID)
	numberingStrategyResolved, _ := h.queueResolver.ResolveDetailed(ctx, "numbering_strategy", req.BranchID, req.ServiceID, req.CounterID)
	defaultEstimatedDurationResolved, _ := h.queueResolver.ResolveDetailed(ctx, "default_estimated_duration", req.BranchID, req.ServiceID, req.CounterID)
	autoCallNextPtr := parseBoolPtr(autoCallNext)

	res := &model.EffectiveQueueConfigResponse{
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
		AutoCallNext:                      autoCallNextPtr,
		Tenant: model.EffectiveConfigTenant{
			TenantID: tenantID,
		},
		Branch: model.EffectiveConfigBranch{
			BranchID:             req.BranchID,
			EffectiveLogoAssetID: effectiveLogoAssetID,
		},
		Queue: model.EffectiveConfigQueue{
			QueueResetTime:           safeResolved(queueResetTimeResolved),
			TicketPrefix:             safeResolved(ticketPrefixResolved),
			NumberingStrategy:        safeResolved(numberingStrategyResolved),
			DefaultEstimatedDuration: safeResolved(defaultEstimatedDurationResolved),
			AutoCallNext:             autoCallNextPtr,
		},
	}
	return res, nil
}

func resettableQueueSettingColumn(table, field string) (string, bool) {
	common := map[string]string{
		"queue_reset_time":           "queue_reset_time",
		"ticket_prefix":              "ticket_prefix",
		"default_estimated_duration": "default_estimated_duration",
		"allow_forward":              "allow_forward",
		"allow_skip":                 "allow_skip",
		"allow_recall":               "allow_recall",
		"allow_cancel":               "allow_cancel",
		"auto_call_next":             "auto_call_next",
		"numbering_strategy":         "numbering_strategy",
		"require_counter":            "require_counter",
		"allow_forward_from":         "allow_forward_from",
		"allow_forward_to":           "allow_forward_to",
	}
	column, ok := common[field]
	if !ok {
		return "", false
	}
	if table == "branch_service_queue_settings" {
		_, ok = map[string]struct{}{"default_estimated_duration": {}, "allow_skip": {}, "allow_recall": {}, "allow_cancel": {}, "auto_call_next": {}, "require_counter": {}, "allow_forward_from": {}, "allow_forward_to": {}}[field]
		return column, ok
	}
	return column, true
}

func (h *SettingsController) tryAudit(ctx context.Context, action, entityID string, values map[string]string) {
	if h.audit == nil {
		return
	}
	userID, _ := authcontext.UserIDFromContext(ctx)
	if userID == "" {
		userID = "system"
	}
	_ = h.audit.LogActivity(ctx, auditModel.CreateAuditLogRequest{UserID: userID, Action: action, Entity: "queue_setting", EntityID: entityID, NewValues: values})
}

func (h *SettingsController) resolveEffectiveLogoAssetID(ctx context.Context, tenantID, branchID string) string {
	if h.db == nil || tenantID == "" || branchID == "" {
		return ""
	}
	var row struct {
		BranchLogo string
		TenantLogo string
	}
	if err := h.db.WithContext(ctx).Table("branches").
		Select("branches.logo_asset_id AS branch_logo, organizations.logo_asset_id AS tenant_logo").
		Joins("JOIN organizations ON organizations.id = branches.tenant_id").
		Where("branches.id = ? AND branches.tenant_id = ?", branchID, tenantID).
		Take(&row).Error; err != nil {
		return ""
	}
	if row.BranchLogo != "" {
		return row.BranchLogo
	}
	return row.TenantLogo
}

func safeResolved(r *model.ResolvedQueueSetting) model.ResolvedQueueSetting {
	if r == nil {
		return model.ResolvedQueueSetting{}
	}
	return *r
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

func NewSettingsController(validate *validator.Validate, resolver QueueSettingResolver, log *logrus.Logger, db *gorm.DB, audit ...AuditLogger) *SettingsController {
	var gormDB *gorm.DB
	if db != nil {
		gormDB = db
	}
	var auditLogger AuditLogger
	if len(audit) > 0 {
		auditLogger = audit[0]
	}
	return &SettingsController{queueResolver: resolver, validate: validate, log: log, db: gormDB, audit: auditLogger}
}
