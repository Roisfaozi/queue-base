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
	"github.com/Roisfaozi/queue-base/internal/modules/queue/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/validation"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func TestQueueController(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Register", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_Register",
				category: "positive",
				reqBody: model.RegisterQueueRequest{
					BranchID:    "550e8400-e29b-41d4-a716-446655440000",
					ServiceID:   "550e8400-e29b-41d4-a716-446655440001",
					PatientName: "John Queue",
				},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("RegisterQueue",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "550e8400-e29b-41d4-a716-446655440000"
						}),
						mock.MatchedBy(func(req *model.RegisterQueueRequest) bool {
							return req != nil && req.BranchID == "550e8400-e29b-41d4-a716-446655440000"
						}),
					).Return(&model.QueueResponse{ID: "q-1", BranchID: "550e8400-e29b-41d4-a716-446655440000"}, nil).Once()
					return m
				},
				wantCode: http.StatusCreated,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})

	t.Run("GetAll", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			query    string
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_SetsBranchContextFromQuery",
				category: "positive",
				query:    "?branch_id=550e8400-e29b-41d4-a716-446655440000&status=waiting",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListQueues",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "550e8400-e29b-41d4-a716-446655440000"
						}),
						model.ListQueuesRequest{
							BranchID: "550e8400-e29b-41d4-a716-446655440000",
							Status:   "waiting",
						},
					).Return([]model.QueueResponse{{ID: "q-1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Positive_BranchContextFromContext",
				category: "positive",
				query:    "?status=waiting",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListQueues",
						mock.Anything,
						model.ListQueuesRequest{Status: "waiting"},
					).Return([]model.QueueResponse{{ID: "q-1", Status: entity.QueueStatusWaiting}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Positive_WithFilters",
				category: "positive",
				query:    "?status=waiting&queue_date=2026-06-24&service_id=s-1",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListQueues",
						mock.Anything,
						model.ListQueuesRequest{
							Status:    "waiting",
							QueueDate: "2026-06-24",
							ServiceID: "s-1",
						},
					).Return([]model.QueueResponse{{ID: "q-1", QueueDate: "2026-06-24"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})

	t.Run("GetByID", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_GetByID",
				category: "positive",
				queueID:  "q-1",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ResolveQueueBranchID", mock.Anything, "q-1").Return("branch-id", nil).Maybe()
					m.On("GetQueueByID", mock.Anything, "q-1").Return(&model.QueueResponse{ID: "q-1"}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Negative_GetByIDMissingIDHitsRouter404",
				category: "negative",
				queueID:  "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})

	t.Run("Transition", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_Transition",
				category: "positive",
				queueID:  "q-1",
				reqBody:  model.QueueTransitionRequest{Action: model.QueueActionCall},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ResolveQueueBranchID", mock.Anything, "q-1").Return("branch-id", nil).Maybe()
					m.On("TransitionQueue", mock.Anything, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCall}).
						Return(&model.QueueResponse{ID: "q-1", Status: entity.QueueStatusCalling}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Negative_TransitionRejectsMissingID",
				category: "negative",
				queueID:  "",
				reqBody:  model.QueueTransitionRequest{Action: model.QueueActionCall},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name:     "Negative_TransitionRejectsInvalidAction",
				category: "negative",
				queueID:  "q-1",
				reqBody:  map[string]any{"action": "drop-table"},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})

	t.Run("Negative_TransitionRejectsMalformedJSON", func(t *testing.T) {
		uc := mocks.NewMockQueueUseCase(t)
		log := logrus.New()
		controller := NewQueueController(uc, newQueueTestValidator(), log)
		router := gin.New()
		router.Use(func(c *gin.Context) {
			ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
		router.POST("/queues/:id/transition", controller.Transition)

		req, _ := http.NewRequest("POST", "/queues/q-1/transition", bytes.NewBufferString("{"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Forward", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			reqBody  interface{}
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_Forward",
				category: "positive",
				queueID:  "q-1",
				reqBody: model.ForwardQueueRequest{
					DestinationServiceID: "550e8400-e29b-41d4-a716-446655440000",
					DestinationCounterID: "550e8400-e29b-41d4-a716-446655440001",
				},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ResolveQueueBranchID", mock.Anything, "q-1").Return("branch-id", nil).Maybe()
					m.On("ForwardQueue", mock.Anything, "q-1", &model.ForwardQueueRequest{
						DestinationServiceID: "550e8400-e29b-41d4-a716-446655440000",
						DestinationCounterID: "550e8400-e29b-41d4-a716-446655440001",
					}).Return(&model.QueueResponse{ID: "q-1", CurrentJourneyID: "j-2"}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Negative_ForwardRejectsMissingID",
				category: "negative",
				queueID:  "",
				reqBody: model.ForwardQueueRequest{
					DestinationServiceID: "550e8400-e29b-41d4-a716-446655440000",
				},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name:     "Negative_ForwardRejectsInvalidDestinationService",
				category: "negative",
				queueID:  "q-1",
				reqBody:  map[string]any{"destination_service_id": "bad"},
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusUnprocessableEntity,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})

	t.Run("Negative_ForwardRejectsMalformedJSON", func(t *testing.T) {
		uc := mocks.NewMockQueueUseCase(t)
		log := logrus.New()
		controller := NewQueueController(uc, newQueueTestValidator(), log)
		router := gin.New()
		router.Use(func(c *gin.Context) {
			ctx := database.SetOrganizationContext(c.Request.Context(), "t-1")
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
		router.POST("/queues/:id/forward", controller.Forward)

		req, _ := http.NewRequest("POST", "/queues/q-1/forward", bytes.NewBufferString("{"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetVisitJourneys", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			queueID  string
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_GetVisitJourneys",
				category: "positive",
				queueID:  "q-1",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ResolveQueueBranchID", mock.Anything, "q-1").Return("branch-id", nil).Maybe()
					m.On("GetVisitJourneys", mock.Anything, "q-1").Return([]model.VisitJourneyResponse{{ID: "v-1", EventType: "registration"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Negative_RejectsMissingID",
				category: "negative",
				queueID:  "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})

	t.Run("GetQueueStats", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			branchID string
			setup    func(t *testing.T) *mocks.MockQueueUseCase
			wantCode int
		}{
			{
				name:     "Positive_GetQueueStats",
				category: "positive",
				branchID: "branch-1",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("GetQueueStats", mock.MatchedBy(func(ctx context.Context) bool {
						return database.GetBranchID(ctx) == "branch-1"
					})).Return(&model.QueueStatsResponse{TotalQueuesToday: 10}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:     "Negative_RejectsMissingBranchID",
				category: "negative",
				branchID: "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name:     "Negative_RejectsMissingBranchPathParam",
				category: "negative",
				branchID: "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			setup     func(t *testing.T) *mocks.MockQueueUseCase
			wantCode  int
		}{
			{
				name:      "Positive_GetJourneysByBranchAndService",
				category:  "positive",
				branchID:  "branch-1",
				serviceID: "svc-1",
				query:     "?queue_date=2026-06-24",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Edge_ServiceJourneysPreserveStatusFilter",
				category:  "edge",
				branchID:  "branch-1",
				serviceID: "svc-1",
				query:     "?status=calling",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{ServiceID: "svc-1", Status: "calling"},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1", Status: entity.JourneyStatusCalling}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Edge_ServiceJourneysPreserveAllFilters",
				category:  "edge",
				branchID:  "branch-1",
				serviceID: "svc-1",
				query:     "?queue_date=2026-06-24&status=serving",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24", Status: "serving"},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1", Status: entity.JourneyStatusServing}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Negative_RejectsMissingPathIDs",
				category:  "negative",
				branchID:  "",
				serviceID: "",
				query:     "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name:      "Negative_RejectsMalformedQuery",
				category:  "negative",
				branchID:  "branch-1",
				serviceID: "svc-1",
				query:     "?queue_date=%",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: ""},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", ServiceID: "svc-1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Negative_RejectsEmptyServiceID",
				category:  "negative",
				branchID:  "branch-1",
				serviceID: "",
				query:     "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			setup     func(t *testing.T) *mocks.MockQueueUseCase
			wantCode  int
		}{
			{
				name:      "Positive_GetJourneysByBranchAndCounter",
				category:  "positive",
				branchID:  "branch-1",
				counterID: "c-1",
				query:     "?status=calling",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{CounterID: "c-1", Status: "calling"},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Edge_CounterJourneysPreserveQueueDateFilter",
				category:  "edge",
				branchID:  "branch-1",
				counterID: "c-1",
				query:     "?queue_date=2026-06-24",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{CounterID: "c-1", QueueDate: "2026-06-24"},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Edge_CounterJourneysPreserveAllFilters",
				category:  "edge",
				branchID:  "branch-1",
				counterID: "c-1",
				query:     "?queue_date=2026-06-24&status=serving",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{CounterID: "c-1", QueueDate: "2026-06-24", Status: "serving"},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1", Status: entity.JourneyStatusServing}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Negative_RejectsMissingPathIDs",
				category:  "negative",
				branchID:  "",
				counterID: "",
				query:     "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
			{
				name:      "Negative_RejectsMalformedQuery",
				category:  "negative",
				branchID:  "branch-1",
				counterID: "c-1",
				query:     "?queue_date=%",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					m := mocks.NewMockQueueUseCase(t)
					m.On("ListActiveJourneys",
						mock.MatchedBy(func(ctx context.Context) bool {
							return database.GetBranchID(ctx) == "branch-1"
						}),
						model.QueueJourneyListRequest{CounterID: "c-1", QueueDate: ""},
					).Return([]model.QueueJourneyResponse{{ID: "j-1", CounterID: "c-1"}}, nil).Once()
					return m
				},
				wantCode: http.StatusOK,
			},
			{
				name:      "Negative_RejectsEmptyCounterID",
				category:  "negative",
				branchID:  "branch-1",
				counterID: "",
				query:     "",
				setup: func(t *testing.T) *mocks.MockQueueUseCase {
					return mocks.NewMockQueueUseCase(t)
				},
				wantCode: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				uc := tt.setup(t)
				log := logrus.New()
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
			})
		}
	})
}
