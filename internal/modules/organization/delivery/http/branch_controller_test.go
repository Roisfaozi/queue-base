package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newBranchTestValidator(t *testing.T) *validator.Validate {
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

func TestBranchController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Positive_Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockBranchUseCase
			wantCode int
		}{
			{
				name:    "Positive_CreateUsesTenantContext",
				reqBody: model.CreateBranchRequest{Code: "main", Name: "Main Branch"},
				setup: func(t *testing.T) *mocks.MockBranchUseCase {
					m := mocks.NewMockBranchUseCase(t)
					m.On("CreateBranch", mock.Anything, mock.MatchedBy(func(req *model.CreateBranchRequest) bool {
						return req != nil && req.Code == "main"
					})).Return(&model.BranchResponse{ID: "branch-1", TenantID: "tenant-1"}, nil).Once()
					return m
				},
				wantCode: http.StatusCreated,
			},
			{
				name:    "Negative_CreateRejectsInvalidBody",
				reqBody: map[string]interface{}{"code": "", "name": ""},
				setup: func(t *testing.T) *mocks.MockBranchUseCase {
					return mocks.NewMockBranchUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
				controller := NewBranchController(uc, newBranchTestValidator(t), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.POST("/branches", controller.Create)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("POST", "/branches", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
			})
		}
	})

	t.Run("Positive_GetByID", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockBranchUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetByIDReturnsBranch",
				setup: func(t *testing.T) *mocks.MockBranchUseCase {
					m := mocks.NewMockBranchUseCase(t)
					m.On("ResolveBranch", mock.Anything, "branch-1").
						Return(&model.BranchResponse{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"code":"MAIN"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				log := logrus.New()
				uc := tt.setup(t)
				controller := NewBranchController(uc, newBranchTestValidator(t), log)
				router := gin.New()
				router.GET("/branches/:id", controller.GetByID)

				req, _ := http.NewRequest("GET", "/branches/branch-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, w.Body.String())
				}
			})
		}
	})

	t.Run("Positive_Update", func(t *testing.T) {
		name := " Main Office "
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockBranchUseCase
			wantCode int
		}{
			{
				name:    "Positive_UpdateSanitizesFields",
				reqBody: model.UpdateBranchRequest{Name: &name},
				setup: func(t *testing.T) *mocks.MockBranchUseCase {
					m := mocks.NewMockBranchUseCase(t)
					m.On("UpdateBranch", mock.Anything, "branch-1", mock.MatchedBy(func(req *model.UpdateBranchRequest) bool {
						return req != nil && req.Name != nil && *req.Name == " Main Office "
					})).Return(&model.BranchResponse{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()

				controller := NewBranchController(uc, newBranchTestValidator(t), log)
				router := gin.New()
				router.PUT("/branches/:id", controller.Update)

				body, err := json.Marshal(tt.reqBody)
				require.NoError(t, err)
				req, _ := http.NewRequest("PUT", "/branches/branch-1", bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
			})
		}
	})

	t.Run("Positive_GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockBranchUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetAllReturnsBranches",
				setup: func(t *testing.T) *mocks.MockBranchUseCase {
					m := mocks.NewMockBranchUseCase(t)
					m.On("ListBranches", mock.Anything).
						Return([]model.BranchResponse{{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"branch-1"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()

				controller := NewBranchController(uc, newBranchTestValidator(t), log)
				router := gin.New()
				router.GET("/branches", controller.GetAll)

				req, _ := http.NewRequest("GET", "/branches", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, w.Body.String())
				}
			})
		}
	})

	t.Run("Positive_Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func(t *testing.T) *mocks.MockBranchUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func(t *testing.T) *mocks.MockBranchUseCase {
					m := mocks.NewMockBranchUseCase(t)
					m.On("DeleteBranch", mock.Anything, "branch-1").Return(nil).Once()
					return m
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()

				controller := NewBranchController(uc, newBranchTestValidator(t), log)
				router := gin.New()
				router.DELETE("/branches/:id", controller.Delete)

				req, _ := http.NewRequest("DELETE", "/branches/branch-1", nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
			})
		}
	})
}
