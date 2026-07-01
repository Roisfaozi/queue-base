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

	"github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
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

type stubCounterControllerUseCase struct {
	createReq *model.CreateCounterRequest
	updateReq *model.UpdateCounterRequest
	createRes *model.CounterResponse
	updateRes *model.CounterResponse
	getRes    *model.CounterResponse
	listRes   []model.CounterResponse
}

func (s *stubCounterControllerUseCase) CreateCounter(ctx context.Context, req *model.CreateCounterRequest) (*model.CounterResponse, error) {
	s.createReq = req
	return s.createRes, nil
}

func (s *stubCounterControllerUseCase) GetCounter(ctx context.Context, counterID string) (*model.CounterResponse, error) {
	return s.getRes, nil
}

func (s *stubCounterControllerUseCase) ListCounters(ctx context.Context) ([]model.CounterResponse, error) {
	return s.listRes, nil
}

func (s *stubCounterControllerUseCase) UpdateCounter(ctx context.Context, counterID string, req *model.UpdateCounterRequest) (*model.CounterResponse, error) {
	s.updateReq = req
	return s.updateRes, nil
}

func (s *stubCounterControllerUseCase) DeleteCounter(ctx context.Context, counterID string) error {
	return nil
}

func TestCounterController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubCounterControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubCounterControllerUseCase)
		}{
			{
				name:    "Positive_CreateIncludesBranchID",
				reqBody: model.CreateCounterRequest{BranchID: "550e8400-e29b-41d4-a716-446655440000", Code: "A1", Name: "Front Desk"},
				setup: func() *stubCounterControllerUseCase {
					return &stubCounterControllerUseCase{createRes: &model.CounterResponse{ID: "counter-1", BranchID: "550e8400-e29b-41d4-a716-446655440000"}}
				},
				wantCode: http.StatusCreated,
				assert: func(t *testing.T, uc *stubCounterControllerUseCase) {
					require.NotNil(t, uc.createReq)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uc.createReq.BranchID)
				},
			},
			{
				name:    "Negative_CreateRejectsInvalidBranchID",
				reqBody: map[string]interface{}{"branch_id": "bad-id", "code": "A1", "name": "Front Desk"},
				setup: func() *stubCounterControllerUseCase {
					return &stubCounterControllerUseCase{}
				},
				wantCode: http.StatusUnprocessableEntity,
				assert: func(t *testing.T, uc *stubCounterControllerUseCase) {
					assert.Nil(t, uc.createReq)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewCounterController(uc, newCounterTestValidator(t))
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
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubCounterControllerUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetByIDReturnsBranchID",
				setup: func() *stubCounterControllerUseCase {
					return &stubCounterControllerUseCase{getRes: &model.CounterResponse{ID: "counter-1", BranchID: "550e8400-e29b-41d4-a716-446655440000"}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"branch_id":"550e8400-e29b-41d4-a716-446655440000"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewCounterController(uc, newCounterTestValidator(t))
				router := gin.New()
				router.GET("/counters/:id", controller.GetByID)

				req, _ := http.NewRequest("GET", "/counters/counter-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, w.Body.String())
				}
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		status := "inactive"
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubCounterControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubCounterControllerUseCase)
		}{
			{
				name:    "Positive_UpdateCanToggleStatus",
				reqBody: model.UpdateCounterRequest{Status: &status},
				setup: func() *stubCounterControllerUseCase {
					return &stubCounterControllerUseCase{updateRes: &model.CounterResponse{ID: "counter-1", Status: status}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubCounterControllerUseCase) {
					require.NotNil(t, uc.updateReq)
					require.NotNil(t, uc.updateReq.Status)
					assert.Equal(t, "inactive", *uc.updateReq.Status)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewCounterController(uc, newCounterTestValidator(t))
				router := gin.New()
				router.PUT("/counters/:id", controller.Update)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("PUT", "/counters/counter-1", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubCounterControllerUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetAllReturnsCounters",
				setup: func() *stubCounterControllerUseCase {
					return &stubCounterControllerUseCase{listRes: []model.CounterResponse{{ID: "counter-1", BranchID: "b-1", Code: "A1"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"counter-1"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewCounterController(uc, newCounterTestValidator(t))
				router := gin.New()
				router.GET("/counters", controller.GetAll)

				req, _ := http.NewRequest("GET", "/counters", nil)
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
			setup    func() *stubCounterControllerUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func() *stubCounterControllerUseCase {
					return &stubCounterControllerUseCase{}
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				controller := NewCounterController(uc, newCounterTestValidator(t))
				router := gin.New()
				router.DELETE("/counters/:id", controller.Delete)

				req, _ := http.NewRequest("DELETE", "/counters/counter-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
			})
		}
	})
}
