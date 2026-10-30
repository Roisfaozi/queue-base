package usecase_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Helpers
// =============================================================================

type mockQueueRepoConfig struct {
	FindQueueByTenantIDFunc func(ctx context.Context, tenantID, queueID string) (*entity.Queue, error)
	FindCurrentJourneyFunc  func(ctx context.Context, tenantID, branchID, queueID, journeyID string) (*entity.QueueJourney, error)
	q                       *entity.Queue
	queues                  []*entity.Queue
	j                       *entity.QueueJourney
	visit                   *entity.VisitJourney
	err                     error
	seenNum                 int
	exists                  bool
	journeyList             []*entity.QueueJourney
	visits                  []*entity.VisitJourney
	statsRes                model.QueueStatsResponse
	nextWaiting             *entity.Queue
	waitingLeft             int
}

type mockRepoTracker struct {
	Queue           *entity.Queue
	Journey         *entity.QueueJourney
	Visit           *entity.VisitJourney
	SeenNum         int
	LastPrefix      string
	ListReq         model.ListQueuesRequest
	JourneyReq      model.QueueJourneyListRequest
	JourneyTenantID string
	JourneyBranchID string
}

func newMockQueueRepo(t *testing.T, cfg mockQueueRepoConfig) (*mocks.MockQueueRepository, *mockRepoTracker) {
	mockRepo := mocks.NewMockQueueRepository(t)
	tracker := &mockRepoTracker{
		Queue:   cfg.q,
		Journey: cfg.j,
		Visit:   cfg.visit,
		SeenNum: cfg.seenNum,
	}

	mockRepo.On("NextQueueNumber", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error) {
			tracker.SeenNum++
			tracker.LastPrefix = prefix
			if cfg.err != nil {
				return 0, cfg.err
			}
			return tracker.SeenNum, nil
		},
	)

	mockRepo.On("ExistsRegistration", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueDate, patientID, patientName string) (bool, error) {
			if cfg.err != nil {
				return false, cfg.err
			}
			return cfg.exists, nil
		},
	)

	mockRepo.On("CreateRegistration", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error {
			tracker.Queue = queue
			tracker.Journey = journey
			tracker.Visit = visit
			return cfg.err
		},
	)

	mockRepo.On("CreateRegistrationWithNumber", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney, date time.Time, prefix string) error {
			tracker.SeenNum++
			tracker.LastPrefix = prefix
			queue.QueueNo = tracker.SeenNum
			queue.TicketNo = fmt.Sprintf("%s%03d", prefix, queue.QueueNo)
			tracker.Queue = queue
			tracker.Journey = journey
			tracker.Visit = visit
			return cfg.err
		},
	)

	mockRepo.On("ListQueues", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID string, req model.ListQueuesRequest) ([]*entity.Queue, error) {
			tracker.ListReq = req
			if cfg.err != nil {
				return nil, cfg.err
			}
			if req.Status == "" && req.QueueDate == "" && req.ServiceID == "" {
				return cfg.queues, nil
			}
			filtered := make([]*entity.Queue, 0)
			for _, queue := range cfg.queues {
				if req.Status != "" && queue.Status != req.Status {
					continue
				}
				if req.QueueDate != "" && queue.QueueDate != req.QueueDate {
					continue
				}
				if req.ServiceID != "" && queue.CurrentJourneyID != req.ServiceID {
					continue
				}
				filtered = append(filtered, queue)
			}
			return filtered, nil
		},
	)

	mockRepo.On("ListActiveJourneys", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID string, req model.QueueJourneyListRequest) ([]*entity.QueueJourney, error) {
			if cfg.err != nil {
				return nil, cfg.err
			}
			tracker.JourneyReq = req
			tracker.JourneyTenantID = tenantID
			tracker.JourneyBranchID = branchID
			if cfg.journeyList != nil {
				return cfg.journeyList, nil
			}
			return []*entity.QueueJourney{{ID: "j-1", QueueID: "q-1", TenantID: tenantID, ServiceID: req.ServiceID, CounterID: req.CounterID, SeqNo: 1, Status: entity.JourneyStatusCalling}}, nil
		},
	)

	mockRepo.On("FindVisitJourneysByQueueID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueID string) ([]*entity.VisitJourney, error) {
			if cfg.err != nil {
				return nil, cfg.err
			}
			return cfg.visits, nil
		},
	)

	mockRepo.On("GetQueueStats", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueDate string) (model.QueueStatsResponse, error) {
			if cfg.err != nil {
				return cfg.statsRes, cfg.err
			}
			return cfg.statsRes, nil
		},
	)

	mockRepo.On("CountWaitingQueueLeft", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueDate, serviceID string, queueNo int) (int, error) {
			if cfg.err != nil {
				return 0, cfg.err
			}
			return cfg.waitingLeft, nil
		},
	)

	mockRepo.On("FindQueueByTenantID", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, queueID string) (*entity.Queue, error) {
			if cfg.FindQueueByTenantIDFunc != nil {
				return cfg.FindQueueByTenantIDFunc(ctx, tenantID, queueID)
			}
			return nil, nil
		},
	)

	mockRepo.On("FindQueueByID", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueID string) (*entity.Queue, error) {
			if cfg.err != nil {
				return nil, cfg.err
			}
			return tracker.Queue, nil
		},
	)

	mockRepo.On("FindCurrentJourney", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueID, journeyID string) (*entity.QueueJourney, error) {
			if cfg.FindCurrentJourneyFunc != nil {
				return cfg.FindCurrentJourneyFunc(ctx, tenantID, branchID, queueID, journeyID)
			}
			if cfg.err != nil {
				return nil, cfg.err
			}
			return tracker.Journey, nil
		},
	)

	mockRepo.On("FindNextWaitingQueue", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueDate, afterQueueID string) (*entity.Queue, error) {
			if cfg.err != nil {
				return nil, cfg.err
			}
			if cfg.nextWaiting == nil {
				return nil, exception.ErrNotFound
			}
			return cfg.nextWaiting, nil
		},
	)

	mockRepo.On("NextJourneySequence", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, tenantID, branchID, queueID string) (int, error) {
			if cfg.err != nil {
				return 0, cfg.err
			}
			if tracker.Journey == nil {
				return 1, nil
			}
			return tracker.Journey.SeqNo + 1, nil
		},
	)

	mockRepo.On("CreateForwarding", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, nextJourney *entity.QueueJourney, visit *entity.VisitJourney) error {
			tracker.Queue = queue
			tracker.Journey = nextJourney
			tracker.Visit = visit
			return cfg.err
		},
	)

	mockRepo.On("UpdateQueueState", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, visit *entity.VisitJourney) error {
			tracker.Queue = queue
			tracker.Journey = currentJourney
			tracker.Visit = visit
			return cfg.err
		},
	)

	return mockRepo, tracker
}

