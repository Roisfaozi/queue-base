package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	"github.com/Roisfaozi/queue-base/internal/modules/counter/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newCounterTestValidator(t *testing.T) *validator.Validate {
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

func TestCounterController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Positive_Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockCounterUseCase
			wantCode int
		}{
			{
				name:    "Positive_CreateIncludesBranchID",
				reqBody: model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Front Desk"},
				setup: func(t *testing.T) *mocks.MockCounterUseCase {
					m := mocks.NewMockCounterUseCase(t)
					m.On("CreateCounter", mock.Anything, mock.MatchedBy(func(req *model.CreateCounterRequest) bool {
						return req != nil && req.BranchID == "550e8400-e29b-41d4-a716-446655440000"
					})).Return(&model.CounterResponse{ID: "counter-1", BranchID: "550e8400-e29b-41d4-a716-446655440000"}, nil).Once()
					return m
				},
				wantCode: http.StatusCreated,
			},
			{
				name:    "Negative_CreateRejectsInvalidBranchID",
				reqBody: map[string]interface{}{"branch_id": "bad-id", "code": "A1", "name": "Front Desk"},
				setup: func(t *testing.T) *mocks.MockCounterUseCase {
					return mocks.NewMockCounterUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewCounterController(uc, newCounterTestValidator(t), nil)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.POST("/counters", controller.Create)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("POST", "/counters", bytes.NewBuffer(body))
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
			setup    func(t *testing.T) *mocks.MockCounterUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetByIDReturnsBranchID",
				setup: func(t *testing.T) *mocks.MockCounterUseCase {
					m := mocks.NewMockCounterUseCase(t)
					m.On("GetCounter", mock.Anything, "counter-1").Return(&model.CounterResponse{ID: "counter-1", BranchID: "550e8400-e29b-41d4-a716-446655440000"}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"branch_id":"550e8400-e29b-41d4-a716-446655440000"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewCounterController(uc, newCounterTestValidator(t), nil)
				router := gin.New()
				router.GET("/counters/:id", controller.GetByID)

				req, _ := http.NewRequest("GET", "/counters/counter-1", nil)
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

	t.Run("Positive_Update", func(t *testing.T) {
		status := "inactive"
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockCounterUseCase
			wantCode int
		}{
			{
				name:    "Positive_UpdateCanToggleStatus",
				reqBody: model.UpdateCounterRequest{Status: &status},
				setup: func(t *testing.T) *mocks.MockCounterUseCase {
					m := mocks.NewMockCounterUseCase(t)
					m.On("UpdateCounter", mock.Anything, "counter-1", mock.MatchedBy(func(req *model.UpdateCounterRequest) bool {
						return req != nil && req.Status != nil && *req.Status == "inactive"
					})).Return(&model.CounterResponse{ID: "counter-1", Status: status}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewCounterController(uc, newCounterTestValidator(t), nil)
				router := gin.New()
				router.PUT("/counters/:id", controller.Update)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("PUT", "/counters/counter-1", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				uc.AssertExpectations(t)
			})
		}
	})

	t.Run("Positive_GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockCounterUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetAllReturnsCounters",
				setup: func(t *testing.T) *mocks.MockCounterUseCase {
					m := mocks.NewMockCounterUseCase(t)
					m.On("ListCounters", mock.Anything).Return([]model.CounterResponse{{ID: "counter-1", BranchID: "b-1", Code: "A1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"counter-1"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewCounterController(uc, newCounterTestValidator(t), nil)
				router := gin.New()
				router.GET("/counters", controller.GetAll)

				req, _ := http.NewRequest("GET", "/counters", nil)
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
			setup    func(t *testing.T) *mocks.MockCounterUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func(t *testing.T) *mocks.MockCounterUseCase {
					m := mocks.NewMockCounterUseCase(t)
					m.On("DeleteCounter", mock.Anything, "counter-1").Return(nil).Once()
					return m
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				controller := NewCounterController(uc, newCounterTestValidator(t), nil)
				router := gin.New()
				router.DELETE("/counters/:id", controller.Delete)

				req, _ := http.NewRequest("DELETE", "/counters/counter-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				uc.AssertExpectations(t)
			})
		}
	})
}
