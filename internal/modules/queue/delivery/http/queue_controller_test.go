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
	getCalled        bool
	getID            string
	listRes          []model.QueueResponse
	getRes           *model.QueueResponse
	transitionCalled bool
	transitionReq    *model.QueueTransitionRequest
	transitionID     string
	transitionRes    *model.QueueResponse
	transitionErr    error
}

func (s *stubQueueControllerUseCase) RegisterQueue(ctx context.Context, req *model.RegisterQueueRequest) (*model.QueueResponse, error) {
	return nil, nil
}

func (s *stubQueueControllerUseCase) ListQueues(ctx context.Context) ([]model.QueueResponse, error) {
	s.listCalled = true
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
