package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

type stubQueueControllerUseCase struct {
	listCalled       bool
	listReq          model.ListQueuesRequest
	getCalled        bool
	getID            string
	listRes          []model.QueueResponse
	getRes           *model.QueueResponse
	transitionCalled bool
	transitionReq    *model.QueueTransitionRequest
	transitionID     string
	transitionRes    *model.QueueResponse
	transitionErr    error
	journeyReq       model.QueueJourneyListRequest
	journeyRes       []model.QueueJourneyResponse
}

func (s *stubQueueControllerUseCase) RegisterQueue(ctx context.Context, req *model.RegisterQueueRequest) (*model.QueueResponse, error) {
	return nil, nil
}

func (s *stubQueueControllerUseCase) ListQueues(ctx context.Context, req model.ListQueuesRequest) ([]model.QueueResponse, error) {
	s.listCalled = true
	s.listReq = req
	return s.listRes, nil
}

func (s *stubQueueControllerUseCase) GetQueueByID(ctx context.Context, queueID string) (*model.QueueResponse, error) {
	s.getCalled = true
	s.getID = queueID
	return s.getRes, nil
}

func (s *stubQueueControllerUseCase) ForwardQueue(ctx context.Context, queueID string, req *model.ForwardQueueRequest) (*model.QueueResponse, error) {
	return nil, nil
}

func (s *stubQueueControllerUseCase) TransitionQueue(ctx context.Context, queueID string, req *model.QueueTransitionRequest) (*model.QueueResponse, error) {
	s.transitionCalled = true
	s.transitionID = queueID
	s.transitionReq = req
	return s.transitionRes, s.transitionErr
}

func (s *stubQueueControllerUseCase) ListActiveJourneys(ctx context.Context, req model.QueueJourneyListRequest) ([]model.QueueJourneyResponse, error) {
	s.journeyReq = req
	return s.journeyRes, nil
}

func TestQueueController_Transition(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{transitionRes: &model.QueueResponse{ID: "q-1", Status: entity.QueueStatusCalling}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.POST("/queues/:id/transition", controller.Transition)

	body, _ := json.Marshal(model.QueueTransitionRequest{Action: model.QueueActionCall})
	req, _ := http.NewRequest("POST", "/queues/q-1/transition", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.transitionCalled)
	assert.Equal(t, "q-1", uc.transitionID)
	assert.Equal(t, model.QueueActionCall, uc.transitionReq.Action)
}

func TestQueueController_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{getRes: &model.QueueResponse{ID: "q-1"}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queues/:id", controller.GetByID)

	req, _ := http.NewRequest("GET", "/queues/q-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.getCalled)
	assert.Equal(t, "q-1", uc.getID)
}

func TestQueueController_GetAll(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{listRes: []model.QueueResponse{{ID: "q-1", Status: entity.QueueStatusWaiting}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queues", controller.GetAll)

	req, _ := http.NewRequest("GET", "/queues?status=waiting", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.listCalled)
	assert.Equal(t, model.ListQueuesRequest{Status: "waiting"}, uc.listReq)
}

func TestQueueController_GetAll_WithFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{listRes: []model.QueueResponse{{ID: "q-1", QueueDate: "2026-06-24"}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queues", controller.GetAll)

	req, _ := http.NewRequest("GET", "/queues?status=waiting&queue_date=2026-06-24&service_id=s-1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.listCalled)
	assert.Equal(t, model.ListQueuesRequest{Status: "waiting", QueueDate: "2026-06-24", ServiceID: "s-1"}, uc.listReq)
}

func TestQueueController_GetJourneysByService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{journeyRes: []model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1", Status: entity.JourneyStatusCalling}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queues/services/:service_id/queue-journeys", controller.GetJourneysByService)

	req, _ := http.NewRequest("GET", "/queues/services/svc-1/queue-journeys?queue_date=2026-06-24", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"}, uc.journeyReq)
}

func TestQueueController_GetJourneysByCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{journeyRes: []model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1", Status: entity.JourneyStatusCalling}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queues/counters/:counter_id/queue-journeys", controller.GetJourneysByCounter)

	req, _ := http.NewRequest("GET", "/queues/counters/c-1/queue-journeys?status=calling", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, model.QueueJourneyListRequest{CounterID: "c-1", Status: "calling"}, uc.journeyReq)
}
