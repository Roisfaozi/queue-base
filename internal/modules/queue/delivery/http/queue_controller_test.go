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
	visitRes         []model.VisitJourneyResponse
	statsRes         *model.QueueStatsResponse
	statsCalled      bool
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

func (s *stubQueueControllerUseCase) GetVisitJourneys(ctx context.Context, queueID string) ([]model.VisitJourneyResponse, error) {
	s.getCalled = true
	s.getID = queueID
	return s.visitRes, nil
}

func (s *stubQueueControllerUseCase) GetQueueStats(ctx context.Context) (*model.QueueStatsResponse, error) {
	s.statsCalled = true
	return s.statsRes, nil
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

func TestQueueController_GetQueueStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{statsRes: &model.QueueStatsResponse{TotalQueuesToday: 10}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/branches/:branch_id/queue-stats", controller.GetQueueStats)

	req, _ := http.NewRequest("GET", "/branches/branch-1/queue-stats", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.statsCalled)
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

func TestQueueController_GetJourneysByBranchAndService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{journeyRes: []model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1"}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/branches/:branch_id/services/:service_id/queue-journeys", controller.GetJourneysByBranchAndService)

	req, _ := http.NewRequest("GET", "/branches/branch-1/services/svc-1/queue-journeys?queue_date=2026-06-24", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"}, uc.journeyReq)
}

func TestQueueController_GetJourneysByBranchAndCounter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{journeyRes: []model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1"}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/branches/:branch_id/counters/:counter_id/queue-journeys", controller.GetJourneysByBranchAndCounter)

	req, _ := http.NewRequest("GET", "/branches/branch-1/counters/c-1/queue-journeys?status=calling", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, model.QueueJourneyListRequest{CounterID: "c-1", Status: "calling"}, uc.journeyReq)
}

func TestQueueController_GetVisitJourneys(t *testing.T) {
	gin.SetMode(gin.TestMode)
	uc := &stubQueueControllerUseCase{visitRes: []model.VisitJourneyResponse{{ID: "v-1", EventType: "registration"}}}
	controller := NewQueueController(uc, validator.New())
	router := gin.New()
	router.Use(func(c *gin.Context) {
		ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	router.GET("/queues/:id/visit-journeys", controller.GetVisitJourneys)

	req, _ := http.NewRequest("GET", "/queues/q-1/visit-journeys", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, uc.getCalled)
	assert.Equal(t, "q-1", uc.getID)
}
