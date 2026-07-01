package test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	accessHandler "github.com/Roisfaozi/queue-base/internal/modules/access/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/access/model"
	"github.com/Roisfaozi/queue-base/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupAccessTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func newTestAccessController(mockUseCase *mocks.MockIAccessUseCase) *accessHandler.AccessController {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	v := validator.New()
	_ = validation.RegisterCustomValidations(v)
	return accessHandler.NewAccessController(mockUseCase, v, log)
}

func TestAccessHandler_CreateAccessRight(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setupMock   func(*mocks.MockIAccessUseCase)
		wantStatus  int
		assertMocks bool
	}{
		{
			name: "success",
			body: `{"name":"test_access_right","description":"A description"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.CreateAccessRightRequest{Name: "test_access_right", Description: "A description"}
				resBody := &model.AccessRightResponse{ID: "1", Name: "test_access_right"}
				mockUseCase.On("CreateAccessRight", mock.Anything, reqBody).Return(resBody, nil).Once()
			},
			wantStatus:  http.StatusCreated,
			assertMocks: true,
		},
		{
			name:       "invalid body",
			body:       `{"name":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation errors",
			body:       `{"name":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "usecase error",
			body: `{"name":"test_access_right"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.CreateAccessRightRequest{Name: "test_access_right"}
				mockUseCase.On("CreateAccessRight", mock.Anything, reqBody).Return(nil, errors.New("db error")).Once()
			},
			wantStatus:  http.StatusInternalServerError,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.POST("/access-rights", handler.CreateAccessRight)

			if tt.setupMock != nil {
				tt.setupMock(mockUseCase)
			}

			req, _ := http.NewRequest(http.MethodPost, "/access-rights", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertMocks {
				mockUseCase.AssertExpectations(t)
			} else {
				mockUseCase.AssertNotCalled(t, "CreateAccessRight", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestAccessHandler_GetAllAccessRights(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mocks.MockIAccessUseCase)
		wantStatus int
	}{
		{
			name: "success",
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				resBody := &model.AccessRightListResponse{Data: []model.AccessRightResponse{{ID: "1", Name: "right1"}, {ID: "2", Name: "right2"}}}
				mockUseCase.On("GetAllAccessRights", mock.Anything).Return(resBody, nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "usecase error",
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				mockUseCase.On("GetAllAccessRights", mock.Anything).Return(nil, errors.New("db error")).Once()
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.GET("/access-rights", handler.GetAllAccessRights)
			tt.setupMock(mockUseCase)

			req, _ := http.NewRequest(http.MethodGet, "/access-rights", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestAccessHandler_CreateEndpoint(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setupMock   func(*mocks.MockIAccessUseCase)
		wantStatus  int
		assertMocks bool
	}{
		{
			name: "success",
			body: `{"path":"/test/path","method":"GET"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.CreateEndpointRequest{Path: "/test/path", Method: "GET"}
				resBody := &model.EndpointResponse{ID: "1", Path: "/test/path", Method: "GET"}
				mockUseCase.On("CreateEndpoint", mock.Anything, reqBody).Return(resBody, nil).Once()
			},
			wantStatus:  http.StatusCreated,
			assertMocks: true,
		},
		{
			name:       "invalid body",
			body:       `{"path":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "validation errors",
			body:       `{"path":""}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "usecase error",
			body: `{"path":"/test/path","method":"GET"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.CreateEndpointRequest{Path: "/test/path", Method: "GET"}
				mockUseCase.On("CreateEndpoint", mock.Anything, reqBody).Return(nil, errors.New("db error")).Once()
			},
			wantStatus:  http.StatusInternalServerError,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.POST("/endpoints", handler.CreateEndpoint)

			if tt.setupMock != nil {
				tt.setupMock(mockUseCase)
			}

			req, _ := http.NewRequest(http.MethodPost, "/endpoints", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertMocks {
				mockUseCase.AssertExpectations(t)
			} else {
				mockUseCase.AssertNotCalled(t, "CreateEndpoint", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestAccessHandler_LinkEndpointToAccessRight(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setupMock   func(*mocks.MockIAccessUseCase)
		wantStatus  int
		assertMocks bool
	}{
		{
			name: "success",
			body: `{"access_right_id":"1","endpoint_id":"1"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "1"}
				mockUseCase.On("LinkEndpointToAccessRight", mock.Anything, reqBody).Return(nil).Once()
			},
			wantStatus:  http.StatusOK,
			assertMocks: true,
		},
		{name: "invalid body", body: `{"access_right_id":`, wantStatus: http.StatusBadRequest},
		{name: "validation errors", body: `{"access_right_id":"","endpoint_id":""}`, wantStatus: http.StatusUnprocessableEntity},
		{
			name: "usecase error",
			body: `{"access_right_id":"1","endpoint_id":"1"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "1"}
				mockUseCase.On("LinkEndpointToAccessRight", mock.Anything, reqBody).Return(errors.New("db error")).Once()
			},
			wantStatus:  http.StatusInternalServerError,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.POST("/access-rights/link", handler.LinkEndpointToAccessRight)

			if tt.setupMock != nil {
				tt.setupMock(mockUseCase)
			}

			req, _ := http.NewRequest(http.MethodPost, "/access-rights/link", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertMocks {
				mockUseCase.AssertExpectations(t)
			} else {
				mockUseCase.AssertNotCalled(t, "LinkEndpointToAccessRight", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestAccessHandler_DeleteAccessRight(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func(*mocks.MockIAccessUseCase)
		wantStatus int
	}{
		{
			name: "success",
			path: "/access-rights/1",
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				mockUseCase.On("DeleteAccessRight", mock.Anything, "1").Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/access-rights/1",
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				mockUseCase.On("DeleteAccessRight", mock.Anything, "1").Return(exception.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.DELETE("/access-rights/:id", handler.DeleteAccessRight)
			tt.setupMock(mockUseCase)

			req, _ := http.NewRequest(http.MethodDelete, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestAccessHandler_DeleteEndpoint(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		setupMock  func(*mocks.MockIAccessUseCase)
		wantStatus int
	}{
		{
			name: "success",
			path: "/endpoints/1",
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				mockUseCase.On("DeleteEndpoint", mock.Anything, "1").Return(nil).Once()
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found",
			path: "/endpoints/1",
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				mockUseCase.On("DeleteEndpoint", mock.Anything, "1").Return(exception.ErrNotFound).Once()
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.DELETE("/endpoints/:id", handler.DeleteEndpoint)
			tt.setupMock(mockUseCase)

			req, _ := http.NewRequest(http.MethodDelete, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestAccessHandler_GetEndpointsDynamic(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mocks.MockIAccessUseCase, *querybuilder.DynamicFilter)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"filter":{"Method":{"type":"equals","from":"GET"}}}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase, filter *querybuilder.DynamicFilter) {
				expectedEndpoints := []*model.EndpointResponse{{ID: "1", Path: "/test", Method: "GET"}}
				mockUseCase.On("GetEndpointsDynamic", mock.Anything, filter).Return(expectedEndpoints, int64(1), nil).Once()
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.POST("/endpoints/search", handler.GetEndpointsDynamic)

			var filter querybuilder.DynamicFilter
			_ = json.Unmarshal([]byte(tt.body), &filter)
			tt.setupMock(mockUseCase, &filter)

			req, _ := http.NewRequest(http.MethodPost, "/endpoints/search", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestAccessHandler_GetAccessRightsDynamic(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		setupMock  func(*mocks.MockIAccessUseCase, *querybuilder.DynamicFilter)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"filter":{"Name":{"type":"contains","from":"Manage"}}}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase, filter *querybuilder.DynamicFilter) {
				expectedResponse := &model.AccessRightListResponse{Data: []model.AccessRightResponse{{ID: "1", Name: "Manage Users"}}}
				mockUseCase.On("GetAccessRightsDynamic", mock.Anything, filter).Return(expectedResponse, int64(1), nil).Once()
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.POST("/access-rights/search", handler.GetAccessRightsDynamic)

			var filter querybuilder.DynamicFilter
			_ = json.Unmarshal([]byte(tt.body), &filter)
			tt.setupMock(mockUseCase, &filter)

			req, _ := http.NewRequest(http.MethodPost, "/access-rights/search", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			mockUseCase.AssertExpectations(t)
		})
	}
}

func TestAccessHandler_CreateAccessRight_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.CreateAccessRightRequest
		expectedCode int
	}{
		{
			name: "XSS in Name",
			payload: model.CreateAccessRightRequest{
				Name:        "<script>alert(1)</script>",
				Description: "Valid Description",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Description",
			payload: model.CreateAccessRightRequest{
				Name:        "Valid Name",
				Description: "<img src=x onerror=alert(1)>",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.CreateAccessRightRequest{
				Name:        "Safe Name",
				Description: "Safe Description",
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccessUseCase := mocks.NewMockIAccessUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := accessHandler.NewAccessController(mockAccessUseCase, v, logger)

			if tt.expectedCode == http.StatusCreated {
				mockAccessUseCase.On("CreateAccessRight", mock.Anything, mock.Anything).Return(nil, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("POST", "/access-rights", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.CreateAccessRight(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestAccessHandler_CreateEndpoint_XSS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		payload      model.CreateEndpointRequest
		expectedCode int
	}{
		{
			name: "XSS in Path",
			payload: model.CreateEndpointRequest{
				Path:   "/api/<script>alert(1)</script>",
				Method: "GET",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "XSS in Method",
			payload: model.CreateEndpointRequest{
				Path:   "/api/valid",
				Method: "<script>",
			},
			expectedCode: http.StatusUnprocessableEntity,
		},
		{
			name: "Safe content",
			payload: model.CreateEndpointRequest{
				Path:   "/api/valid",
				Method: "POST",
			},
			expectedCode: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAccessUseCase := mocks.NewMockIAccessUseCase(t)

			v := validator.New()
			_ = validation.RegisterCustomValidations(v)
			logger := logrus.New()

			controller := accessHandler.NewAccessController(mockAccessUseCase, v, logger)

			if tt.expectedCode == http.StatusCreated {
				mockAccessUseCase.On("CreateEndpoint", mock.Anything, mock.Anything).Return(nil, nil)
			}

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest("POST", "/endpoints", bytes.NewBuffer(jsonBytes))
			c.Request.Header.Set("Content-Type", "application/json")

			controller.CreateEndpoint(c)

			assert.Equal(t, tt.expectedCode, w.Code)
		})
	}
}

func TestAccessHandler_UnlinkEndpointFromAccessRight(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		setupMock   func(*mocks.MockIAccessUseCase)
		wantStatus  int
		assertMocks bool
	}{
		{
			name: "success",
			body: `{"access_right_id":"1","endpoint_id":"1"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "1"}
				mockUseCase.On("UnlinkEndpointFromAccessRight", mock.Anything, reqBody).Return(nil).Once()
			},
			wantStatus:  http.StatusOK,
			assertMocks: true,
		},
		{name: "invalid body", body: `{"access_right_id":`, wantStatus: http.StatusBadRequest},
		{name: "validation errors", body: `{"access_right_id":"","endpoint_id":""}`, wantStatus: http.StatusUnprocessableEntity},
		{
			name: "usecase error",
			body: `{"access_right_id":"1","endpoint_id":"1"}`,
			setupMock: func(mockUseCase *mocks.MockIAccessUseCase) {
				reqBody := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "1"}
				mockUseCase.On("UnlinkEndpointFromAccessRight", mock.Anything, reqBody).Return(errors.New("db error")).Once()
			},
			wantStatus:  http.StatusInternalServerError,
			assertMocks: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUseCase := new(mocks.MockIAccessUseCase)
			handler := newTestAccessController(mockUseCase)
			router := setupAccessTestRouter()
			router.POST("/access-rights/unlink", handler.UnlinkEndpointFromAccessRight)

			if tt.setupMock != nil {
				tt.setupMock(mockUseCase)
			}

			req, _ := http.NewRequest(http.MethodPost, "/access-rights/unlink", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.assertMocks {
				mockUseCase.AssertExpectations(t)
			} else {
				mockUseCase.AssertNotCalled(t, "UnlinkEndpointFromAccessRight", mock.Anything, mock.Anything)
			}
		})
	}
}
