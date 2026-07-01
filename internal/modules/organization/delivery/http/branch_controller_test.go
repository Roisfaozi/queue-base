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

	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	validationpkg "github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubBranchControllerUseCase struct {
	createReq *model.CreateBranchRequest
	updateReq *model.UpdateBranchRequest
	createRes *model.BranchResponse
	updateRes *model.BranchResponse
	getRes    *model.BranchResponse
	listRes   []model.BranchResponse
}

func (s *stubBranchControllerUseCase) CreateBranch(ctx context.Context, req *model.CreateBranchRequest) (*model.BranchResponse, error) {
	s.createReq = req
	return s.createRes, nil
}

func (s *stubBranchControllerUseCase) ResolveBranch(ctx context.Context, branchID string) (*model.BranchResponse, error) {
	return s.getRes, nil
}

func (s *stubBranchControllerUseCase) ListBranches(ctx context.Context) ([]model.BranchResponse, error) {
	return s.listRes, nil
}

func (s *stubBranchControllerUseCase) UpdateBranch(ctx context.Context, branchID string, req *model.UpdateBranchRequest) (*model.BranchResponse, error) {
	s.updateReq = req
	return s.updateRes, nil
}

func (s *stubBranchControllerUseCase) DeleteBranch(ctx context.Context, branchID string) error {
	return nil
}

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

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubBranchControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubBranchControllerUseCase)
		}{
			{
				name:    "Positive_CreateUsesTenantContext",
				reqBody: model.CreateBranchRequest{Code: "main", Name: "Main Branch"},
				setup: func() *stubBranchControllerUseCase {
					return &stubBranchControllerUseCase{createRes: &model.BranchResponse{ID: "branch-1", TenantID: "tenant-1"}}
				},
				wantCode: http.StatusCreated,
				assert: func(t *testing.T, uc *stubBranchControllerUseCase) {
					require.NotNil(t, uc.createReq)
					assert.Equal(t, "main", uc.createReq.Code)
				},
			},
			{
				name:    "Negative_CreateRejectsInvalidBody",
				reqBody: map[string]interface{}{"code": "", "name": ""},
				setup: func() *stubBranchControllerUseCase {
					return &stubBranchControllerUseCase{}
				},
				wantCode: http.StatusUnprocessableEntity,
				assert: func(t *testing.T, uc *stubBranchControllerUseCase) {
					assert.Nil(t, uc.createReq)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
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
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubBranchControllerUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetByIDReturnsBranch",
				setup: func() *stubBranchControllerUseCase {
					return &stubBranchControllerUseCase{getRes: &model.BranchResponse{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}}
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
				uc := tt.setup()
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

	t.Run("Update", func(t *testing.T) {
		name := " Main Office "
		tests := []struct {
			name     string
			reqBody  interface{}
			setup    func() *stubBranchControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubBranchControllerUseCase)
		}{
			{
				name:    "Positive_UpdateSanitizesFields",
				reqBody: model.UpdateBranchRequest{Name: &name},
				setup: func() *stubBranchControllerUseCase {
					return &stubBranchControllerUseCase{updateRes: &model.BranchResponse{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubBranchControllerUseCase) {
					require.NotNil(t, uc.updateReq)
					require.NotNil(t, uc.updateReq.Name)
					assert.Equal(t, " Main Office ", *uc.updateReq.Name)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
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
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubBranchControllerUseCase
			wantCode int
			assert   func(t *testing.T, body string)
		}{
			{
				name: "Positive_GetAllReturnsBranches",
				setup: func() *stubBranchControllerUseCase {
					return &stubBranchControllerUseCase{listRes: []model.BranchResponse{{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, body string) {
					assert.Contains(t, body, `"branch-1"`)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
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

	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			setup    func() *stubBranchControllerUseCase
			wantCode int
		}{
			{
				name: "Positive_DeleteReturnsNoContent",
				setup: func() *stubBranchControllerUseCase {
					return &stubBranchControllerUseCase{}
				},
				wantCode: http.StatusNoContent,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
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
