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

	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestValidator(t *testing.T) *validator.Validate {
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

type stubServiceControllerUseCase struct {
	createReq *model.CreateServiceRequest
	updateReq *model.UpdateServiceRequest
	createRes *model.ServiceResponse
	updateRes *model.ServiceResponse
}

func (s *stubServiceControllerUseCase) CreateService(ctx context.Context, req *model.CreateServiceRequest) (*model.ServiceResponse, error) {
	s.createReq = req
	return s.createRes, nil
}

func (s *stubServiceControllerUseCase) GetService(ctx context.Context, serviceID string) (*model.ServiceResponse, error) {
	return nil, nil
}

func (s *stubServiceControllerUseCase) ListServices(ctx context.Context) ([]model.ServiceResponse, error) {
	return nil, nil
}

func (s *stubServiceControllerUseCase) UpdateService(ctx context.Context, serviceID string, req *model.UpdateServiceRequest) (*model.ServiceResponse, error) {
	s.updateReq = req
	return s.updateRes, nil
}

func (s *stubServiceControllerUseCase) DeleteService(ctx context.Context, serviceID string) error {
	return nil
}

func TestServiceController_CreateIncludesPharmacyFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubServiceControllerUseCase{createRes: &model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: true}}
	controller := NewServiceController(uc, newTestValidator(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/services", controller.Create)

	body, err := json.Marshal(model.CreateServiceRequest{Code: "pha", Name: "Pharmacy", IsPharmacy: true, IsPharmacyReception: true})
	require.NoError(t, err)
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, uc.createReq)
	assert.True(t, uc.createReq.IsPharmacy)
	assert.True(t, uc.createReq.IsPharmacyReception)
}

func TestServiceController_UpdateCanTogglePharmacyFlags(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubServiceControllerUseCase{updateRes: &model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: false}}
	controller := NewServiceController(uc, newTestValidator(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.PUT("/services/:id", controller.Update)

	flag := false
	body, err := json.Marshal(model.UpdateServiceRequest{IsPharmacyReception: &flag})
	require.NoError(t, err)
	req, _ := http.NewRequest("PUT", "/services/svc-1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, uc.updateReq)
	require.NotNil(t, uc.updateReq.IsPharmacyReception)
	assert.False(t, *uc.updateReq.IsPharmacyReception)
}