func newMockSettingsResolver(t *testing.T, values map[string]string) (*mocks.MockSettingsResolver, *[]string) {
	mockResolver := mocks.NewMockSettingsResolver(t)
	calls := &[]string{}
	mockResolver.On("Resolve", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(
		func(ctx context.Context, key, branchID, serviceID, counterID string) (string, error) {
			*calls = append(*calls, key)
			if val, ok := values[key]; ok {
				return val, nil
			}
			return "", exception.ErrNotFound
		},
	)
	return mockResolver, calls
}

func newMockRelationValidator(t *testing.T, err error) *mocks.MockRelationValidator {
	mockVal := mocks.NewMockRelationValidator(t)
	mockVal.On("Validate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(err)
	return mockVal
}

func newMockAuditLogger(t *testing.T, err error) (*mocks.MockAuditLogger, *[]auditModel.CreateAuditLogRequest) {
	mockAudit := mocks.NewMockAuditLogger(t)
	entries := &[]auditModel.CreateAuditLogRequest{}
	mockAudit.On("LogActivity", mock.Anything, mock.Anything).Maybe().Run(func(args mock.Arguments) {
		req := args.Get(1).(auditModel.CreateAuditLogRequest)
		*entries = append(*entries, req)
	}).Return(err)
	return mockAudit, entries
}

func newMockBroadcaster(t *testing.T) (*mocks.MockEventBroadcaster, *[]string) {
	mockB := mocks.NewMockEventBroadcaster(t)
	events := &[]string{}
	mockB.On("Broadcast", mock.Anything, mock.Anything).Maybe().Run(func(args mock.Arguments) {
		name := args.String(0)
		*events = append(*events, name)
	}).Return()
	return mockB, events
}

func newMockWSBroadcaster(t *testing.T) (*mocks.MockWSBroadcaster, *[]string) {
	mockWS := mocks.NewMockWSBroadcaster(t)
	channels := &[]string{}
	mockWS.On("BroadcastToChannel", mock.Anything, mock.Anything).Maybe().Run(func(args mock.Arguments) {
		channel := args.String(0)
		*channels = append(*channels, channel)
	}).Return()
	return mockWS, channels
}

func TestQueueAuditLogging(t *testing.T) {
	t.Run("Edge_Register_EmitsAuditAndSurvivesAuditFailure", func(t *testing.T) {
		mockRepo, _ := newMockQueueRepo(t, mockQueueRepoConfig{})
		audit, entries := newMockAuditLogger(t, assert.AnError)
		broadcast, events := newMockBroadcaster(t)
		wsBroadcast, channels := newMockWSBroadcaster(t)
		mockResolver, _ := newMockSettingsResolver(t, nil)
		uc := usecase.NewQueueUseCase(mockRepo, mockResolver, nil, audit)
		uc.SetEventBroadcaster(broadcast)
		uc.SetWSBroadcaster(wsBroadcast)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		ctx = authcontext.WithUserID(ctx, "u-1")

		res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John"})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, *entries, 1)
		require.Len(t, *events, 1)
		require.Len(t, *channels, 1)
		assert.Equal(t, "queue_registered", (*events)[0])
		assert.Equal(t, "queue:t-1:b-1", (*channels)[0])
		values, ok := (*entries)[0].NewValues.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "QUEUE_REGISTER", (*entries)[0].Action)
		assert.Equal(t, "queue", (*entries)[0].Entity)
		assert.Equal(t, "t-1", (*entries)[0].OrganizationID)
		assert.Equal(t, "u-1", (*entries)[0].UserID)
		assert.Equal(t, "b-1", values["branch_id"])
		assert.Equal(t, res.TicketNo, values["ticket_no"])
	})

	t.Run("Edge_Forward_EmitsAudit", func(t *testing.T) {
		mockRepo, _ := newMockQueueRepo(t, mockQueueRepoConfig{
			q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
			j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
		})
		audit, entries := newMockAuditLogger(t, nil)
		broadcast, events := newMockBroadcaster(t)
		wsBroadcast, _ := newMockWSBroadcaster(t)
		validator := newMockRelationValidator(t, nil)
		uc := usecase.NewQueueUseCase(mockRepo, nil, validator, audit)
		uc.SetEventBroadcaster(broadcast)
		uc.SetWSBroadcaster(wsBroadcast)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "svc-2", DestinationCounterID: "ctr-2"})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, *entries, 1)
		require.Len(t, *events, 1)
		assert.Equal(t, "queue_forwarded", (*events)[0])
		values, ok := (*entries)[0].NewValues.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "QUEUE_FORWARD", (*entries)[0].Action)
		assert.Equal(t, "system", (*entries)[0].UserID)
		assert.Equal(t, "q-1", (*entries)[0].EntityID)
		assert.Equal(t, "j-1", values["from_journey_id"])
		assert.Equal(t, "svc-2", values["to_service_id"])
	})

	t.Run("Edge_Transition_EmitsAudit", func(t *testing.T) {
		mockRepo, _ := newMockQueueRepo(t, mockQueueRepoConfig{
			q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
			j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
		})
		audit, entries := newMockAuditLogger(t, nil)
		broadcast, events := newMockBroadcaster(t)
		uc := usecase.NewQueueUseCase(mockRepo, nil, nil, audit)
		uc.SetEventBroadcaster(broadcast)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCall})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, *entries, 1)
		require.Len(t, *events, 1)
		assert.Equal(t, "queue_transitioned", (*events)[0])
		values, ok := (*entries)[0].NewValues.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "QUEUE_CALL", (*entries)[0].Action)
		assert.Equal(t, "j-1", values["journey_id"])
		assert.Equal(t, entity.QueueStatusCalling, values["status"])
	})

	t.Run("Edge_Register_AuditFailureDoesNotFailBusinessFlow", func(t *testing.T) {
		mockRepo, _ := newMockQueueRepo(t, mockQueueRepoConfig{})
		audit, entries := newMockAuditLogger(t, assert.AnError)
		mockResolver, _ := newMockSettingsResolver(t, nil)
		uc := usecase.NewQueueUseCase(mockRepo, mockResolver, nil, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John"})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, *entries, 1)
	})
}

