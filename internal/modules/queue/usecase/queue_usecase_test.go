package usecase

import (
	"context"
	"fmt"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Stubs
// =============================================================================

type stubQueueRepo struct {
	FindQueueByTenantIDFunc func(ctx context.Context, tenantID, queueID string) (*entity.Queue, error)
	q                       *entity.Queue
	queues                  []*entity.Queue
	j                       *entity.QueueJourney
	visit                   *entity.VisitJourney
	err                     error
	seenNum                 int
	exists                  bool
	listReq                 model.ListQueuesRequest
	journeyReq              model.QueueJourneyListRequest
	journeyTenantID         string
	journeyBranchID         string
	journeyList             []*entity.QueueJourney
	lastPrefix              string
	visits                  []*entity.VisitJourney
	statsRes                model.QueueStatsResponse
}

type stubSettingsResolver struct {
	values map[string]string
	Calls  []string
}

func (s *stubSettingsResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	s.Calls = append(s.Calls, key)
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", exception.ErrNotFound
}

type stubRelationValidator struct {
	err error
}

type stubAuditLogger struct {
	entries []auditModel.CreateAuditLogRequest
	err     error
}

func (s *stubAuditLogger) LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error {
	s.entries = append(s.entries, req)
	return s.err
}

func (s *stubRelationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	return s.err
}

func TestQueueAuditLogging(t *testing.T) {
	t.Run("Register_EmitsAuditAndSurvivesAuditFailure", func(t *testing.T) {
		repo := &stubQueueRepo{}
		audit := &stubAuditLogger{err: assert.AnError}
		uc := NewQueueUseCase(repo, &stubSettingsResolver{}, nil, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		ctx = authcontext.WithUserID(ctx, "u-1")

		res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John"})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, audit.entries, 1)
		values, ok := audit.entries[0].NewValues.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "QUEUE_REGISTER", audit.entries[0].Action)
		assert.Equal(t, "queue", audit.entries[0].Entity)
		assert.Equal(t, "t-1", audit.entries[0].OrganizationID)
		assert.Equal(t, "u-1", audit.entries[0].UserID)
		assert.Equal(t, "b-1", values["branch_id"])
		assert.Equal(t, res.TicketNo, values["ticket_no"])
	})

	t.Run("Forward_EmitsAudit", func(t *testing.T) {
		repo := &stubQueueRepo{
			q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
			j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
		}
		audit := &stubAuditLogger{}
		uc := NewQueueUseCase(repo, nil, &stubRelationValidator{}, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "svc-2", DestinationCounterID: "ctr-2"})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, audit.entries, 1)
		values, ok := audit.entries[0].NewValues.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "QUEUE_FORWARD", audit.entries[0].Action)
		assert.Equal(t, "system", audit.entries[0].UserID)
		assert.Equal(t, "q-1", audit.entries[0].EntityID)
		assert.Equal(t, "j-1", values["from_journey_id"])
		assert.Equal(t, "svc-2", values["to_service_id"])
	})

	t.Run("Transition_EmitsAudit", func(t *testing.T) {
		repo := &stubQueueRepo{
			q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
			j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
		}
		audit := &stubAuditLogger{}
		uc := NewQueueUseCase(repo, nil, nil, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCall})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, audit.entries, 1)
		values, ok := audit.entries[0].NewValues.(map[string]string)
		require.True(t, ok)
		assert.Equal(t, "QUEUE_CALL", audit.entries[0].Action)
		assert.Equal(t, "j-1", values["journey_id"])
		assert.Equal(t, entity.QueueStatusCalling, values["status"])
	})

	t.Run("Register_AuditFailureDoesNotFailBusinessFlow", func(t *testing.T) {
		repo := &stubQueueRepo{}
		audit := &stubAuditLogger{err: assert.AnError}
		uc := NewQueueUseCase(repo, &stubSettingsResolver{}, nil, audit)

		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John"})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Len(t, audit.entries, 1)
	})
}

