package http_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	roleHttp "github.com/Roisfaozi/queue-base/internal/modules/role/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/role/model"
	"github.com/Roisfaozi/queue-base/internal/modules/role/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/role/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/pkg/response"
	"github.com/Roisfaozi/queue-base/pkg/validation" // Import validation pkg
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

func setupRoleTestRouter(uc usecase.RoleUseCase) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	v := validator.New()
	_ = validation.RegisterCustomValidations(v)

	handler := roleHttp.NewRoleController(uc, logrus.New(), v)
	apiV1 := router.Group("/api/v1")
	{
		apiV1.POST("/roles", handler.Create)
		apiV1.GET("/roles", handler.GetAll)
		apiV1.PUT("/roles/:id", handler.Update)
		apiV1.DELETE("/roles/:id", handler.Delete)
		apiV1.POST("/roles/search", handler.GetRolesDynamic)
	}
	return router
}

func TestRoleHandler_Create(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		body         string
		setupMock    func(*mocks.MockRoleUseCase)
		wantCode     int
		wantContains string
		assertNoCall bool
	}{
		{
			name:     "Success",
			category: "positive",
			body:     `{"name":"admin","description":"Administrator role"}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase) {
				req := model.CreateRoleRequest{Name: "admin", Description: "Administrator role"}
				mockUseCase.On("Create", mock.Anything, &req).Return(&model.RoleResponse{ID: "uuid", Name: "admin"}, nil)
			},
			wantCode: http.StatusCreated,
		},
		{name: "BindingError", category: "negative", body: `invalid json`, wantCode: http.StatusBadRequest, wantContains: "invalid request body", assertNoCall: true},
		{name: "ValidationError", category: "negative", body: `{"name":"","description":"Administrator role"}`, wantCode: http.StatusUnprocessableEntity, wantContains: "validation error", assertNoCall: true},
		{
			name:     "UseCaseError",
			category: "negative",
			body:     `{"name":"existing","description":"Existing role"}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase) {
				req := model.CreateRoleRequest{Name: "existing", Description: "Existing role"}
				mockUseCase.On("Create", mock.Anything, &req).Return(nil, exception.ErrConflict)
			},
			wantCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockRoleUseCase)
			router := setupRoleTestRouter(mockUseCase)
			if tt.setupMock != nil {
				tt.setupMock(mockUseCase)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantContains)
			}
			if tt.assertNoCall {
				mockUseCase.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			} else {
				mockUseCase.AssertExpectations(t)
			}
		})
	}
}

func TestRoleHandler_GetAll(t *testing.T) {
	tests := []struct {
		name         string
		setupMock    func(*mocks.MockRoleUseCase)
		wantCode     int
		wantDataLen  int
		wantContains string
	}{
		{
			name: "Success",
			setupMock: func(mockUseCase *mocks.MockRoleUseCase) {
				expectedRoles := []model.RoleResponse{{ID: "1", Name: "admin"}, {ID: "2", Name: "user"}}
				mockUseCase.On("GetAll", mock.Anything).Return(expectedRoles, nil).Once()
			},
			wantCode:    http.StatusOK,
			wantDataLen: 2,
		},
		{
			name: "UseCaseError",
			setupMock: func(mockUseCase *mocks.MockRoleUseCase) {
				mockUseCase.On("GetAll", mock.Anything).Return(nil, errors.New("some database error")).Once()
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockRoleUseCase)
			router := setupRoleTestRouter(mockUseCase)
			tt.setupMock(mockUseCase)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/api/v1/roles", nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantDataLen > 0 {
				var responseBody response.WebResponseSuccess[[]model.RoleResponse]
				err := json.Unmarshal(w.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Len(t, responseBody.Data, tt.wantDataLen)
			}
			if tt.wantContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantContains)
			}
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestRoleHandler_Delete(t *testing.T) {
	tests := []struct {
		name     string
		category string
		roleID   string
		err      error
		wantCode int
	}{
		{name: "Success", category: "positive", roleID: "test-uuid", wantCode: http.StatusOK},
		{name: "NotFound", category: "negative", roleID: "non-existent-uuid", err: exception.ErrNotFound, wantCode: http.StatusNotFound},
		{name: "Forbidden", category: "vulnerability", roleID: "superadmin-uuid", err: exception.ErrForbidden, wantCode: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockRoleUseCase)
			router := setupRoleTestRouter(mockUseCase)
			mockUseCase.On("Delete", mock.Anything, tt.roleID).Return(tt.err)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+tt.roleID, nil)
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestRoleHandler_GetRolesDynamic(t *testing.T) {
	tests := []struct {
		name         string
		body         string
		setupMock    func(*mocks.MockRoleUseCase, *querybuilder.DynamicFilter)
		wantCode     int
		wantDataLen  int
		wantContains string
		assertNoCall bool
	}{
		{
			name: "Success",
			body: `{"filter":{"Name":{"type":"contains","from":"test"}}}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase, dynamicFilter *querybuilder.DynamicFilter) {
				expectedRoles := []model.RoleResponse{{ID: "1", Name: "test_role"}}
				mockUseCase.On("GetAllRolesDynamic", mock.Anything, dynamicFilter).Return(expectedRoles, nil).Once()
			},
			wantCode:    http.StatusOK,
			wantDataLen: 1,
		},
		{name: "BindingError", body: `invalid json`, wantCode: http.StatusBadRequest, wantContains: "invalid request body", assertNoCall: true},
		{
			name: "UseCaseError",
			body: `{"filter":{}}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase, dynamicFilter *querybuilder.DynamicFilter) {
				mockUseCase.On("GetAllRolesDynamic", mock.Anything, dynamicFilter).Return(nil, errors.New("db error")).Once()
			},
			wantCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockRoleUseCase)
			router := setupRoleTestRouter(mockUseCase)

			var dynamicFilter querybuilder.DynamicFilter
			if tt.body != "invalid json" {
				err := json.Unmarshal([]byte(tt.body), &dynamicFilter)
				assert.NoError(t, err)
			}
			if tt.setupMock != nil {
				tt.setupMock(mockUseCase, &dynamicFilter)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles/search", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantDataLen > 0 {
				var responseBody response.WebResponseSuccess[[]model.RoleResponse]
				err := json.Unmarshal(w.Body.Bytes(), &responseBody)
				assert.NoError(t, err)
				assert.Len(t, responseBody.Data, tt.wantDataLen)
			}
			if tt.wantContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantContains)
			}
			if tt.assertNoCall {
				mockUseCase.AssertNotCalled(t, "GetAllRolesDynamic", mock.Anything, mock.Anything)
			} else {
				mockUseCase.AssertExpectations(t)
			}
		})
	}
}

