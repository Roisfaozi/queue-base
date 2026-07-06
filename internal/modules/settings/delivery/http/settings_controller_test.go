package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubQueueResolver struct {
	values map[string]string
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
	}{
		{
			name:            "Positive_ResolvesTypedConfig",
			query:           "?branch_id=550e8400-e29b-41d4-a716-446655440000",
			tenantID:        "tenant-1",
			wantCode:        http.StatusOK,
			wantAuto:        boolPtr(true),
			wantCanOverride: true,
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
			}}, nil)

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
				assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", resp.Data.Branch.BranchID)
			}
		})
	}
}