func (s *stubQueueRepo) NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error) {
	s.seenNum++
	s.lastPrefix = prefix
	if s.err != nil {
		return 0, s.err
	}
	return s.seenNum, nil
}
func (s *stubQueueRepo) ExistsRegistration(ctx context.Context, tenantID, branchID, queueDate, patientID, patientName string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.exists, nil
}
func (s *stubQueueRepo) CreateRegistration(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error {
	s.q = queue
	s.j = journey
	s.visit = visit
	return s.err
}

func (s *stubQueueRepo) CreateRegistrationWithNumber(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney, date time.Time, prefix string) error {
	s.seenNum++
	s.lastPrefix = prefix
	queue.QueueNo = s.seenNum
	queue.TicketNo = fmt.Sprintf("%s%03d", prefix, queue.QueueNo)
	return s.CreateRegistration(ctx, queue, journey, visit)
}

func (s *stubQueueRepo) ListQueues(ctx context.Context, tenantID, branchID string, req model.ListQueuesRequest) ([]*entity.Queue, error) {
	s.listReq = req
	if s.err != nil {
		return nil, s.err
	}
	if req.Status == "" && req.QueueDate == "" && req.ServiceID == "" {
		return s.queues, nil
	}
	filtered := make([]*entity.Queue, 0)
	for _, queue := range s.queues {
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
}

func (s *stubQueueRepo) ListActiveJourneys(ctx context.Context, tenantID, branchID string, req model.QueueJourneyListRequest) ([]*entity.QueueJourney, error) {
	if s.err != nil {
		return nil, s.err
	}
	s.journeyReq = req
	s.journeyTenantID = tenantID
	s.journeyBranchID = branchID
	if s.journeyList != nil {
		return s.journeyList, nil
	}
	return []*entity.QueueJourney{{ID: "j-1", QueueID: "q-1", TenantID: tenantID, ServiceID: req.ServiceID, CounterID: req.CounterID, SeqNo: 1, Status: entity.JourneyStatusCalling}}, nil
}

func (s *stubQueueRepo) FindVisitJourneysByQueueID(ctx context.Context, tenantID, branchID, queueID string) ([]*entity.VisitJourney, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.visits, nil
}

func (s *stubQueueRepo) GetQueueStats(ctx context.Context, tenantID, branchID, queueDate string) (model.QueueStatsResponse, error) {
	if s.err != nil {
		return s.statsRes, s.err
	}
	return s.statsRes, nil
}

func (s *stubQueueRepo) FindQueueByTenantID(ctx context.Context, tenantID, queueID string) (*entity.Queue, error) {
	if s.FindQueueByTenantIDFunc != nil {
		return s.FindQueueByTenantIDFunc(ctx, tenantID, queueID)
	}
	return nil, nil
}

func (s *stubQueueRepo) FindQueueByID(ctx context.Context, tenantID, branchID, queueID string) (*entity.Queue, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.q, nil
}

func (s *stubQueueRepo) FindCurrentJourney(ctx context.Context, tenantID, branchID, queueID, journeyID string) (*entity.QueueJourney, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.j, nil
}

func (s *stubQueueRepo) NextJourneySequence(ctx context.Context, tenantID, branchID, queueID string) (int, error) {
	if s.err != nil {
		return 0, s.err
	}
	if s.j == nil {
		return 1, nil
	}
	return s.j.SeqNo + 1, nil
}

func (s *stubQueueRepo) CreateForwarding(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, nextJourney *entity.QueueJourney, visit *entity.VisitJourney) error {
	s.q = queue
	s.j = nextJourney
	s.visit = visit
	return s.err
}

func (s *stubQueueRepo) UpdateQueueState(ctx context.Context, queue *entity.Queue, currentJourney *entity.QueueJourney, visit *entity.VisitJourney) error {
	s.q = queue
	s.j = currentJourney
	s.visit = visit
	return s.err
}

// =============================================================================
// TestRegisterQueue
// =============================================================================

func TestRegisterQueue(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      *model.RegisterQueueRequest
		repo     *stubQueueRepo
		settings map[string]string
		tenantID string
		branchID string
		wantErr  error
		wantRes  func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John Doe"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
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
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Equal(t, "queue_reset_time", resolver.Calls[0])
			},
		},
		{
			name:     "Negative_IgnoresLegacyResetTimeKeyInUsecase",
			category: "negative",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"reset_time": "05:00"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Equal(t, "queue_reset_time", resolver.Calls[0])
			},
		},
		{
			name:     "Positive_UsesTicketPrefixSettingFirst",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"ticket_prefix": "JS"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Equal(t, "JS", repo.lastPrefix)
				assert.Equal(t, "JS001", res.TicketNo)
				assert.Contains(t, resolver.Calls, "ticket_prefix")
			},
		},
		{
			name:     "Positive_FallbacksToLegacyPrefixKey",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"prefix": "RX"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Equal(t, "RX", repo.lastPrefix)
				assert.Equal(t, "RX001", res.TicketNo)
				assert.Contains(t, resolver.Calls, "ticket_prefix")
				assert.Contains(t, resolver.Calls, "prefix")
			},
		},
		{
			name:     "Positive_DefaultsPrefixToA",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Equal(t, "A", repo.lastPrefix)
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
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Equal(t, 1, res.QueueNo)
				assert.Contains(t, resolver.Calls, "numbering_strategy")
			},
		},
		{
			name:     "Positive_FallbacksToLegacyNumberingKey",
			category: "positive",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"numbering": "sequential"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.Contains(t, resolver.Calls, "numbering_strategy")
				assert.Contains(t, resolver.Calls, "numbering")
			},
		},
		{
			name:     "Edge_InvalidNumberingStrategyFallsBackToSequential",
			category: "edge",
			req:      &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"},
			settings: map[string]string{"numbering_strategy": "random"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
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
			repo:     &stubQueueRepo{exists: true},
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
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
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
			wantRes: func(t *testing.T, repo *stubQueueRepo, resolver *stubSettingsResolver, res *model.QueueResponse) {
				assert.NotEmpty(t, repo.q)
				assert.NotEmpty(t, repo.j)
				assert.NotEmpty(t, repo.visit)
				assert.Equal(t, "P", repo.lastPrefix)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			resolver := &stubSettingsResolver{values: tt.settings}
			uc := NewQueueUseCase(repo, resolver, nil)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
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
				tt.wantRes(t, repo, resolver, res)
			}
		})
	}
}