// =============================================================================
// TestRegisterQueue
// =============================================================================

func TestRegisterQueue(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      *model.RegisterQueueRequest
		repoCfg  mockQueueRepoConfig
		settings map[string]string
		tenantID string
		branchID string
		wantErr  error
		wantRes  func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John Doe"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, "t-1", res.TenantID)
				assert.Equal(t, "b-1", res.BranchID)
				assert.Equal(t, 1, res.QueueNo)
				assert.Equal(t, entity.QueueStatusWaiting, res.Status)
			},
		},
		{
			name:     "Positive_UsesQueueResetTimeKeyFirst",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"queue_reset_time": "05:00"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, "queue_reset_time", resolverCalls[0])
			},
		},
		{
			name:     "Negative_IgnoresLegacyResetTimeKeyInUsecase",
			category: "negative",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"reset_time": "05:00"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, "queue_reset_time", resolverCalls[0])
			},
		},
		{
			name:     "Positive_UsesTicketPrefixSettingFirst",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"ticket_prefix": "JS"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, "JS", tracker.LastPrefix)
				assert.Equal(t, "JS001", res.TicketNo)
				assert.Contains(t, resolverCalls, "ticket_prefix")
			},
		},
		{
			name:     "Positive_DefaultsPrefixToA",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, "A", tracker.LastPrefix)
				assert.Equal(t, "A001", res.TicketNo)
			},
		},
		{
			name:     "Positive_UsesNumberingStrategySettingFirst",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"numbering_strategy": "sequential"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, 1, res.QueueNo)
				assert.Contains(t, resolverCalls, "numbering_strategy")
			},
		},
		{
			name:     "Edge_InvalidNumberingStrategyFallsBackToSequential",
			category: "edge",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"numbering_strategy": "random"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, 1, res.QueueNo)
				assert.Equal(t, "A001", res.TicketNo)
			},
		},
		{
			name:     "Vulnerability_NumberingStrategyDoesNotBypassTenantBranchScope",
			category: "vulnerability",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"numbering_strategy": "sequential"},
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_DuplicateReturnsConflict",
			category: "negative",
			req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John Doe"},
			repoCfg:  mockQueueRepoConfig{exists: true},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrConflict,
		},
		{
			name:     "Negative_NoTenant",
			category: "negative",
			req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John"},
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_NoBranch",
			category: "negative",
			req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John"},
			tenantID: "t-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_SQLInjection",
			category: "vulnerability",
			req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John'; DROP TABLE queues;--"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.Equal(t, "John&#39;; DROP TABLE queues;--", res.PatientName)
			},
		},
		{
			name:     "Positive_PharmacyServiceStillCreatesSingleMasterQueue",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "pharmacy-svc", PatientName: "John Doe"},
			settings: map[string]string{"queue_reset_time": "04:00", "ticket_prefix": "P"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				assert.NotEmpty(t, tracker.Queue)
				assert.NotEmpty(t, tracker.Journey)
				assert.NotEmpty(t, tracker.Visit)
				assert.Equal(t, "P", tracker.LastPrefix)
			},
		},
		{
			name:     "Positive_EstimateUsesQueueLeftTimesEffectiveDuration",
			category: "positive",
			repoCfg:  mockQueueRepoConfig{waitingLeft: 4},
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"default_estimated_duration": "5"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, resolverCalls []string, res *model.QueueResponse) {
				require.NotNil(t, res.QueueLeft)
				require.NotNil(t, res.EstimateMinutes)
				assert.Equal(t, 4, *res.QueueLeft)
				assert.Equal(t, 20, *res.EstimateMinutes)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, tracker := newMockQueueRepo(t, tt.repoCfg)
			resolver, calls := newMockSettingsResolver(t, tt.settings)
			uc := usecase.NewQueueUseCase(mockRepo, resolver, nil)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.RegisterQueue(ctx, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, res)
			if tt.wantRes != nil {
				tt.wantRes(t, tracker, *calls, res)
			}
		})
	}
}

