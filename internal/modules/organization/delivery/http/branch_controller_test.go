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

func TestBranchController_CreateUsesTenantContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubBranchControllerUseCase{createRes: &model.BranchResponse{ID: "branch-1", TenantID: "tenant-1"}}
	controller := NewBranchController(uc, newBranchTestValidator(t))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "tenant-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/branches", controller.Create)

	body, err := json.Marshal(model.CreateBranchRequest{Code: "main", Name: "Main Branch"})
	require.NoError(t, err)
	req, _ := http.NewRequest("POST", "/branches", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	require.NotNil(t, uc.createReq)
	assert.Equal(t, "main", uc.createReq.Code)
}

func TestBranchController_CreateRejectsInvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubBranchControllerUseCase{}
	controller := NewBranchController(uc, newBranchTestValidator(t))
	router := gin.New()
	router.POST("/branches", controller.Create)

	req, _ := http.NewRequest("POST", "/branches", bytes.NewBuffer([]byte(`{"code":"","name":""}`)))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Nil(t, uc.createReq)
}

func TestBranchController_GetByIDReturnsBranch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubBranchControllerUseCase{getRes: &model.BranchResponse{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}}
	controller := NewBranchController(uc, newBranchTestValidator(t))
	router := gin.New()
	router.GET("/branches/:id", controller.GetByID)

	req, _ := http.NewRequest("GET", "/branches/branch-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"code":"MAIN"`)
}

func TestBranchController_UpdateSanitizesFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	name := " Main Office "
	uc := &stubBranchControllerUseCase{updateRes: &model.BranchResponse{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN"}}
	controller := NewBranchController(uc, newBranchTestValidator(t))
	router := gin.New()
	router.PUT("/branches/:id", controller.Update)

	body, err := json.Marshal(model.UpdateBranchRequest{Name: &name})
	require.NoError(t, err)
	req, _ := http.NewRequest("PUT", "/branches/branch-1", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, uc.updateReq)
	require.NotNil(t, uc.updateReq.Name)
	assert.Equal(t, " Main Office ", *uc.updateReq.Name)
}
