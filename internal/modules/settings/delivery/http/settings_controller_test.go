package http

import (
	"bytes"
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

type stubSettingsControllerUseCase struct {
	createReq  *model.CreateSettingRequest
	resolveReq *model.ResolveSettingRequest
	createRes  *model.SettingResponse
	resolveRes *model.SettingResponse
}

func (s *stubSettingsControllerUseCase) CreateSetting(ctx context.Context, req *model.CreateSettingRequest) (*model.SettingResponse, error) {
	s.createReq = req
	return s.createRes, nil
}

func (s *stubSettingsControllerUseCase) GetSetting(ctx context.Context, settingID string) (*model.SettingResponse, error) {
	return nil, nil
}

func (s *stubSettingsControllerUseCase) UpdateSetting(ctx context.Context, settingID string, req *model.UpdateSettingRequest) (*model.SettingResponse, error) {
	return nil, nil
}

func (s *stubSettingsControllerUseCase) DeleteSetting(ctx context.Context, settingID string) error {
	return nil
}

func (s *stubSettingsControllerUseCase) ResolveSetting(ctx context.Context, req *model.ResolveSettingRequest) (*model.SettingResponse, error) {
	s.resolveReq = req
	return s.resolveRes, nil
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

func TestSettingsController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubSettingsControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubSettingsControllerUseCase)
		}{
			{
				name:    "Positive_CreateWorkflowSetting",
				reqBody: model.CreateSettingRequest{ScopeType: "service", ScopeID: "550e8400-e29b-41d4-a716-446655440000", Key: model.SettingKeyPharmacyFlowEnabled, Value: "true", ValueType: "boolean"},
				setup: func() *stubSettingsControllerUseCase {
					return &stubSettingsControllerUseCase{createRes: &model.SettingResponse{ID: "set-1", Key: model.SettingKeyPharmacyFlowEnabled, Value: "true"}}
				},
				wantCode: http.StatusCreated,
				assert: func(t *testing.T, uc *stubSettingsControllerUseCase) {
					require.NotNil(t, uc.createReq)
					assert.Equal(t, model.SettingKeyPharmacyFlowEnabled, uc.createReq.Key)
					assert.Equal(t, "service", uc.createReq.ScopeType)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewSettingsController(uc, newSettingsTestValidator(t))
				router := gin.New()
				router.POST("/settings", controller.Create)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("POST", "/settings", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("Resolve", func(t *testing.T) {
		tests := []struct {
			name     string
			query    string
			setup    func() *stubSettingsControllerUseCase
			tenantID string
			wantCode int
			assert   func(t *testing.T, uc *stubSettingsControllerUseCase)
		}{
			{
				name:  "Positive_ResolveWorkflowSetting",
				query: "?Key=require_counter_for_service",
				setup: func() *stubSettingsControllerUseCase {
					return &stubSettingsControllerUseCase{resolveRes: &model.SettingResponse{ID: "set-1", Key: model.SettingKeyRequireCounterForService, Value: "true"}}
				},
				tenantID: "tenant-1",
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubSettingsControllerUseCase) {
					require.NotNil(t, uc.resolveReq)
					assert.Equal(t, model.SettingKeyRequireCounterForService, uc.resolveReq.Key)
				},
			},
			{
				name:  "Negative_ResolveRejectsInvalidWorkflowScopeIDs",
				query: "?Key=pharmacy_flow_enabled&ServiceID=bad-id",
				setup: func() *stubSettingsControllerUseCase {
					return &stubSettingsControllerUseCase{}
				},
				tenantID: "tenant-1",
				wantCode: http.StatusUnprocessableEntity,
				assert: func(t *testing.T, uc *stubSettingsControllerUseCase) {
					assert.Nil(t, uc.resolveReq)
				},
			},
			{
				name:  "Negative_ResolveRejectsMissingTenantContext",
				query: "?Key=reset_time",
				setup: func() *stubSettingsControllerUseCase {
					return &stubSettingsControllerUseCase{}
				},
				tenantID: "",
				wantCode: http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubSettingsControllerUseCase) {
					assert.Nil(t, uc.resolveReq)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewSettingsController(uc, newSettingsTestValidator(t))
				router := gin.New()
				if tt.tenantID != "" {
					router.Use(func(c *gin.Context) {
						ctx := database.SetOrganizationContext(c.Request.Context(), tt.tenantID)
						c.Request = c.Request.WithContext(ctx)
						c.Next()
					})
				}
				router.GET("/settings/resolve", controller.Resolve)

				req, _ := http.NewRequest("GET", "/settings/resolve"+tt.query, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubSettingsControllerUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func() *stubSettingsControllerUseCase {
					return &stubSettingsControllerUseCase{}
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewSettingsController(uc, newSettingsTestValidator(t))
				router := gin.New()
				router.DELETE("/settings/:id", controller.Delete)

				req, _ := http.NewRequest("DELETE", "/settings/set-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
			})
		}
	})
}