func TestAttachQueueEstimate(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		repoCfg      mockQueueRepoConfig
		settings     map[string]string
		res          model.QueueResponse
		wantLeft     int
		wantEstimate int
	}{
		{
			name:     "Positive_WaitingUsesQueueLeftTimesEffectiveDuration",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				j:           &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "svc-1", Status: entity.JourneyStatusPending},
				waitingLeft: 4,
			},
			settings:     map[string]string{"default_estimated_duration": "5"},
			res:          model.QueueResponse{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1", QueueDate: "2026-07-07", QueueNo: 5, Status: entity.QueueStatusWaiting},
			wantLeft:     4,
			wantEstimate: 20,
		},
		{
			name:         "Edge_ServingReturnsZeroQueueLeftAndEstimate",
			category:     "edge",
			repoCfg:      mockQueueRepoConfig{waitingLeft: 4},
			settings:     map[string]string{"default_estimated_duration": "5"},
			res:          model.QueueResponse{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1", QueueDate: "2026-07-07", QueueNo: 5, Status: entity.QueueStatusServing},
			wantLeft:     0,
			wantEstimate: 0,
		},
		{
			name:     "Edge_InvalidDurationReturnsZeroEstimate",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				j:           &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "svc-1", Status: entity.JourneyStatusPending},
				waitingLeft: 4,
			},
			settings:     map[string]string{"default_estimated_duration": "bad"},
			res:          model.QueueResponse{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1", QueueDate: "2026-07-07", QueueNo: 5, Status: entity.QueueStatusWaiting},
			wantLeft:     4,
			wantEstimate: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, _ := newMockQueueRepo(t, tt.repoCfg)
			resolver, _ := newMockSettingsResolver(t, tt.settings)
			uc := usecase.NewQueueUseCase(mockRepo, resolver, nil)
			res := tt.res

			err := usecase.AttachQueueEstimate(context.Background(), uc, &res)

			require.NoError(t, err)
			require.NotNil(t, res.QueueLeft)
			require.NotNil(t, res.EstimateMinutes)
			assert.Equal(t, tt.wantLeft, *res.QueueLeft)
			assert.Equal(t, tt.wantEstimate, *res.EstimateMinutes)
		})
	}
}

// =============================================================================
// TestForwardQueue
// =============================================================================