// =============================================================================
// TestForwardQueue
// =============================================================================

func TestForwardQueue(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		repo      *stubQueueRepo
		validator *stubRelationValidator
		queueID   string
		req       *model.ForwardQueueRequest
		tenantID  string
		branchID  string
		wantErr   error
		wantRes   func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", TicketNo: "A001", QueueNo: 1, Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-2"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
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
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-1"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.NotEmpty(t, res.CurrentJourneyID)
			},
		},
		{
			name:     "Edge_MissingActiveJourneyRejected",
			category: "edge",
			repo: &stubQueueRepo{
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
			repo:     &stubQueueRepo{err: exception.ErrNotFound},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "s-2"},
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Negative_InvalidDestinationRelationRejected",
			category: "negative",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			validator: &stubRelationValidator{err: exception.ErrForbidden},
			queueID:   "q-1",
			req:       &model.ForwardQueueRequest{DestinationServiceID: "s-2", DestinationCounterID: "c-2"},
			tenantID:  "t-1",
			branchID:  "b-1",
			wantErr:   exception.ErrForbidden,
		},
		{
			name:     "Positive_PharmacyFlowKeepsSingleQueueRecord",
			category: "positive",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.ForwardQueueRequest{DestinationServiceID: "pharmacy-svc"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, "q-1", res.ID)
				assert.NotEmpty(t, repo.q)
				assert.NotEqual(t, "j-1", repo.j.ID)
				assert.NotEmpty(t, repo.visit)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			var val RelationValidator
			if tt.validator != nil {
				val = tt.validator
			}
			uc := NewQueueUseCase(repo, nil, val)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
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
				tt.wantRes(t, repo, res)
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
		repo     *stubQueueRepo
		queueID  string
		req      *model.QueueTransitionRequest
		tenantID string
		branchID string
		wantErr  error
		wantRes  func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse)
	}{
		{
			name:     "Positive_SuccessCalling",
			category: "positive",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCall},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCalling, res.Status)
				assert.Equal(t, entity.JourneyStatusCalling, repo.j.Status)
				assert.Equal(t, "call", repo.visit.EventType)
			},
		},
		{
			name:     "Positive_SuccessCancel",
			category: "positive",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCanceled, res.Status)
			},
		},
		{
			name:     "Positive_SuccessSkip",
			category: "positive",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionSkip},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusSkipped, res.Status)
			},
		},
		{
			name:     "Edge_ServingFromCalling",
			category: "edge",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionServe},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusServing, res.Status)
			},
		},
		{
			name:     "Edge_SkipFromCalling",
			category: "edge",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusCalling},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionSkip},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusSkipped, res.Status)
				assert.Equal(t, entity.JourneyStatusSkipped, repo.j.Status)
				assert.Equal(t, "skip", repo.visit.EventType)
			},
		},
		{
			name:     "Edge_CancelFromWaiting",
			category: "edge",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusPending},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCanceled, res.Status)
				assert.Equal(t, entity.JourneyStatusCanceled, repo.j.Status)
				assert.Equal(t, "cancel", repo.visit.EventType)
			},
		},
		{
			name:     "Edge_CompleteAfterServing",
			category: "edge",
			repo: &stubQueueRepo{
				q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusServing, CurrentJourneyID: "j-1"},
				j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.JourneyStatusServing},
			},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionComplete},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res *model.QueueResponse) {
				assert.Equal(t, entity.QueueStatusCompleted, res.Status)
			},
		},
		{
			name:     "Negative_InvalidStateChange",
			category: "negative",
			repo: &stubQueueRepo{
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
			repo: &stubQueueRepo{
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
			name:     "Negative_CompleteFromWaitingRejected",
			category: "negative",
			repo: &stubQueueRepo{
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
			repo: &stubQueueRepo{
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
			repo: &stubQueueRepo{
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
			repo:     &stubQueueRepo{err: exception.ErrNotFound},
			queueID:  "q-1",
			req:      &model.QueueTransitionRequest{Action: model.QueueActionCancel},
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			uc := NewQueueUseCase(repo, nil, nil)
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
				tt.wantRes(t, repo, res)
			}
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
		repo     *stubQueueRepo
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
			repo:     &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}},
			req:      model.ListQueuesRequest{},
			tenantID: "t-1",
			branchID: "b-1",
			wantLen:  1,
			wantID:   "q-1",
		},
		{
			name:     "Negative_MissingTenantOrBranch",
			category: "negative",
			repo:     &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1"}}},
			req:      model.ListQueuesRequest{},
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Edge_StatusFilter",
			category: "edge",
			repo:     &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}, {ID: "q-2", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCompleted}}},
			req:      model.ListQueuesRequest{Status: entity.QueueStatusCompleted},
			tenantID: "t-1",
			branchID: "b-1",
			wantLen:  1,
			wantID:   "q-2",
		},
		{
			name:     "Positive_QueueDateFilter",
			category: "positive",
			repo:     &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting}, {ID: "q-2", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-23", Status: entity.QueueStatusWaiting}}},
			req:      model.ListQueuesRequest{QueueDate: "2026-06-24"},
			tenantID: "t-1",
			branchID: "b-1",
			wantLen:  1,
			wantID:   "q-1",
		},
		{
			name:     "Vulnerability_PassesScopedFiltersOnly",
			category: "vulnerability",
			repo:     &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", CurrentJourneyID: "svc-1"}}},
			req:      model.ListQueuesRequest{Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", ServiceID: "svc-1"},
			tenantID: "tenant-safe",
			branchID: "branch-safe",
			wantLen:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			uc := NewQueueUseCase(repo, nil, nil)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
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
				assert.Equal(t, tt.req, repo.listReq)
			}
		})
	}
}

