package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/scanner/model"
	"github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type stubScannerControllerUseCase struct {
	called bool
	last   *usecase.CheckInRequest
	res    *usecase.CheckInResponse
	err    error
}

func (s *stubScannerControllerUseCase) CheckIn(ctx context.Context, req *usecase.CheckInRequest) (*usecase.CheckInResponse, error) {
	s.called = true
	s.last = req
	return s.res, s.err
}

func TestScannerController_CheckIn_UsesHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{res: &usecase.CheckInResponse{Action: usecase.ActionRegister}}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/scanner/check-in", controller.CheckIn)

	body, _ := json.Marshal(model.CheckInRequest{Action: model.CheckInRequest{Action: "register"}.Action, ServiceID: "service-1", PatientName: "John Doe"})
	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
	req.Header.Set("X-Client-ID", "client-1")
	req.Header.Set("X-API-Key", "key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.called)
	assert.Equal(t, "client-1", uc.last.ClientID)
	assert.Equal(t, "key-1", uc.last.APIKey)
}

func TestScannerController_CheckIn_RejectsMissingHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/scanner/check-in", controller.CheckIn)

	body, _ := json.Marshal(model.CheckInRequest{Action: "register", ServiceID: "service-1", PatientName: "John Doe"})
	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, uc.called)
}

func TestScannerController_CheckIn_ForwardsPharmacyPayload(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{res: &usecase.CheckInResponse{Action: usecase.ActionForward}}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/scanner/check-in", controller.CheckIn)

	body, _ := json.Marshal(model.CheckInRequest{Action: "forward", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"})
	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
	req.Header.Set("X-Client-ID", "client-1")
	req.Header.Set("X-API-Key", "key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.called)
	assert.Equal(t, "forward", uc.last.Action)
	assert.Equal(t, "pharmacy-svc", uc.last.DestinationServiceID)
	assert.Equal(t, "counter-1", uc.last.DestinationCounterID)
}

func TestScannerController_CheckIn_PropagatesWorkflowForbidden(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{err: exception.ErrForbidden}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/scanner/check-in", controller.CheckIn)

	body, _ := json.Marshal(model.CheckInRequest{Action: "forward", QueueID: "q-1", DestinationServiceID: "pharmacy-svc", DestinationCounterID: "counter-1"})
	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
	req.Header.Set("X-Client-ID", "client-1")
	req.Header.Set("X-API-Key", "key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.True(t, uc.called)
}

func TestScannerController_CheckIn_RejectsInvalidAction(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.POST("/scanner/check-in", controller.CheckIn)

	body := []byte(`{"action":"drop-table","service_id":"s-1","patient_name":"John"}`)
	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
	req.Header.Set("X-Client-ID", "client-1")
	req.Header.Set("X-API-Key", "key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.False(t, uc.called)
}

func TestScannerController_CheckIn_PropagatesUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{err: exception.ErrUnauthorized}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/scanner/check-in", controller.CheckIn)

	body, _ := json.Marshal(model.CheckInRequest{Action: "register", ServiceID: "s-1", PatientName: "John"})
	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer(body))
	req.Header.Set("X-Client-ID", "client-1")
	req.Header.Set("X-API-Key", "bad-key")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.True(t, uc.called)
}

func TestScannerController_CheckIn_RejectsEmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubScannerControllerUseCase{}
	controller := NewScannerController(uc, validator.New())
	router := gin.New()
	router.POST("/scanner/check-in", controller.CheckIn)

	req, _ := http.NewRequest("POST", "/scanner/check-in", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("X-Client-ID", "client-1")
	req.Header.Set("X-API-Key", "key-1")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.False(t, uc.called)
}