func TestForwardQueue(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		repoCfg      mockQueueRepoConfig
		validatorErr error
		hasValidator bool
		queueID      string
		req          *model.ForwardQueueRequest
		tenantID     string
		branchID     string
		wantErr      error
		wantRes      func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", TicketNo: "A001", QueueNo: 1, Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-2"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, "q-1", res.ID)
				assert.NotEqual(t, "j-1", res.CurrentJourneyID)
			},
		},
		{
			name:     "Negative_NoTenant",
			category: "negative",
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-2"},
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Edge_SameServiceStillCreatesJourney",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-1"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.NotEmpty(t, res.CurrentJourneyID)
			},
		},
		{
			name:     "Edge_MissingActiveJourneyRejected",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-missing"},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-2"},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Vulnerability_CrossTenantRejected",
			category: "vulnerability",
			repoCfg:  mockQueueRepoConfig{err: exception.ErrNotFound},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-2"},
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Negative_InvalidDestinationRelationRejected",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			hasValidator: true,
			validatorErr: exception.ErrForbidden,
			queueID:      "q-1",
			req:          &model.ForwardQueueRequest{DestinationServiceID: "s-2", DestinationCounterID: "c-2"},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantErr:      exception.ErrForbidden,
		},
		{
			name:     "Positive_PharmacyFlowKeepsSingleQueueRecord",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "pharmacy-svc"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, "q-1", res.ID)
				assert.NotEmpty(t, tracker.Queue)
				assert.NotEqual(t, "j-1", tracker.Journey.ID)
				assert.NotEmpty(t, tracker.Visit)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, tracker := newMockQueueRepo(t, tt.repoCfg)
			var val usecase.RelationValidator
			if tt.hasValidator {
				val = newMockRelationValidator(t, tt.validatorErr)
			}
			uc := usecase.NewQueueUseCase(mockRepo, nil, val)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.ForwardQueue(ctx, tt.queueID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, tracker, res)
			}
		})
	}
}

// =============================================================================
// TestTransitionQueue
// =============================================================================

func TestTransitionQueue(t *testing.T) {
	tests := []struct {
		name     string
		category string
		repoCfg  mockQueueRepoConfig
		settings map[string]string
		queueID  string
		req      *model.QueueTransitionRequest
		tenantID string
		branchID string
		wantErr  error
		wantRes  func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse)
	}{
		{
			name:     "Positive_SuccessCalling",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCall},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCalling, res.Status)
				assert.Equal(t, entity.JourneyStatusCalling, tracker.Journey.Status)
				assert.Equal(t, "call", tracker.Visit.EventType)
			},
		},
		{
			name:     "Positive_SuccessCancel",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCanceled, res.Status)
			},
		},
		{
			name:     "Edge_RecallWhenAllowRecallTrue",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			settings: map[string]string{"allow_recall": "true"},
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCall},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCalling, res.Status)
				assert.Equal(t, "recall", tracker.Visit.EventType)
			},
		},
		{
			name:     "Negative_RecallWhenAllowRecallFalse",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			settings: map[string]string{"allow_recall": "false"},
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCall},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_RecallRejectsMismatchedJourneyState",
			category: "vulnerability",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			settings: map[string]string{"allow_recall": "true"},
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCall},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_SkipWhenAllowSkipFalse",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			settings: map[string]string{"allow_skip": "false"},
			req:      &model.QueueTransitionRequest{Action: model.QueueActionSkip},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_CancelWhenAllowCancelFalse",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			settings: map[string]string{"allow_cancel": "false"},
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Positive_SuccessSkip",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionSkip},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusSkipped, res.Status)
			},
		},
		{
			name:     "Edge_ServingFromCalling",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionServe},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusServing, res.Status)
			},
		},
		{
			name:     "Edge_SkipFromCalling",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionSkip},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusSkipped, res.Status)
				assert.Equal(t, entity.JourneyStatusSkipped, tracker.Journey.Status)
				assert.Equal(t, "skip", tracker.Visit.EventType)
			},
		},
		{
			name:     "Edge_CancelFromWaiting",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCanceled, res.Status)
				assert.Equal(t, entity.JourneyStatusCanceled, tracker.Journey.Status)
				assert.Equal(t, "cancel", tracker.Visit.EventType)
			},
		},
		{
			name:     "Edge_CompleteAfterServing",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusServing, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusServing},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionComplete},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCompleted, res.Status)
			},
		},
		{
			name:     "Edge_AutoCallNextAfterComplete",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q:           &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusServing, CurrentJourneyID: "j-1", QueueDate: "2026-07-03"},
				j:           &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusServing, ServiceID: "svc-1"},
				nextWaiting: &entity.Queue{ID: "q-2", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-2", QueueDate: "2026-07-03"},
				FindCurrentJourneyFunc: func(ctx context.Context, tenantID, branchID, queueID, journeyID string) (*entity.QueueJourney, error) {
					if queueID == "q-2" {
						return &entity.QueueJourney{ID: "j-2", QueueID: "q-2", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending, ServiceID: "svc-1"}, nil
					}
					return &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusServing, ServiceID: "svc-1"}, nil
				},
			},
			settings: map[string]string{"auto_call_next": "true"},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionComplete},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCompleted, res.Status)
				assert.Equal(t, "call", tracker.Visit.EventType)
			},
		},
		{
			name:     "Negative_InvalidStateChange",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCompleted, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCompleted},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCall},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_ServeFromSkippedRejected",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusSkipped, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusSkipped},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionServe},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_ServeRejectsMismatchedJourneyState",
			category: "vulnerability",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionServe},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_CompleteFromWaitingRejected",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionComplete},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Edge_CancelTerminalStateRejected",
			category: "edge",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCanceled, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCanceled},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_EmptyAction",
			category: "negative",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: ""},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_CrossTenantRejected",
			category: "vulnerability",
			repoCfg:  mockQueueRepoConfig{err: exception.ErrNotFound},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, tracker := newMockQueueRepo(t, tt.repoCfg)
			var settings usecase.SettingsResolver
			if tt.settings != nil {
				settings, _ = newMockSettingsResolver(t, tt.settings)
			}
			uc := usecase.NewQueueUseCase(mockRepo, settings, nil)
			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.TransitionQueue(ctx, tt.queueID, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, tracker, res)
			}
		})
	}
}

