package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/service/model"
	"github.com/Roisfaozi/queue-base/internal/modules/service/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestServiceController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Positive_Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockServiceUseCase
			wantCode int
		}{
			{
				name:    "Positive_CreateIncludesPharmacyFlags",
				reqBody: model.CreateServiceRequest{Code: "pha", Name: "Pharmacy", IsPharmacy: true, IsPharmacyReception: true},
				setup: func(t *testing.T) *mocks.MockServiceUseCase {
					m := mocks.NewMockServiceUseCase(t)
					m.On("CreateService", mock.Anything, mock.MatchedBy(func(req *model.CreateServiceRequest) bool {
						return req != nil && req.IsPharmacy && req.IsPharmacyReception
					})).Return(&model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: true}, nil).Once()
					return m
				},
				wantCode: http.StatusCreated,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewServiceController(uc, newTestValidator(t), nil)
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
				uc.AssertExpectations(t)
			})
		}
	})

	t.Run("Positive_Update", func(t *testing.T) {
		flag := false
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockServiceUseCase
			wantCode int
		}{
			{
				name:    "Positive_UpdateCanTogglePharmacyFlags",
				reqBody: model.UpdateServiceRequest{IsPharmacyReception: &flag},
				setup: func(t *testing.T) *mocks.MockServiceUseCase {
					m := mocks.NewMockServiceUseCase(t)
					m.On("UpdateService", mock.Anything, "svc-1", mock.MatchedBy(func(req *model.UpdateServiceRequest) bool {
						return req != nil && req.IsPharmacyReception != nil && !*req.IsPharmacyReception
					})).Return(&model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: false}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewServiceController(uc, newTestValidator(t), nil)
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
				uc.AssertExpectations(t)
			})
		}
	})

	t.Run("Positive_GetByID", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockServiceUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetByIDReturnsPharmacyFlags",
				setup: func(t *testing.T) *mocks.MockServiceUseCase {
					m := mocks.NewMockServiceUseCase(t)
					m.On("GetService", mock.Anything, "svc-1").Return(&model.ServiceResponse{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: true}, nil).Once()
					return m
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
				uc := tt.setup(t)
				controller := NewServiceController(uc, newTestValidator(t), nil)
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
				uc.AssertExpectations(t)
			})
		}
	})

	t.Run("Positive_GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockServiceUseCase
			tenantID string
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetAllReturnsPharmacyFlags",
				setup: func(t *testing.T) *mocks.MockServiceUseCase {
					m := mocks.NewMockServiceUseCase(t)
					m.On("ListServices", mock.Anything).Return([]model.ServiceResponse{{ID: "svc-1", IsPharmacy: true, IsPharmacyReception: false}}, nil).Once()
					return m
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
				setup: func(t *testing.T) *mocks.MockServiceUseCase {
					return mocks.NewMockServiceUseCase(t)
				},
				tenantID: "",
				wantCode: http.StatusBadRequest,
				assert:   nil,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewServiceController(uc, newTestValidator(t), nil)
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
				uc.AssertExpectations(t)
			})
		}
	})

	t.Run("Positive_Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockServiceUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func(t *testing.T) *mocks.MockServiceUseCase {
					m := mocks.NewMockServiceUseCase(t)
					m.On("DeleteService", mock.Anything, "svc-1").Return(nil).Once()
					return m
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewServiceController(uc, newTestValidator(t), nil)
				router := gin.New()
				router.DELETE("/services/:id", controller.Delete)

				req, _ := http.NewRequest("DELETE", "/services/svc-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				uc.AssertExpectations(t)
			})
		}
	})

	t.Run("Positive_EnableDisable", func(t *testing.T) {
		tests := []struct {
			name       string
			methodPath string
			wantStatus string
		}{
			{name: "Positive_EnableSetsActive", methodPath: "/services/svc-1/enable", wantStatus: "active"},
			{name: "Positive_DisableSetsInactive", methodPath: "/services/svc-1/disable", wantStatus: "inactive"},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := mocks.NewMockServiceUseCase(t)
				status := tt.wantStatus
				uc.On("UpdateService", mock.Anything, "svc-1", mock.MatchedBy(func(req *model.UpdateServiceRequest) bool {
					return req != nil && req.Status != nil && *req.Status == tt.wantStatus
				})).Return(&model.ServiceResponse{ID: "svc-1", Status: status}, nil).Once()

				controller := NewServiceController(uc, newTestValidator(t), nil)
				router := gin.New()
				router.POST("/services/:id/enable", controller.Enable)
				router.POST("/services/:id/disable", controller.Disable)

				req, _ := http.NewRequest("POST", tt.methodPath, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusOK, w.Code)
				uc.AssertExpectations(t)
			})
		}
	})
}