func TestRoleHandler_Create_XSS_Name(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	createRequest := model.CreateRoleRequest{Name: "<script>alert(1)</script>", Description: "XSS role"}
	requestBody, _ := json.Marshal(createRequest)

	// Expect sanitized input
	sanitizedRequest := model.CreateRoleRequest{Name: "&lt;script&gt;alert(1)&lt;/script&gt;", Description: "XSS role"}
	mockUseCase.On("Create", mock.Anything, &sanitizedRequest).Return(&model.RoleResponse{ID: "uuid", Name: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/roles", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Should return 201 Created due to sanitization
	assert.Equal(t, http.StatusCreated, w.Code)
	mockUseCase.AssertExpectations(t)
}

func TestRoleHandler_Update(t *testing.T) {
	tests := []struct {
		name         string
		roleID       string
		body         string
		setupMock    func(*mocks.MockRoleUseCase, string)
		wantCode     int
		wantContains string
		assertNoCall bool
	}{
		{
			name:   "Success",
			roleID: "test-uuid",
			body:   `{"description":"Updated description"}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase, roleID string) {
				updateRequest := model.UpdateRoleRequest{Description: "Updated description"}
				mockUseCase.On("Update", mock.Anything, roleID, &updateRequest).Return(&model.RoleResponse{ID: roleID, Name: "admin", Description: "Updated description"}, nil).Once()
			},
			wantCode: http.StatusOK,
		},
		{
			name:         "BindingError",
			roleID:       "test-uuid",
			body:         `invalid json`,
			wantCode:     http.StatusBadRequest,
			wantContains: "invalid request body",
			assertNoCall: true,
		},
		{
			name:   "XSS Sanitization",
			roleID: "test-uuid",
			body:   `{"description":"<script>alert(1)</script>"}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase, roleID string) {
				sanitizedRequest := model.UpdateRoleRequest{Description: "&lt;script&gt;alert(1)&lt;/script&gt;"}
				mockUseCase.On("Update", mock.Anything, roleID, &sanitizedRequest).Return(&model.RoleResponse{ID: roleID, Name: "admin", Description: "&lt;script&gt;alert(1)&lt;/script&gt;"}, nil).Once()
			},
			wantCode: http.StatusOK,
		},
		{
			name:   "UseCaseError",
			roleID: "test-uuid",
			body:   `{"description":"Updated description"}`,
			setupMock: func(mockUseCase *mocks.MockRoleUseCase, roleID string) {
				updateRequest := model.UpdateRoleRequest{Description: "Updated description"}
				mockUseCase.On("Update", mock.Anything, roleID, &updateRequest).Return(nil, exception.ErrNotFound).Once()
			},
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockRoleUseCase)
			router := setupRoleTestRouter(mockUseCase)
			if tt.setupMock != nil {
				tt.setupMock(mockUseCase, tt.roleID)
			}

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/api/v1/roles/"+tt.roleID, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)
			if tt.wantContains != "" {
				assert.Contains(t, w.Body.String(), tt.wantContains)
			}
			if tt.assertNoCall {
				mockUseCase.AssertNotCalled(t, "Update", mock.Anything, mock.Anything, mock.Anything)
			} else {
				mockUseCase.AssertExpectations(t)
			}
		})
	}
}

func TestRoleHandler_HandleError_Variants(t *testing.T) {
	mockUseCase := new(mocks.MockRoleUseCase)
	router := setupRoleTestRouter(mockUseCase)

	roleID := "test-uuid"

	tests := []struct {
		err                  error
		expectedCode         int
		expectedBodyContains string
	}{
		{exception.ErrBadRequest, http.StatusBadRequest, "failed to delete role"},
		{exception.ErrUnauthorized, http.StatusUnauthorized, "failed to delete role"},
		{exception.ErrForbidden, http.StatusForbidden, "failed to delete role"},
		{exception.ErrNotFound, http.StatusNotFound, "failed to delete role"},
		{exception.ErrConflict, http.StatusConflict, "failed to delete role"},
		{errors.New("unknown error"), http.StatusInternalServerError, "something went wrong"},
	}

	for _, tt := range tests {
		mockUseCase.ExpectedCalls = nil // Clear expected calls
		mockUseCase.On("Delete", mock.Anything, roleID).Return(tt.err).Once()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodDelete, "/api/v1/roles/"+roleID, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, tt.expectedCode, w.Code, "Expected code %d for error %v", tt.expectedCode, tt.err)
		assert.Contains(t, w.Body.String(), tt.expectedBodyContains)
	}
}