func TestQueueJourneyLifecycle(t *testing.T) {
	mockRepo, tracker := newMockQueueRepo(t, mockQueueRepoConfig{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "svc-1", Status: entity.JourneyStatusPending},
	})
	mockSettings, _ := newMockSettingsResolver(t, map[string]string{"allow_recall": "true", "allow_skip": "true"})
	uc := usecase.NewQueueUseCase(mockRepo, mockSettings, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	steps := []struct {
		name       string
		action     string
		wantStatus string
		wantEvent  string
	}{
		{name: "call", action: model.QueueActionCall, wantStatus: entity.QueueStatusCalling, wantEvent: "call"},
		{name: "recall", action: model.QueueActionCall, wantStatus: entity.QueueStatusCalling, wantEvent: "recall"},
		{name: "serve", action: model.QueueActionServe, wantStatus: entity.QueueStatusServing, wantEvent: "serve"},
		{name: "complete", action: model.QueueActionComplete, wantStatus: entity.QueueStatusCompleted, wantEvent: "complete"},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: step.action})

			require.NoError(t, err)
			assert.Equal(t, step.wantStatus, res.Status)
			assert.Equal(t, step.wantEvent, tracker.Visit.EventType)
		})
	}
}

// =============================================================================
// TestListQueues
// =============================================================================

