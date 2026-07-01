package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	permHandler "github.com/Roisfaozi/queue-base/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/model"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupPermissionTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func newTestPermissionController(mockUseCase *mocks.MockIPermissionUseCase) *permHandler.PermissionController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return permHandler.NewPermissionController(mockUseCase, log, v)
}

func TestGrantPermission(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				mockUseCase := new(mocks.MockIPermissionUseCase)
				handler := newTestPermissionController(mockUseCase)
				router := setupPermissionTestRouter()
				router.POST("/permissions/grant", handler.GrantPermission)

				reqBody := model.GrantPermissionRequest{Role: "editor", Path: "/articles", Method: "POST", Domain: "global"}
				mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(nil).Once()

				body := `{"role":"editor","path":"/articles","method":"POST","domain":"global"}`
				req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusCreated, w.Code)
				var responseBody map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				data, ok := responseBody["data"].(map[string]interface{})
				assert.True(t, ok, "Response should have a 'data' object")
				assert.Equal(t, "permission granted successfully", data["message"])
				mockUseCase.AssertExpectations(t)
			},
		},
		{
			name:     "Negative_InvalidBody",
			category: "negative",
			run: func(t *testing.T) {
				mockUseCase := new(mocks.MockIPermissionUseCase)
				handler := newTestPermissionController(mockUseCase)
				router := setupPermissionTestRouter()
				router.POST("/permissions/grant", handler.GrantPermission)

				body := `{"role": "editor",`
				req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusBadRequest, w.Code)
			},
		},
		{
			name:     "Negative_UseCaseError",
			category: "negative",
			run: func(t *testing.T) {
				mockUseCase := new(mocks.MockIPermissionUseCase)
				handler := newTestPermissionController(mockUseCase)
				router := setupPermissionTestRouter()
				router.POST("/permissions/grant", handler.GrantPermission)

				reqBody := model.GrantPermissionRequest{Role: "editor", Path: "/articles", Method: "POST", Domain: "global"}
				mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(errors.New("use case failed")).Once()

				body := `{"role":"editor","path":"/articles","method":"POST","domain":"global"}`
				req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(body))
				req.Header.Set("Content-Type", "application/json")
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, http.StatusInternalServerError, w.Code)
				mockUseCase.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

// --- Handler Tests ---

func setupPermissionControllerTest() (*mocks.MockIPermissionUseCase, *permHandler.PermissionController) {
	mockUC := new(mocks.MockIPermissionUseCase)
	logger := logrus.New()
	validate := validator.New()
	_ = validation.RegisterCustomValidations(validate)
	controller := permHandler.NewPermissionController(mockUC, logger, validate)
	return mockUC, controller
}

func TestPermissionController_AssignRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				mockUC, controller := setupPermissionControllerTest()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				body := `{"user_id":"u1","role":"role:admin","domain":"global"}`
				c.Request, _ = http.NewRequest(http.MethodPost, "/permission/assign-role", bytes.NewBufferString(body))
				c.Request.Header.Set("Content-Type", "application/json")

				mockUC.On("AssignRoleToUser", c.Request.Context(), "u1", "role:admin", "global").Return(nil).Once()

				controller.AssignRole(c)

				assert.Equal(t, http.StatusOK, w.Code)
				mockUC.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestPermissionController_RevokeRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				mockUC, controller := setupPermissionControllerTest()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				body := `{"user_id":"u1","role":"role:admin","domain":"global"}`
				c.Request, _ = http.NewRequest(http.MethodPost, "/permission/revoke-role", bytes.NewBufferString(body))
				c.Request.Header.Set("Content-Type", "application/json")

				mockUC.On("RevokeRoleFromUser", c.Request.Context(), "u1", "role:admin", "global").Return(nil).Once()

				controller.RevokeRole(c)

				assert.Equal(t, http.StatusOK, w.Code)
				mockUC.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestPermissionController_BatchCheck(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				mockUC, controller := setupPermissionControllerTest()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				body := `{"items":[{"resource":"/api/v1/users","action":"GET"}]}`
				c.Request, _ = http.NewRequest(http.MethodPost, "/permission/batch-check", bytes.NewBufferString(body))
				c.Request.Header.Set("Content-Type", "application/json")

				var req model.BatchPermissionCheckRequest
				err := json.Unmarshal([]byte(body), &req)
				require.NoError(t, err)

				c.Set("user_id", "u1")
				mockUC.On("BatchCheckPermission", mock.Anything, "u1", req.Items).Return(map[string]bool{"/api/v1/users:GET": true}, nil).Once()

				controller.BatchCheck(c)

				assert.Equal(t, http.StatusOK, w.Code)

				var resp map[string]interface{}
				err = json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				data, ok := resp["data"].(map[string]interface{})
				require.True(t, ok)
				res, ok := data["results"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, res["/api/v1/users:GET"])

				mockUC.AssertExpectations(t)
			},
		},
		{
			name:     "Negative_Unauthorized",
			category: "negative",
			run: func(t *testing.T) {
				_, controller := setupPermissionControllerTest()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				body := `{"items":[{"resource":"/api/v1/test","action":"GET"}]}`
				c.Request, _ = http.NewRequest(http.MethodPost, "/permission/batch-check", bytes.NewBufferString(body))
				c.Request.Header.Set("Content-Type", "application/json")

				controller.BatchCheck(c)

				assert.Equal(t, http.StatusUnauthorized, w.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestPermissionController_GetAllPermissions(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_FiltersTenantDomain",
			category: "positive",
			run: func(t *testing.T) {
				mockUC, controller := setupPermissionControllerTest()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest(http.MethodGet, "/permissions", nil)
				c.Set("organization_id", "org-123")

				mockUC.On("GetAllPermissions", c.Request.Context()).Return([][]string{
					{"role:admin", "global", "/api/v1/users", "GET"},
					{"role:admin", "org-123", "/api/v1/projects", "GET"},
				}, nil).Once()

				controller.GetAllPermissions(c)

				assert.Equal(t, http.StatusOK, w.Code)
				assert.NotContains(t, w.Body.String(), "\"global\"")
				assert.Contains(t, w.Body.String(), "\"org-123\"")
				mockUC.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestPermissionController_GetUsersForRole(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_UsesResolvedTenantDomain",
			category: "positive",
			run: func(t *testing.T) {
				mockUC, controller := setupPermissionControllerTest()

				gin.SetMode(gin.TestMode)
				w := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(w)
				c.Request, _ = http.NewRequest(http.MethodGet, "/permissions/roles/role:admin/users?domain=global", nil)
				c.Params = gin.Params{{Key: "role", Value: "role:admin"}}
				c.Set("organization_id", "org-123")

				mockUC.On("GetUsersForRole", c.Request.Context(), "role:admin", "org-123").Return([]string{"u1"}, nil).Once()

				controller.GetUsersForRole(c)

				assert.Equal(t, http.StatusOK, w.Code)
				mockUC.AssertExpectations(t)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
