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

	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func newQueueTestValidator() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	if err := validation.RegisterCustomValidations(v); err != nil {
		panic(err)
	}
	return v
}

type stubQueueControllerUseCase struct {
	registerCalled   bool
	registerReq      *model.RegisterQueueRequest
	registerBranchID string
	listCalled       bool
	listReq          model.ListQueuesRequest
	listBranchID     string
	getCalled        bool
	getID            string
	forwardCalled    bool
	forwardReq       *model.ForwardQueueRequest
	forwardID        string
	forwardRes       *model.QueueResponse
	forwardErr       error
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
	s.registerCalled = true
	s.registerReq = req
	s.registerBranchID = database.GetBranchID(ctx)
	return &model.QueueResponse{ID: "q-1", BranchID: s.registerBranchID}, nil
}
func (s *stubQueueControllerUseCase) ListQueues(ctx context.Context, req model.ListQueuesRequest) ([]model.QueueResponse, error) {
	s.listCalled = true
	s.listReq = req
	s.listBranchID = database.GetBranchID(ctx)
	return s.listRes, nil
}
func (s *stubQueueControllerUseCase) GetQueueByID(ctx context.Context, queueID string) (*model.QueueResponse, error) {
	s.getCalled = true
	s.getID = queueID
	return s.getRes, nil
}
func (s *stubQueueControllerUseCase) ForwardQueue(ctx context.Context, queueID string, req *model.ForwardQueueRequest) (*model.QueueResponse, error) {
	s.forwardCalled = true
	s.forwardID = queueID
	s.forwardReq = req
	return s.forwardRes, s.forwardErr
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

func TestQueueController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Register", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			reqBody  interface{}
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:    "Positive_Register",
				category: "positive",
				reqBody: model.RegisterQueueRequest{
					BranchID: "550e8400-e29b-41d4-a716-446655440000",
					ServiceID: "550e8400-e29b-41d4-a716-446655440001",
					PatientName: "John Queue",
				},
				setup:    func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode: http.StatusCreated,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.registerCalled)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uc.registerBranchID)
					if assert.NotNil(t, uc.registerReq) {
						assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uc.registerReq.BranchID)
					}
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:= logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.POST("/queues", controller.Register)

				body, _ := json.Marshal(tt.reqBody)
				req, _ := http.NewRequest("POST", "/queues", bytes.NewBuffer(body))
				req.Header.Set("Content-Type", "application/json")
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
			category string
			query    string
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:     "Positive_SetsBranchContextFromQuery",
				category: "positive",
				query:    "?branch_id=550e8400-e29b-41d4-a716-446655440000&status=waiting",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{listRes: []model.QueueResponse{{ID: "q-1"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.listCalled)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uc.listBranchID)
					assert.Equal(t, model.ListQueuesRequest{
						BranchID: "550e8400-e29b-41d4-a716-446655440000",
						Status:   "waiting",
					}, uc.listReq)
				},
			},
			{
				name:     "Positive_BranchContextFromContext",
				category: "positive",
				query:    "?status=waiting",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{listRes: []model.QueueResponse{{ID: "q-1", Status: entity.QueueStatusWaiting}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.listCalled)
					assert.Equal(t, model.ListQueuesRequest{Status: "waiting"}, uc.listReq)
				},
			},
			{
				name:     "Positive_WithFilters",
				category: "positive",
				query:    "?status=waiting&queue_date=2026-06-24&service_id=s-1",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{listRes: []model.QueueResponse{{ID: "q-1", QueueDate: "2026-06-24"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.listCalled)
					assert.Equal(t, model.ListQueuesRequest{
						Status:    "waiting",
						QueueDate: "2026-06-24",
						ServiceID: "s-1",
					}, uc.listReq)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					ctx = database.SetBranchContext(ctx, "b-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/queues", controller.GetAll)

				req, _ := http.NewRequest("GET", "/queues"+tt.query, nil)
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
			category string
			queueID  string
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:     "Positive_GetByID",
				category: "positive",
				queueID:  "q-1",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{getRes: &model.QueueResponse{ID: "q-1"}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.getCalled)
					assert.Equal(t, "q-1", uc.getID)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/queues/:id", controller.GetByID)

				req, _ := http.NewRequest("GET", "/queues/"+tt.queueID, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("Transition", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			reqBody  interface{}
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:     "Positive_Transition",
				category: "positive",
				queueID:  "q-1",
				reqBody:  model.QueueTransitionRequest{Action: model.QueueActionCall},
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{transitionRes: &model.QueueResponse{ID: "q-1", Status: entity.QueueStatusCalling}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.transitionCalled)
					assert.Equal(t, "q-1", uc.transitionID)
					assert.Equal(t, model.QueueActionCall, uc.transitionReq.Action)
				},
			},
			{
				name:     "Negative_TransitionRejectsMissingID",
				category: "negative",
				queueID:  "",
				reqBody:  model.QueueTransitionRequest{Action: model.QueueActionCall},
				setup:    func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode: http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.False(t, uc.transitionCalled)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.POST("/queues/:id/transition", controller.Transition)

				body, _ := json.Marshal(tt.reqBody)
				path := "/queues/" + tt.queueID + "/transition"
				req, _ := http.NewRequest("POST", path, bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("Forward", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			reqBody  interface{}
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:     "Positive_Forward",
				category: "positive",
				queueID:  "q-1",
				reqBody: model.ForwardQueueRequest{
					DestinationServiceID: "550e8400-e29b-41d4-a716-446655440000",
					DestinationCounterID: "550e8400-e29b-41d4-a716-446655440001",
				},
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{forwardRes: &model.QueueResponse{ID: "q-1", CurrentJourneyID: "j-2"}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.forwardCalled)
					assert.Equal(t, "q-1", uc.forwardID)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", uc.forwardReq.DestinationServiceID)
					assert.Equal(t, "550e8400-e29b-41d4-a716-446655440001", uc.forwardReq.DestinationCounterID)
				},
			},
			{
				name:     "Negative_ForwardRejectsMissingID",
				category: "negative",
				queueID:  "",
				reqBody: model.ForwardQueueRequest{
					DestinationServiceID: "550e8400-e29b-41d4-a716-446655440000",
				},
				setup:    func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode: http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.False(t, uc.forwardCalled)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.POST("/queues/:id/forward", controller.Forward)

				body, _ := json.Marshal(tt.reqBody)
				path := "/queues/" + tt.queueID + "/forward"
				req, _ := http.NewRequest("POST", path, bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetVisitJourneys", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:     "Positive_GetVisitJourneys",
				category: "positive",
				queueID:  "q-1",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{visitRes: []model.VisitJourneyResponse{{ID: "v-1", EventType: "registration"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.getCalled)
					assert.Equal(t, "q-1", uc.getID)
				},
			},
			{
				name:     "Negative_RejectsMissingID",
				category: "negative",
				queueID:  "",
				setup:    func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/queues/:id/visit-journeys", controller.GetVisitJourneys)

				path := "/queues/" + tt.queueID + "/visit-journeys"
				req, _ := http.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetQueueStats", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			branchID string
			setup    func() *stubQueueControllerUseCase
			wantCode int
			assert   func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:     "Positive_GetQueueStats",
				category: "positive",
				branchID: "branch-1",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{statsRes: &model.QueueStatsResponse{TotalQueuesToday: 10}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.True(t, uc.statsCalled)
				},
			},
			{
				name:     "Negative_RejectsMissingBranchID",
				category: "negative",
				branchID: "",
				setup:    func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode: http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.False(t, uc.statsCalled)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/branches/:id/queue-stats", controller.GetQueueStats)

				path := "/branches/" + tt.branchID + "/queue-stats"
				req, _ := http.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetJourneysByBranchAndService", func(t *testing.T) {
		tests := []struct {
			name      string
			category  string
			branchID  string
			serviceID string
			query     string
			setup     func() *stubQueueControllerUseCase
			wantCode  int
			assert    func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:      "Positive_GetJourneysByBranchAndService",
				category:  "positive",
				branchID:  "branch-1",
				serviceID: "svc-1",
				query:     "?queue_date=2026-06-24",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{journeyRes: []model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.Equal(t, model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"}, uc.journeyReq)
				},
			},
			{
				name:      "Negative_RejectsMissingPathIDs",
				category:  "negative",
				branchID:  "",
				serviceID: "",
				query:     "",
				setup:     func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode:  http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.Equal(t, model.QueueJourneyListRequest{}, uc.journeyReq)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/branches/:id/services/:service_id/queue-journeys", controller.GetJourneysByBranchAndService)

				path := "/branches/" + tt.branchID + "/services/" + tt.serviceID + "/queue-journeys" + tt.query
				req, _ := http.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})

	t.Run("GetJourneysByBranchAndCounter", func(t *testing.T) {
		tests := []struct {
			name      string
			category  string
			branchID  string
			counterID string
			query     string
			setup     func() *stubQueueControllerUseCase
			wantCode  int
			assert    func(t *testing.T, uc *stubQueueControllerUseCase)
		}{
			{
				name:      "Positive_GetJourneysByBranchAndCounter",
				category:  "positive",
				branchID:  "branch-1",
				counterID: "c-1",
				query:     "?status=calling",
				setup: func() *stubQueueControllerUseCase {
					return &stubQueueControllerUseCase{journeyRes: []model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1"}}}
				},
				wantCode: http.StatusOK,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.Equal(t, model.QueueJourneyListRequest{CounterID: "c-1", Status: "calling"}, uc.journeyReq)
				},
			},
			{
				name:      "Negative_RejectsMissingPathIDs",
				category:  "negative",
				branchID:  "",
				counterID: "",
				query:     "",
				setup:     func() *stubQueueControllerUseCase { return &stubQueueControllerUseCase{} },
				wantCode:  http.StatusBadRequest,
				assert: func(t *testing.T, uc *stubQueueControllerUseCase) {
					assert.Equal(t, model.QueueJourneyListRequest{}, uc.journeyReq)
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup()
				log:=logrus.New()
				controller := NewQueueController(uc, newQueueTestValidator(), log)
				router := gin.New()
				router.Use(func(c *gin.Context) {
					ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
					c.Request = c.Request.WithContext(ctx)
					c.Next()
				})
				router.GET("/branches/:id/counters/:counter_id/queue-journeys", controller.GetJourneysByBranchAndCounter)

				path := "/branches/" + tt.branchID + "/counters/" + tt.counterID + "/queue-journeys" + tt.query
				req, _ := http.NewRequest("GET", path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantCode, w.Code)
				if tt.assert != nil {
					tt.assert(t, uc)
				}
			})
		}
	})
}
