package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubQueueResolver struct {
	values map[string]string
}

type stubSettingsAudit struct {
	requests []auditModel.CreateAuditLogRequest
}

func (s *stubSettingsAudit) LogActivity(_ context.Context, req auditModel.CreateAuditLogRequest) error {
	s.requests = append(s.requests, req)
	return nil
}

func (s stubQueueResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	return s.values[key], nil
}

func (s stubQueueResolver) ResolveDetailed(ctx context.Context, key string, branchID string, serviceID string, counterID string) (*model.ResolvedQueueSetting, error) {
	return &model.ResolvedQueueSetting{Key: key, Value: s.values[key], Source: "tenant", Inherited: branchID != "" || serviceID != "" || counterID != "", CanOverride: true, CanReset: branchID != "" || serviceID != "" || counterID != ""}, nil
}

func newSettingsTestValidator(t *testing.T) *validator.Validate {
	t.Helper()
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	require.NoError(t, validationpkg.RegisterCustomValidations(v))
	return v
}

func newSettingsControllerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE organizations (id TEXT PRIMARY KEY, logo_asset_id TEXT);
		CREATE TABLE branches (id TEXT PRIMARY KEY, tenant_id TEXT, logo_asset_id TEXT);
		CREATE TABLE branch_queue_settings (id TEXT PRIMARY KEY, tenant_id TEXT, branch_id TEXT, ticket_prefix TEXT, queue_reset_time TEXT, default_estimated_duration INTEGER, auto_call_next BOOLEAN);
		CREATE TABLE branch_service_queue_settings (id TEXT PRIMARY KEY, tenant_id TEXT, branch_id TEXT, branch_service_id TEXT, default_estimated_duration INTEGER, auto_call_next BOOLEAN);
		CREATE TABLE counter_queue_settings (id TEXT PRIMARY KEY, tenant_id TEXT, counter_id TEXT, ticket_prefix TEXT, queue_reset_time TEXT, default_estimated_duration INTEGER, auto_call_next BOOLEAN);
		INSERT INTO organizations (id, logo_asset_id) VALUES ('tenant-1', 'tenant-logo');
		INSERT INTO branches (id, tenant_id, logo_asset_id) VALUES ('550e8400-e29b-41d4-a716-446655440000', 'tenant-1', 'branch-logo');
		INSERT INTO branches (id, tenant_id, logo_asset_id) VALUES ('550e8400-e29b-41d4-a716-446655440001', 'tenant-1', '');
		INSERT INTO branch_queue_settings (id, tenant_id, branch_id, ticket_prefix) VALUES ('bqs-1', 'tenant-1', '550e8400-e29b-41d4-a716-446655440000', 'B');
		INSERT INTO branch_service_queue_settings (id, tenant_id, branch_id, branch_service_id, default_estimated_duration) VALUES ('bsqs-1', 'tenant-1', '550e8400-e29b-41d4-a716-446655440000', '550e8400-e29b-41d4-a716-446655440002', 9);
		INSERT INTO counter_queue_settings (id, tenant_id, counter_id, ticket_prefix) VALUES ('cqs-1', 'tenant-1', '550e8400-e29b-41d4-a716-446655440003', 'C');
	`).Error)
	return db
}

func TestSettingsController_EffectiveQueueConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)
	boolPtr := func(v bool) *bool { return &v }

	tests := []struct {
		name            string
		query           string
		tenantID        string
		wantCode        int
		wantAuto        *bool
		wantCanOverride bool
		wantLogo        string
	}{
		{
			name:            "Positive_ResolvesTypedConfig",
			query:           "?branch_id=550e8400-e29b-41d4-a716-446655440000",
			tenantID:        "tenant-1",
			wantCode:        http.StatusOK,
			wantAuto:        boolPtr(true),
			wantCanOverride: true,
			wantLogo:        "branch-logo",
		},
		{
			name:            "Edge_FallsBackToTenantLogoWhenBranchLogoEmpty",
			query:           "?branch_id=550e8400-e29b-41d4-a716-446655440001",
			tenantID:        "tenant-1",
			wantCode:        http.StatusOK,
			wantAuto:        boolPtr(true),
			wantCanOverride: true,
			wantLogo:        "tenant-logo",
		},
		{
			name:     "Negative_RejectsMissingTenantContext",
			query:    "?branch_id=550e8400-e29b-41d4-a716-446655440000",
			tenantID: "",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := NewSettingsController(newSettingsTestValidator(t), stubQueueResolver{values: map[string]string{
				"queue_reset_time":           "04:00",
				"ticket_prefix":              "A",
				"numbering_strategy":         "daily_branch_sequence",
				"default_estimated_duration": "5",
				"auto_call_next":             "true",
			}}, nil, newSettingsControllerTestDB(t))

			router := gin.New()
			router.GET("/settings/effective", func(c *gin.Context) {
				ctx := c.Request.Context()
				if tt.tenantID != "" {
					ctx = database.SetOrganizationContext(ctx, tt.tenantID)
				}
				c.Request = c.Request.WithContext(ctx)
				controller.EffectiveQueueConfig(c)
			})

			req, _ := http.NewRequest("GET", "/settings/effective"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantCode == http.StatusOK {
				var resp struct {
					Data model.EffectiveQueueConfigResponse `json:"data"`
				}
				require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
				assert.Equal(t, tt.wantAuto, resp.Data.AutoCallNext)
				assert.Equal(t, tt.wantAuto, resp.Data.Queue.AutoCallNext)
				assert.Equal(t, tt.wantCanOverride, resp.Data.Queue.QueueResetTime.CanOverride)
				assert.Equal(t, "tenant-1", resp.Data.Tenant.TenantID)
				assert.NotEmpty(t, resp.Data.Branch.BranchID)
				assert.Equal(t, tt.wantLogo, resp.Data.Branch.EffectiveLogoAssetID)
			}
		})
	}
}

func TestSettingsController_EffectiveConfigAliasPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewSettingsController(newSettingsTestValidator(t), stubQueueResolver{values: map[string]string{
		"queue_reset_time":           "04:00",
		"ticket_prefix":              "A",
		"numbering_strategy":         "daily_branch_sequence",
		"default_estimated_duration": "5",
		"auto_call_next":             "true",
	}}, nil, newSettingsControllerTestDB(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queue-config/effective", func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		controller.EffectiveQueueConfig(c)
	})
	router.GET("/branches/:branch_id/effective-config", controller.EffectiveBranchConfig)
	router.GET("/branches/:branch_id/services/:service_id/effective-config", controller.EffectiveBranchServiceConfig)
	router.GET("/branches/:branch_id/counters/:counter_id/effective-config", controller.EffectiveCounterConfig)

	tests := []struct {
		name string
		path string
	}{
		{name: "Positive_QueueConfigEffective", path: "/queue-config/effective?branch_id=550e8400-e29b-41d4-a716-446655440000"},
		{name: "Positive_BranchEffectiveConfig", path: "/branches/550e8400-e29b-41d4-a716-446655440000/effective-config"},
		{name: "Positive_ServiceEffectiveConfig", path: "/branches/550e8400-e29b-41d4-a716-446655440000/services/550e8400-e29b-41d4-a716-446655440002/effective-config"},
		{name: "Positive_CounterEffectiveConfig", path: "/branches/550e8400-e29b-41d4-a716-446655440000/counters/550e8400-e29b-41d4-a716-446655440003/effective-config"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), `"effective_logo_asset_id":"branch-logo"`)
		})
	}
}

func TestSettingsController_ResetQueueSetting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := newSettingsControllerTestDB(t)
	audit := &stubSettingsAudit{}
	controller := NewSettingsController(newSettingsTestValidator(t), stubQueueResolver{values: map[string]string{}}, nil, db, audit)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.DELETE("/branches/:branch_id/queue-settings/:field", controller.ResetBranchQueueSetting)
	router.DELETE("/branches/:branch_id/services/:branch_service_id/queue-settings/:field", controller.ResetBranchServiceQueueSetting)
	router.DELETE("/branches/:branch_id/counters/:counter_id/queue-settings/:field", controller.ResetCounterQueueSetting)

	tests := []struct {
		name      string
		path      string
		table     string
		column    string
		where     string
		whereArgs []any
	}{
		{name: "Positive_ResetBranchTicketPrefix", path: "/branches/550e8400-e29b-41d4-a716-446655440000/queue-settings/ticket_prefix", table: "branch_queue_settings", column: "ticket_prefix", where: "tenant_id = ? AND branch_id = ?", whereArgs: []any{"tenant-1", "550e8400-e29b-41d4-a716-446655440000"}},
		{name: "Positive_ResetBranchServiceDuration", path: "/branches/550e8400-e29b-41d4-a716-446655440000/services/550e8400-e29b-41d4-a716-446655440002/queue-settings/default_estimated_duration", table: "branch_service_queue_settings", column: "default_estimated_duration", where: "tenant_id = ? AND branch_service_id = ?", whereArgs: []any{"tenant-1", "550e8400-e29b-41d4-a716-446655440002"}},
		{name: "Positive_ResetCounterTicketPrefix", path: "/branches/550e8400-e29b-41d4-a716-446655440000/counters/550e8400-e29b-41d4-a716-446655440003/queue-settings/ticket_prefix", table: "counter_queue_settings", column: "ticket_prefix", where: "tenant_id = ? AND counter_id = ?", whereArgs: []any{"tenant-1", "550e8400-e29b-41d4-a716-446655440003"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNoContent, w.Code)
			var value *string
			require.NoError(t, db.Table(tt.table).Select(tt.column).Where(tt.where, tt.whereArgs...).Scan(&value).Error)
			assert.Nil(t, value)
		})
	}
	require.Len(t, audit.requests, len(tests))
	assert.Equal(t, "SETTING_RESET", audit.requests[0].Action)
}

func TestSettingsController_ResetQueueSetting_InvalidField(t *testing.T) {
	gin.SetMode(gin.TestMode)
	controller := NewSettingsController(newSettingsTestValidator(t), stubQueueResolver{values: map[string]string{}}, nil, newSettingsControllerTestDB(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.DELETE("/branches/:branch_id/queue-settings/:field", controller.ResetBranchQueueSetting)

	req, _ := http.NewRequest(http.MethodDelete, "/branches/550e8400-e29b-41d4-a716-446655440000/queue-settings/not_allowed", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
