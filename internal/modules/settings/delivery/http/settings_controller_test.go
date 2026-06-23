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

func TestSettingsController_CreateWorkflowSetting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubSettingsControllerUseCase{createRes: &model.SettingResponse{ID: "set-1", Key: model.SettingKeyPharmacyFlowEnabled, Value: "true"}}
	controller := NewSettingsController(uc, newSettingsTestValidator(t))
	router := gin.New()
	router.POST("/settings", controller.Create)

	body, _ := json.Marshal(model.CreateSettingRequest{ScopeType: "service", ScopeID: "550e8400-e29b-41d4-a716-446655440000", Key: model.SettingKeyPharmacyFlowEnabled, Value: "true", ValueType: "boolean"})
	req, _ := http.NewRequest("POST", "/settings", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, uc.createReq)
	assert.Equal(t, model.SettingKeyPharmacyFlowEnabled, uc.createReq.Key)
	assert.Equal(t, "service", uc.createReq.ScopeType)
}

func TestSettingsController_ResolveWorkflowSetting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubSettingsControllerUseCase{resolveRes: &model.SettingResponse{ID: "set-1", Key: model.SettingKeyRequireCounterForService, Value: "true"}}
	controller := NewSettingsController(uc, newSettingsTestValidator(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/settings/resolve", controller.Resolve)

	req, _ := http.NewRequest("GET", "/settings/resolve?Key=require_counter_for_service", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, uc.resolveReq)
	assert.Equal(t, model.SettingKeyRequireCounterForService, uc.resolveReq.Key)
}

func TestSettingsController_ResolveRejectsInvalidWorkflowScopeIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubSettingsControllerUseCase{}
	controller := NewSettingsController(uc, newSettingsTestValidator(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/settings/resolve", controller.Resolve)

	req, _ := http.NewRequest("GET", "/settings/resolve?Key=pharmacy_flow_enabled&ServiceID=bad-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Nil(t, uc.resolveReq)
}

func TestSettingsController_DeleteReturnsNoContent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubSettingsControllerUseCase{}
	controller := NewSettingsController(uc, newSettingsTestValidator(t))
	router := gin.New()
	router.DELETE("/settings/:id", controller.Delete)

	req, _ := http.NewRequest("DELETE", "/settings/set-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestSettingsController_ResolveRejectsMissingTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubSettingsControllerUseCase{}
	controller := NewSettingsController(uc, newSettingsTestValidator(t))
	router := gin.New()
	router.GET("/settings/resolve", controller.Resolve)

	req, _ := http.NewRequest("GET", "/settings/resolve?Key=reset_time", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
