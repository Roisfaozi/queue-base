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

func TestGrantPermission_Success(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	router := setupPermissionTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
		Domain: "global",
	}
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(nil)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
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
}

func TestGrantPermission_InvalidBody(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	router := setupPermissionTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBufferString(`{"role": "editor",`))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGrantPermission_UseCaseError(t *testing.T) {
	mockUseCase := new(mocks.MockIPermissionUseCase)
	handler := newTestPermissionController(mockUseCase)
	router := setupPermissionTestRouter()
	router.POST("/permissions/grant", handler.GrantPermission)

	reqBody := model.GrantPermissionRequest{
		Role:   "editor",
		Path:   "/articles",
		Method: "POST",
		Domain: "global",
	}
	mockError := errors.New("use case failed")
	mockUseCase.On("GrantPermissionToRole", mock.Anything, reqBody.Role, reqBody.Path, reqBody.Method, "global").Return(mockError)

	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest(http.MethodPost, "/permissions/grant", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockUseCase.AssertExpectations(t)
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
		name        string
		body        string
		setupMock   func(*mocks.MockIPermissionUseCase, *gin.Context)
		wantStatus  int
		assertBody  func(*testing.T, *httptest.ResponseRecorder)
		assertMocks bool
	}{
		{
			name: "success",
			body: `{"user_id":"u1","role":"role:admin","domain":"global"}`,
			setupMock: func(mockUC *mocks.MockIPermissionUseCase, c *gin.Context) {
				mockUC.On("AssignRoleToUser", c.Request.Context(), "u1", "role:admin", "global").Return(nil).Once()
			},
			wantStatus:  http.StatusOK,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC, controller := setupPermissionControllerTest()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/permission/assign-role", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.setupMock != nil {
				tt.setupMock(mockUC, c)
			}

			controller.AssignRole(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, w)
			}
			if tt.assertMocks {
				mockUC.AssertExpectations(t)
			}
		})
	}
}

func TestPermissionController_RevokeRole(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setupMock   func(*mocks.MockIPermissionUseCase, *gin.Context)
		wantStatus  int
		assertMocks bool
	}{
		{
			name: "success",
			body: `{"user_id":"u1","role":"role:admin","domain":"global"}`,
			setupMock: func(mockUC *mocks.MockIPermissionUseCase, c *gin.Context) {
				mockUC.On("RevokeRoleFromUser", c.Request.Context(), "u1", "role:admin", "global").Return(nil).Once()
			},
			wantStatus:  http.StatusOK,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC, controller := setupPermissionControllerTest()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/permission/revoke-role", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			if tt.setupMock != nil {
				tt.setupMock(mockUC, c)
			}

			controller.RevokeRole(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertMocks {
				mockUC.AssertExpectations(t)
			}
		})
	}
}

func TestPermissionController_BatchCheck(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		userID      string
		setupMock   func(*mocks.MockIPermissionUseCase, *gin.Context, model.BatchPermissionCheckRequest)
		wantStatus  int
		assertBody  func(*testing.T, *httptest.ResponseRecorder)
		assertMocks bool
	}{
		{
			name:   "success",
			body:   `{"items":[{"resource":"/api/v1/users","action":"GET"}]}`,
			userID: "u1",
			setupMock: func(mockUC *mocks.MockIPermissionUseCase, c *gin.Context, req model.BatchPermissionCheckRequest) {
				mockUC.On("BatchCheckPermission", mock.Anything, "u1", req.Items).Return(map[string]bool{"/api/v1/users:GET": true}, nil).Once()
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				data, ok := resp["data"].(map[string]interface{})
				require.True(t, ok)
				res, ok := data["results"].(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, true, res["/api/v1/users:GET"])
			},
			assertMocks: true,
		},
		{
			name:        "unauthorized",
			body:        `{"items":[{"resource":"/api/v1/test","action":"GET"}]}`,
			wantStatus:  http.StatusUnauthorized,
			assertMocks: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC, controller := setupPermissionControllerTest()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodPost, "/permission/batch-check", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req model.BatchPermissionCheckRequest
			err := json.Unmarshal([]byte(tt.body), &req)
			require.NoError(t, err)

			if tt.userID != "" {
				c.Set("user_id", tt.userID)
			}
			if tt.setupMock != nil {
				tt.setupMock(mockUC, c, req)
			}

			controller.BatchCheck(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, w)
			}
			if tt.assertMocks {
				mockUC.AssertExpectations(t)
			}
		})
	}
}

func TestPermissionController_GetAllPermissions(t *testing.T) {
	tests := []struct {
		name        string
		orgID       string
		setupMock   func(*mocks.MockIPermissionUseCase, *gin.Context)
		wantStatus  int
		assertBody  func(*testing.T, *httptest.ResponseRecorder)
		assertMocks bool
	}{
		{
			name:  "filters tenant domain",
			orgID: "org-123",
			setupMock: func(mockUC *mocks.MockIPermissionUseCase, c *gin.Context) {
				mockUC.On("GetAllPermissions", c.Request.Context()).Return([][]string{
					{"role:admin", "global", "/api/v1/users", "GET"},
					{"role:admin", "org-123", "/api/v1/projects", "GET"},
				}, nil).Once()
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.NotContains(t, w.Body.String(), "\"global\"")
				assert.Contains(t, w.Body.String(), "\"org-123\"")
			},
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC, controller := setupPermissionControllerTest()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/permissions", nil)
			if tt.orgID != "" {
				c.Set("organization_id", tt.orgID)
			}
			if tt.setupMock != nil {
				tt.setupMock(mockUC, c)
			}

			controller.GetAllPermissions(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, w)
			}
			if tt.assertMocks {
				mockUC.AssertExpectations(t)
			}
		})
	}
}

func TestPermissionController_GetUsersForRole(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		orgID       string
		params      gin.Params
		setupMock   func(*mocks.MockIPermissionUseCase, *gin.Context)
		wantStatus  int
		assertMocks bool
	}{
		{
			name:   "uses resolved tenant domain",
			path:   "/permissions/roles/role:admin/users?domain=global",
			orgID:  "org-123",
			params: gin.Params{{Key: "role", Value: "role:admin"}},
			setupMock: func(mockUC *mocks.MockIPermissionUseCase, c *gin.Context) {
				mockUC.On("GetUsersForRole", c.Request.Context(), "role:admin", "org-123").Return([]string{"u1"}, nil).Once()
			},
			wantStatus:  http.StatusOK,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC, controller := setupPermissionControllerTest()

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, tt.path, nil)
			c.Params = tt.params
			if tt.orgID != "" {
				c.Set("organization_id", tt.orgID)
			}
			if tt.setupMock != nil {
				tt.setupMock(mockUC, c)
			}

			controller.GetUsersForRole(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertMocks {
				mockUC.AssertExpectations(t)
			}
		})
	}
}
