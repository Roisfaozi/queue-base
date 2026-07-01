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
	getRes    *model.ServiceResponse
	listRes   []model.ServiceResponse
}

func (s *stubServiceControllerUseCase) CreateService(ctx context.Context, req *model.CreateServiceRequest) (*model.ServiceResponse, error) {
	s.createReq = req
	return s.createRes, nil
}

func (s *stubServiceControllerUseCase) GetService(ctx context.Context, serviceID string) (*model.ServiceResponse, error) {
	return s.getRes, nil
}

func (s *stubServiceControllerUseCase) ListServices(ctx context.Context) ([]model.ServiceResponse, error) {
	return s.listRes, nil
}

func (s *stubServiceControllerUseCase) UpdateService(ctx context.Context, serviceID string, req *model.UpdateServiceRequest) (*model.ServiceResponse, error) {
	s.updateReq = req
	return s.updateRes, nil
}

func (s *stubServiceControllerUseCase) DeleteService(ctx context.Context, serviceID string) error {
	return nil
}

func TestServiceController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubServiceControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubServiceControllerUseCase)
		}{
			{
				name:    "Positive_CreateIncludesPharmacyFlags",
				reqBody: model.CreateServiceRequest{Code: "pha", Name: "Pharmacy", IsPharmacy: true, IsPharmacyReception: true},
				setup: func() *stubServiceControllerUseCase {
					return &stubServiceControllerUseCase{createRes: &model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: true}}
				},
				wantCode: http.StatusCreated,
				assert: func(t *testing.T, uc *stubServiceControllerUseCase) {
					require.NotNil(t, uc.createReq)
					assert.True(t, uc.createReq.IsPharmacy)
					assert.True(t, uc.createReq.IsPharmacyReception)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewServiceController(uc, newTestValidator(t))
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.POST("/services", controller.Create)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		flag := false
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubServiceControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubServiceControllerUseCase)
		}{
			{
				name:    "Positive_UpdateCanTogglePharmacyFlags",
				reqBody: model.UpdateServiceRequest{IsPharmacyReception: &flag},
				setup: func() *stubServiceControllerUseCase {
					return &stubServiceControllerUseCase{updateRes: &model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: false}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubServiceControllerUseCase) {
					require.NotNil(t, uc.updateReq)
					require.NotNil(t, uc.updateReq.IsPharmacyReception)
					assert.False(t, *uc.updateReq.IsPharmacyReception)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewServiceController(uc, newTestValidator(t))
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.PUT("/services/:id", controller.Update)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("PUT", "/services/svc-1", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubServiceControllerUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetByIDReturnsPharmacyFlags",
				setup: func() *stubServiceControllerUseCase {
					return &stubServiceControllerUseCase{getRes: &model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: true}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"is_pharmacy":true`)
					assert.Contains(t, body, `"is_pharmacy_reception":true`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewServiceController(uc, newTestValidator(t))
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/services/:id", controller.GetByID)

				req, _ := http.NewRequest("GET", "/services/svc-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, w.Body.String())
				}
			})
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubServiceControllerUseCase
			tenantID string
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetAllReturnsPharmacyFlags",
				setup: func() *stubServiceControllerUseCase {
					return &stubServiceControllerUseCase{listRes: []model.ServiceResponse{{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: false}}}
				},
				tenantID: "tenant-1",
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"is_pharmacy":true`)
					assert.Contains(t, body, `"is_pharmacy_reception":false`)
				},
			},
			{
				name: "Negative_GetAllRejectsMissingTenantContext",
				setup: func() *stubServiceControllerUseCase {
					return &stubServiceControllerUseCase{}
				},
				tenantID: "",
				wantCode: http.StatusBadRequest,
				assert:   nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewServiceController(uc, newTestValidator(t))
				router := gin.New()
				if tt.tenantID != "" {
					router.Use(func(c *gin.Context) {
						ctx := database.SetOrganizationContext(c.Request.Context(), tt.tenantID)
						c.Request = c.Request.WithContext(ctx)
						c.Next()
					})
				}
				router.GET("/services", controller.GetAll)

				req, _ := http.NewRequest("GET", "/services", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, w.Body.String())
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubServiceControllerUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func() *stubServiceControllerUseCase {
					return &stubServiceControllerUseCase{}
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewServiceController(uc, newTestValidator(t))
				router := gin.New()
				router.DELETE("/services/:id", controller.Delete)

				req, _ := http.NewRequest("DELETE", "/services/svc-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
			})
		}
	})
}