func TestListQueues(t *testing.T) {
	tests := []struct {
		name     string
		category string
		repoCfg  mockQueueRepoConfig
		req      model.ListQueuesRequest
		tenantID string
		branchID string
		wantErr  error
		wantLen  int
		wantID   string
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repoCfg:  mockQueueRepoConfig{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}},
			req:      model.ListQueuesRequest{},
			tenantID: "t-1",
			branchID: "b-1",
			wantLen:  1,
			wantID:   "q-1",
		},
		{
			name:     "Negative_MissingTenantOrBranch",
			category: "negative",
			repoCfg:  mockQueueRepoConfig{queues: []*entity.Queue{{ID: "q-1"}}},
			req:      model.ListQueuesRequest{},
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Edge_StatusFilter",
			category: "edge",
			repoCfg:  mockQueueRepoConfig{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}, {ID: "q-2", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCompleted}}},
			req:      model.ListQueuesRequest{Status: entity.QueueStatusCompleted},
			tenantID: "t-1",
			branchID: "b-1",
			wantLen:  1,
			wantID:   "q-2",
		},
		{
			name:     "Positive_QueueDateFilter",
			category: "positive",
			repoCfg:  mockQueueRepoConfig{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting}, {ID: "q-2", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-23", Status: entity.QueueStatusWaiting}}},
			req:      model.ListQueuesRequest{QueueDate: "2026-06-24"},
			tenantID: "t-1",
			branchID: "b-1",
			wantLen:  1,
			wantID:   "q-1",
		},
		{
			name:     "Vulnerability_PassesScopedFiltersOnly",
			category: "vulnerability",
			repoCfg:  mockQueueRepoConfig{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", CurrentJourneyID: "svc-1"}}},
			req:      model.ListQueuesRequest{Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", ServiceID: "svc-1"},
			tenantID: "tenant-safe",
			branchID: "branch-safe",
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, tracker := newMockQueueRepo(t, tt.repoCfg)
			uc := usecase.NewQueueUseCase(mockRepo, nil, nil)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.ListQueues(ctx, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Len(t, res, tt.wantLen)
			if tt.wantID != "" {
				assert.Equal(t, tt.wantID, res[0].ID)
			}
			if tt.name == "Vulnerability_PassesScopedFiltersOnly" {
				assert.Equal(t, tt.req, tracker.ListReq)
			}
		})
	}
}

// =============================================================================
// TestListActiveJourneys
// =============================================================================

func TestListActiveJourneys(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		repoCfg      mockQueueRepoConfig
		validatorErr error
		hasValidator bool
		req          model.QueueJourneyListRequest
		tenantID     string
		branchID     string
		wantErr      error
		wantRes      func(t *testing.T, tracker *mockRepoTracker, res []model.QueueJourneyResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res []model.QueueJourneyResponse) {
				require.Len(t, res, 1)
				assert.Equal(t, "svc-1", res[0].ServiceID)
				assert.Equal(t, entity.JourneyStatusCalling, res[0].Status)
			},
		},
		{
			name:     "Negative_MissingTenantOrBranch",
			category: "negative",
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1"},
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:         "Negative_InvalidServiceRelationRejected",
			category:     "negative",
			hasValidator: true,
			validatorErr: exception.ErrForbidden,
			req:          model.QueueJourneyListRequest{ServiceID: "svc-foreign"},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantErr:      exception.ErrForbidden,
		},
		{
			name:     "Edge_CounterFilter",
			category: "edge",
			req:      model.QueueJourneyListRequest{CounterID: "counter-1"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res []model.QueueJourneyResponse) {
				require.Len(t, res, 1)
				assert.Equal(t, "counter-1", res[0].CounterID)
			},
		},
		{
			name:     "Edge_EmptyList",
			category: "edge",
			repoCfg:  mockQueueRepoConfig{journeyList: []*entity.QueueJourney{}},
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res []model.QueueJourneyResponse) {
				assert.Empty(t, res)
			},
		},
		{
			name:         "Negative_InvalidCounterRelationRejected",
			category:     "negative",
			hasValidator: true,
			validatorErr: exception.ErrForbidden,
			req:          model.QueueJourneyListRequest{CounterID: "counter-foreign"},
			tenantID:     "t-1",
			branchID:     "b-1",
			wantErr:      exception.ErrForbidden,
		},
		{
			name:     "Vulnerability_ScopedTenantPassThrough",
			category: "vulnerability",
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1"},
			tenantID: "tenant-safe",
			branchID: "branch-safe",
			wantRes: func(t *testing.T, tracker *mockRepoTracker, res []model.QueueJourneyResponse) {
				require.Len(t, res, 1)
				assert.Equal(t, "tenant-safe", tracker.JourneyTenantID)
				assert.Equal(t, "branch-safe", tracker.JourneyBranchID)
				assert.Equal(t, model.QueueJourneyListRequest{ServiceID: "svc-1"}, tracker.JourneyReq)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, tracker := newMockQueueRepo(t, tt.repoCfg)
			var val usecase.RelationValidator
			if tt.hasValidator {
				val = newMockRelationValidator(t, tt.validatorErr)
			}
			uc := usecase.NewQueueUseCase(mockRepo, nil, val)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.ListActiveJourneys(ctx, tt.req)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, tracker, res)
			}
		})
	}
}

// =============================================================================
// TestGetVisitJourneys
// =============================================================================