// =============================================================================
// TestListActiveJourneys
// =============================================================================

func TestListActiveJourneys(t *testing.T) {
	tests := []struct {
		name      string
		category  string
		repo      *stubQueueRepo
		validator *stubRelationValidator
		req       model.QueueJourneyListRequest
		tenantID  string
		branchID  string
		wantErr   error
		wantRes   func(t *testing.T, repo *stubQueueRepo, res []model.QueueJourneyResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res []model.QueueJourneyResponse) {
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
			name:      "Negative_InvalidServiceRelationRejected",
			category:  "negative",
			validator: &stubRelationValidator{err: exception.ErrForbidden},
			req:       model.QueueJourneyListRequest{ServiceID: "svc-foreign"},
			tenantID:  "t-1",
			branchID:  "b-1",
			wantErr:   exception.ErrForbidden,
		},
		{
			name:     "Edge_CounterFilter",
			category: "edge",
			req:      model.QueueJourneyListRequest{CounterID: "counter-1"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res []model.QueueJourneyResponse) {
				require.Len(t, res, 1)
				assert.Equal(t, "counter-1", res[0].CounterID)
			},
		},
		{
			name:     "Edge_EmptyList",
			category: "edge",
			repo:     &stubQueueRepo{journeyList: []*entity.QueueJourney{}},
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1"},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res []model.QueueJourneyResponse) {
				assert.Empty(t, res)
			},
		},
		{
			name:      "Negative_InvalidCounterRelationRejected",
			category:  "negative",
			validator: &stubRelationValidator{err: exception.ErrForbidden},
			req:       model.QueueJourneyListRequest{CounterID: "counter-foreign"},
			tenantID:  "t-1",
			branchID:  "b-1",
			wantErr:   exception.ErrForbidden,
		},
		{
			name:     "Vulnerability_ScopedTenantPassThrough",
			category: "vulnerability",
			req:      model.QueueJourneyListRequest{ServiceID: "svc-1"},
			tenantID: "tenant-safe",
			branchID: "branch-safe",
			wantRes: func(t *testing.T, repo *stubQueueRepo, res []model.QueueJourneyResponse) {
				require.Len(t, res, 1)
				assert.Equal(t, "tenant-safe", repo.journeyTenantID)
				assert.Equal(t, "branch-safe", repo.journeyBranchID)
				assert.Equal(t, model.QueueJourneyListRequest{ServiceID: "svc-1"}, repo.journeyReq)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			var val RelationValidator
			if tt.validator != nil {
				val = tt.validator
			}
			uc := NewQueueUseCase(repo, nil, val)

			ctx := context.Background()
			if tt.tenantID != "" {
				ctx = database.SetOrganizationContext(ctx, tt.tenantID)
			}
			if tt.branchID != "" {
				ctx = database.SetBranchContext(ctx, tt.branchID)
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
				tt.wantRes(t, repo, res)
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
		repo     *stubQueueRepo
		queueID  string
		tenantID string
		branchID string
		wantErr  error
		wantRes  func(t *testing.T, res []model.VisitJourneyResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repo: &stubQueueRepo{
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
			repo:     &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1"}},
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
			repo:     &stubQueueRepo{err: exception.ErrNotFound, visits: []*entity.VisitJourney{{ID: "v-1"}}},
			queueID:  "q-1",
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Vulnerability_CrossTenantRejected",
			category: "vulnerability",
			repo:     &stubQueueRepo{err: exception.ErrNotFound},
			queueID:  "q-1",
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
		{
			name:     "Negative_QueueLookupFailureStopsVisitsRead",
			category: "negative",
			repo:     &stubQueueRepo{err: exception.ErrNotFound, visits: []*entity.VisitJourney{{ID: "v-1"}}},
			queueID:  "q-1",
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			uc := NewQueueUseCase(repo, nil, nil)
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
		repo     *stubQueueRepo
		tenantID string
		branchID string
		wantErr  error
		wantID   string
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repo:     &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}},
			tenantID: "t-1",
			branchID: "b-1",
			wantID:   "q-1",
		},
		{
			name:     "Negative_MissingBranch",
			category: "negative",
			repo:     &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}},
			tenantID: "t-1",
			wantErr:  exception.ErrBadRequest,
		},
		{
			name:     "Negative_EmptyQueueID",
			category: "negative",
			repo:     &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}},
			tenantID: "t-1",
			branchID: "b-1",
			wantErr:  exception.ErrBadRequest,
			wantID:   "",
		},
		{
			name:     "Vulnerability_CrossTenantRejected",
			category: "vulnerability",
			repo:     &stubQueueRepo{err: exception.ErrNotFound},
			tenantID: "other-tenant",
			branchID: "b-1",
			wantErr:  exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			uc := NewQueueUseCase(repo, nil, nil)
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
		name      string
		category  string
		repo      *stubQueueRepo
		validator *stubRelationValidator
		tenantID  string
		branchID  string
		wantErr   error
		wantRes   func(t *testing.T, res *model.QueueStatsResponse)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			repo:     &stubQueueRepo{statsRes: model.QueueStatsResponse{TotalQueuesToday: 5, TotalActiveJourneys: 3}},
			tenantID: "t-1",
			branchID: "b-1",
			wantRes: func(t *testing.T, res *model.QueueStatsResponse) {
				assert.Equal(t, int64(5), res.TotalQueuesToday)
				assert.Equal(t, int64(3), res.TotalActiveJourneys)
			},
		},
		{
			name:      "Negative_InvalidBranchRelationRejected",
			category:  "negative",
			validator: &stubRelationValidator{err: exception.ErrForbidden},
			tenantID:  "t-1",
			branchID:  "branch-foreign",
			wantErr:   exception.ErrForbidden,
		},
		{
			name:     "Edge_ZeroStatsAllowed",
			category: "edge",
			repo:     &stubQueueRepo{statsRes: model.QueueStatsResponse{}},
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
			repo := tt.repo
			if repo == nil {
				repo = &stubQueueRepo{}
			}
			var val RelationValidator
			if tt.validator != nil {
				val = tt.validator
			}
			uc := NewQueueUseCase(repo, nil, val)

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
			got := computeBusinessQueueDate(tt.time, tt.resetTime)
			assert.Equal(t, tt.want, got)
		})
	}
}