func TestGetVisitJourneys(t *testing.T) {
	tests := []struct {
		name     string
		category string
		repoCfg  mockQueueRepoConfig
		queueID  string
		tenantID string
		branchID string
		wantErr  error
		wantRes  func(t *testing.T, res []model.VisitJourneyResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repoCfg: mockQueueRepoConfig{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1"},
				visits: []*entity.VisitJourney{
					{ID: "v-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", EventType: "registration", CreatedAt: 100},
					{ID: "v-2", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", EventType: "call", CreatedAt: 200},
				},
			},
			queueID:  "q-1",
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, res []model.VisitJourneyResponse) {
				require.Len(t, res, 2)
				assert.Equal(t, "v-1", res[0].ID)
				assert.Equal(t, "v-2", res[1].ID)
			},
		},
		{
			name:     "Negative_MissingTenant",
			category: "negative",
			queueID:  "q-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_EmptyQueueID",
			category: "negative",
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Edge_EmptyList",
			category: "edge",
			repoCfg:  mockQueueRepoConfig{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1"}},
			queueID:  "q-1",
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, res []model.VisitJourneyResponse) {
				assert.Empty(t, res)
			},
		},
		{
			name:     "Negative_QueueLookupFailureStopsVisitsRead",
			category: "negative",
			repoCfg:  mockQueueRepoConfig{err: exception.ErrNotFound, visits: []*entity.VisitJourney{{ID: "v-1"}}},
			queueID:  "q-1",
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Vulnerability_CrossTenantRejected",
			category: "vulnerability",
			repoCfg:  mockQueueRepoConfig{err: exception.ErrNotFound},
			queueID:  "q-1",
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Negative_QueueLookupFailureStopsVisitsRead",
			category: "negative",
			repoCfg:  mockQueueRepoConfig{err: exception.ErrNotFound, visits: []*entity.VisitJourney{{ID: "v-1"}}},
			queueID:  "q-1",
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, _ := newMockQueueRepo(t, tt.repoCfg)
			uc := usecase.NewQueueUseCase(mockRepo, nil, nil)
			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.GetVisitJourneys(ctx, tt.queueID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

// =============================================================================
// TestGetQueueByID
// =============================================================================

func TestGetQueueByID(t *testing.T) {
	tests := []struct {
		name     string
		category string
		repoCfg  mockQueueRepoConfig
		tenantID string
		branchID string
		wantErr  error
		wantID   string
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repoCfg:  mockQueueRepoConfig{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}},
			tenantID: "t-1",
			branchID: "b-1",
			wantID:   "q-1",
		},
		{
			name:     "Negative_MissingBranch",
			category: "negative",
			repoCfg:  mockQueueRepoConfig{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}},
			tenantID: "t-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_EmptyQueueID",
			category: "negative",
			repoCfg:  mockQueueRepoConfig{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
			wantID:   "",
		},
		{
			name:     "Vulnerability_CrossTenantRejected",
			category: "vulnerability",
			repoCfg:  mockQueueRepoConfig{err: exception.ErrNotFound},
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, _ := newMockQueueRepo(t, tt.repoCfg)
			uc := usecase.NewQueueUseCase(mockRepo, nil, nil)
			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			queueID := "q-1"
			if tt.name == "Negative_EmptyQueueID" {
				queueID = ""
			}

			res, err := uc.GetQueueByID(ctx, queueID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantID, res.ID)
		})
	}
}

// =============================================================================
// TestGetQueueStats
// =============================================================================

func TestGetQueueStats(t *testing.T) {
	tests := []struct {
		name         string
		category     string
		repoCfg      mockQueueRepoConfig
		validatorErr error
		hasValidator bool
		tenantID     string
		branchID     string
		wantErr      error
		wantRes      func(t *testing.T, res *model.QueueStatsResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repoCfg:  mockQueueRepoConfig{statsRes: model.QueueStatsResponse{TotalQueuesToday: 5, TotalActiveJourneys: 3}},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, res *model.QueueStatsResponse) {
				assert.Equal(t, int64(5), res.TotalQueuesToday)
				assert.Equal(t, int64(3), res.TotalActiveJourneys)
			},
		},
		{
			name:         "Negative_InvalidBranchRelationRejected",
			category:     "negative",
			hasValidator: true,
			validatorErr: exception.ErrForbidden,
			tenantID:     "t-1",
			branchID:     "branch-foreign",
			wantErr:      exception.ErrForbidden,
		},
		{
			name:     "Edge_ZeroStatsAllowed",
			category: "edge",
			repoCfg:  mockQueueRepoConfig{statsRes: model.QueueStatsResponse{}},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, res *model.QueueStatsResponse) {
				require.NotNil(t, res)
				assert.Equal(t, int64(0), res.TotalQueuesToday)
				assert.Equal(t, int64(0), res.TotalActiveJourneys)
			},
		},
		{
			name:     "Negative_NoTenantOrBranch",
			category: "negative",
			wantErr:  exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, _ := newMockQueueRepo(t, tt.repoCfg)
			var val usecase.RelationValidator
			if tt.hasValidator {
				val = newMockRelationValidator(t, tt.validatorErr)
			}
			uc := usecase.NewQueueUseCase(mockRepo, nil, val)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}

			res, err := uc.GetQueueStats(ctx)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}

// =============================================================================
// TestComputeBusinessQueueDate
// =============================================================================

func TestComputeBusinessQueueDate(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)

	tests := []struct {
		name      string
		time      time.Time
		resetTime string
		want      string
	}{
		{
			name:      "Edge_BeforeResetUsesPreviousDate",
			time:      time.Date(2026, 6, 24, 3, 59, 0, 0, loc),
			resetTime: "04:00",
			want:      "2026-06-23",
		},
		{
			name:      "Positive_AfterResetUsesCurrentDate",
			time:      time.Date(2026, 6, 24, 4, 1, 0, 0, loc),
			resetTime: "04:00",
			want:      "2026-06-24",
		},
		{
			name:      "Negative_InvalidResetFallsBackCurrentDate",
			time:      time.Date(2026, 6, 24, 2, 0, 0, 0, loc),
			resetTime: "bad",
			want:      "2026-06-24",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usecase.ComputeBusinessQueueDate(tt.time, tt.resetTime)
			assert.Equal(t, tt.want, got)
		})
	}
}
