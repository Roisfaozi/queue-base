package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubQueueRepo struct {
	q               *entity.Queue
	queues          []*entity.Queue
	j               *entity.QueueJourney
	visit           *entity.VisitJourney
	err             error
	seenNum         int
	exists          bool
	listReq         model.ListQueuesRequest
	journeyReq      model.QueueJourneyListRequest
	journeyTenantID string
	journeyBranchID string
	lastPrefix      string
	visits          []*entity.VisitJourney
	statsRes        model.QueueStatsResponse
}

type stubSettingsResolver struct {
	values map[string]string
	calls  []string
}

func (s *stubSettingsResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	s.calls = append(s.calls, key)
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", exception.ErrNotFound
}

type stubRelationValidator struct {
	err error
}

func (s *stubRelationValidator) Validate(ctx context.Context, tenantID, branchID, serviceID, counterID string) error {
	return s.err
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
	return []*entity.QueueJourney{{ID: "j-1", QueueID: "q-1", TenantID: tenantID, ServiceID: req.ServiceID, CounterID: req.CounterID, SeqNo: 1, Status: entity.JourneyStatusCalling}}, nil
}

func (s *stubQueueRepo) FindVisitJourneysByQueueID(ctx context.Context, tenantID, queueID string) ([]*entity.VisitJourney, error) {
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

func (s *stubQueueRepo) FindQueueByID(ctx context.Context, tenantID, queueID string) (*entity.Queue, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.q, nil
}

func (s *stubQueueRepo) FindCurrentJourney(ctx context.Context, queueID, journeyID string) (*entity.QueueJourney, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.j, nil
}

func (s *stubQueueRepo) NextJourneySequence(ctx context.Context, queueID string) (int, error) {
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

func TestRegisterQueue_Success(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	req := &model.RegisterQueueRequest{
		ServiceID:   "s-1",
		PatientName: "John Doe",
	}

	res, err := uc.RegisterQueue(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "t-1", res.TenantID)
	assert.Equal(t, "b-1", res.BranchID)
	assert.Equal(t, 1, res.QueueNo)
	assert.Equal(t, entity.QueueStatusWaiting, res.Status)
}

func TestRegisterQueue_UsesQueueResetTimeKeyFirst(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"queue_reset_time": "05:00"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, "queue_reset_time", resolver.calls[0])
}

func TestRegisterQueue_FallbacksToLegacyResetTimeKey(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"reset_time": "05:00"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, []string{"queue_reset_time", "reset_time"}, resolver.calls[:2])
}

func TestRegisterQueue_UsesTicketPrefixSettingFirst(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"ticket_prefix": "JS"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, "JS", repo.lastPrefix)
	assert.Equal(t, "JS001", res.TicketNo)
	assert.Contains(t, resolver.calls, "ticket_prefix")
}

func TestRegisterQueue_FallbacksToLegacyPrefixKey(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"prefix": "RX"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, "RX", repo.lastPrefix)
	assert.Equal(t, "RX001", res.TicketNo)
	assert.Contains(t, resolver.calls, "ticket_prefix")
	assert.Contains(t, resolver.calls, "prefix")
}

func TestRegisterQueue_DefaultsPrefixToA(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, "A", repo.lastPrefix)
	assert.Equal(t, "A001", res.TicketNo)
}

func TestRegisterQueue_UsesNumberingStrategySettingFirst(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"numbering_strategy": "sequential"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, 1, res.QueueNo)
	assert.Contains(t, resolver.calls, "numbering_strategy")
}

func TestRegisterQueue_FallbacksToLegacyNumberingKey(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"numbering": "sequential"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Contains(t, resolver.calls, "numbering_strategy")
	assert.Contains(t, resolver.calls, "numbering")
}

func TestRegisterQueue_EdgeInvalidNumberingStrategyFallsBackToSequential(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"numbering_strategy": "random"}}
	uc := NewQueueUseCase(repo, resolver, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	require.NoError(t, err)
	assert.Equal(t, 1, res.QueueNo)
	assert.Equal(t, "A001", res.TicketNo)
}

func TestRegisterQueue_SecurityNumberingStrategyDoesNotBypassTenantBranchScope(t *testing.T) {
	repo := &stubQueueRepo{}
	resolver := &stubSettingsResolver{values: map[string]string{"numbering_strategy": "sequential"}}
	uc := NewQueueUseCase(repo, resolver, nil)

	_, err := uc.RegisterQueue(context.Background(), &model.RegisterQueueRequest{ServiceID: "svc-1", PatientName: "John Doe"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRegisterQueue_DuplicateReturnsConflict(t *testing.T) {
	repo := &stubQueueRepo{exists: true}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John Doe"})
	assert.ErrorIs(t, err, exception.ErrConflict)
}

func TestRegisterQueue_NoTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)

	// Test no tenant
	ctx := database.SetBranchContext(context.Background(), "b-1")
	req := &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John"}
	_, err := uc.RegisterQueue(ctx, req)
	assert.ErrorIs(t, err, exception.ErrBadRequest)

	// Test no branch
	ctx = database.SetOrganizationContext(context.Background(), "t-1")
	_, err = uc.RegisterQueue(ctx, req)
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestRegisterQueue_Security_VulnerabilitySQLInjection(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	// Attempting sql injection payloads in strings
	req := &model.RegisterQueueRequest{
		ServiceID:   "s-1",
		PatientName: "John'; DROP TABLE queues;--",
	}

	res, err := uc.RegisterQueue(ctx, req)
	assert.NoError(t, err)
	// GORM safely binds the parameters so raw payload acts purely as text
	assert.Equal(t, "John&#39;; DROP TABLE queues;--", res.PatientName)
}

func TestForwardQueue_Success(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{
			ID:               "q-1",
			TenantID:         "t-1",
			BranchID:         "b-1",
			TicketNo:         "A001",
			QueueNo:          1,
			Status:           entity.QueueStatusWaiting,
			CurrentJourneyID: "j-1",
		},
		j: &entity.QueueJourney{
			ID:        "j-1",
			QueueID:   "q-1",
			TenantID:  "t-1",
			ServiceID: "s-1",
			SeqNo:     1,
			Status:    entity.JourneyStatusPending,
		},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.NoError(t, err)
	assert.Equal(t, "q-1", res.ID)
	assert.NotEqual(t, "j-1", res.CurrentJourneyID)
}

func TestForwardQueue_Negative_NoTenant(t *testing.T) {
	uc := NewQueueUseCase(&stubQueueRepo{}, nil, nil)
	_, err := uc.ForwardQueue(context.Background(), "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestForwardQueue_Edge_SameServiceStillCreatesJourney(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, res.CurrentJourneyID)
}

func TestForwardQueue_Security_CrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.Error(t, err)
}

func TestForwardQueue_NegativeInvalidDestinationRelationRejected(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusPending},
	}
	validator := &stubRelationValidator{err: exception.ErrForbidden}
	uc := NewQueueUseCase(repo, nil, validator)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2", DestinationCounterID: "c-2"})
	assert.ErrorIs(t, err, exception.ErrForbidden)
}

func TestForwardQueue_PharmacyFlowKeepsSingleQueueRecord(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "pharmacy-svc"})
	assert.NoError(t, err)
	assert.Equal(t, "q-1", res.ID)
	assert.NotEmpty(t, repo.q)
	assert.NotEqual(t, "j-1", repo.j.ID)
	assert.NotEmpty(t, repo.visit)
}

func TestRegisterQueue_PharmacyServiceStillCreatesSingleMasterQueue(t *testing.T) {
	repo := &stubQueueRepo{}
	settings := &stubSettingsResolver{values: map[string]string{
		"queue_reset_time": "04:00",
		"ticket_prefix":    "P",
	}}
	uc := NewQueueUseCase(repo, settings, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	req := &model.RegisterQueueRequest{ServiceID: "pharmacy-svc", PatientName: "John Doe"}

	res, err := uc.RegisterQueue(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.NotEmpty(t, repo.q)
	assert.NotEmpty(t, repo.j)
	assert.NotEmpty(t, repo.visit)
	assert.Equal(t, "P", repo.lastPrefix)
}

func TestComputeBusinessQueueDate_DefaultResetTime(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	assert.NoError(t, err)

	t.Run("Edge_BeforeResetUsesPreviousDate", func(t *testing.T) {
		now := time.Date(2026, 6, 24, 3, 59, 0, 0, loc)
		got := computeBusinessQueueDate(now, "04:00")
		assert.Equal(t, "2026-06-23", got)
	})

	t.Run("Positive_AfterResetUsesCurrentDate", func(t *testing.T) {
		now := time.Date(2026, 6, 24, 4, 1, 0, 0, loc)
		got := computeBusinessQueueDate(now, "04:00")
		assert.Equal(t, "2026-06-24", got)
	})

	t.Run("Negative_InvalidResetFallsBackCurrentDate", func(t *testing.T) {
		now := time.Date(2026, 6, 24, 2, 0, 0, 0, loc)
		got := computeBusinessQueueDate(now, "bad")
		assert.Equal(t, "2026-06-24", got)
	})
}

func TestTransitionQueue_SuccessCalling(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCall})
	assert.NoError(t, err)
	assert.Equal(t, entity.QueueStatusCalling, res.Status)
	assert.Equal(t, entity.JourneyStatusCalling, repo.j.Status)
	assert.Equal(t, "call", repo.visit.EventType)
}

func TestTransitionQueue_NegativeInvalidStateChange(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCompleted, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusCompleted},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCall})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestTransitionQueue_EdgeServingFromCalling(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusCalling},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionServe})
	assert.NoError(t, err)
	assert.Equal(t, entity.QueueStatusServing, res.Status)
}

func TestListQueues_Success(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ListQueues(ctx, model.ListQueuesRequest{})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "q-1", res[0].ID)
}

func TestListQueues_NegativeMissingTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1"}}}
	uc := NewQueueUseCase(repo, nil, nil)

	_, err := uc.ListQueues(context.Background(), model.ListQueuesRequest{})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestListQueues_EdgeStatusFilter(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}, {ID: "q-2", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCompleted}}}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ListQueues(ctx, model.ListQueuesRequest{Status: entity.QueueStatusCompleted})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "q-2", res[0].ID)
}

func TestListQueues_PositiveQueueDateFilter(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting}, {ID: "q-2", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-23", Status: entity.QueueStatusWaiting}}}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ListQueues(ctx, model.ListQueuesRequest{QueueDate: "2026-06-24"})
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "q-1", res[0].ID)
}

func TestListQueues_VulnerabilityPassesScopedFiltersOnly(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-safe")
	ctx = database.SetBranchContext(ctx, "branch-safe")
	req := model.ListQueuesRequest{Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", ServiceID: "svc-1"}

	_, err := uc.ListQueues(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, req, repo.listReq)
}

func TestListActiveJourneys_Success(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ListActiveJourneys(ctx, model.QueueJourneyListRequest{ServiceID: "svc-1", QueueDate: "2026-06-24"})
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, "svc-1", res[0].ServiceID)
	assert.Equal(t, entity.JourneyStatusCalling, res[0].Status)
}

func TestListActiveJourneys_NegativeMissingTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)

	_, err := uc.ListActiveJourneys(context.Background(), model.QueueJourneyListRequest{ServiceID: "svc-1"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestListActiveJourneys_EdgeCounterFilter(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ListActiveJourneys(ctx, model.QueueJourneyListRequest{CounterID: "counter-1"})
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, "counter-1", res[0].CounterID)
}

func TestListActiveJourneys_SecurityScopedTenantPassThrough(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-safe")
	ctx = database.SetBranchContext(ctx, "branch-safe")

	res, err := uc.ListActiveJourneys(ctx, model.QueueJourneyListRequest{ServiceID: "svc-1"})
	assert.NoError(t, err)
	require.Len(t, res, 1)
	assert.Equal(t, "tenant-safe", repo.journeyTenantID)
	assert.Equal(t, "branch-safe", repo.journeyBranchID)
	assert.Equal(t, model.QueueJourneyListRequest{ServiceID: "svc-1"}, repo.journeyReq)
}

func TestGetQueueByID_Success(t *testing.T) {
	repo := &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.GetQueueByID(ctx, "q-1")
	assert.NoError(t, err)
	assert.Equal(t, "q-1", res.ID)
}

func TestGetVisitJourneys_Success(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1"},
		visits: []*entity.VisitJourney{
			{ID: "v-1", QueueID: "q-1", TenantID: "t-1", EventType: "registration", CreatedAt: 100},
			{ID: "v-2", QueueID: "q-1", TenantID: "t-1", EventType: "call", CreatedAt: 200},
		},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.GetVisitJourneys(ctx, "q-1")
	assert.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "v-1", res[0].ID)
	assert.Equal(t, "v-2", res[1].ID)
}

func TestGetVisitJourneys_NegativeMissingTenant(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)

	_, err := uc.GetVisitJourneys(context.Background(), "q-1")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestGetVisitJourneys_EdgeEmptyList(t *testing.T) {
	repo := &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1"}}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.GetVisitJourneys(ctx, "q-1")
	assert.NoError(t, err)
	assert.Empty(t, res)
}

func TestGetVisitJourneys_SecurityCrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.GetVisitJourneys(ctx, "q-1")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestGetQueueByID_NegativeCrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.GetQueueByID(ctx, "q-1")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestTransitionQueue_SecurityCrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCancel})
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestTransitionQueue_NegativeEmptyAction(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: ""})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestTransitionQueue_EdgeCancelTerminalStateRejected(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCanceled, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusCanceled},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCancel})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestGetQueueStats_Success(t *testing.T) {
	repo := &stubQueueRepo{
		statsRes: model.QueueStatsResponse{TotalQueuesToday: 5, TotalActiveJourneys: 3},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.GetQueueStats(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), res.TotalQueuesToday)
	assert.Equal(t, int64(3), res.TotalActiveJourneys)
}

func TestGetQueueStats_NegativeNoTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)

	_, err := uc.GetQueueStats(context.Background())
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestGetVisitJourneys_NegativeEmptyQueueID(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.GetVisitJourneys(ctx, "")
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestTransitionQueue_SuccessCancel(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusCalling},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCancel})
	assert.NoError(t, err)
	assert.Equal(t, entity.QueueStatusCanceled, res.Status)
}

func TestTransitionQueue_SuccessSkip(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionSkip})
	assert.NoError(t, err)
	assert.Equal(t, entity.QueueStatusSkipped, res.Status)
}

func TestTransitionQueue_EdgeCompleteAfterServing(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusServing, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusServing},
	}
	uc := NewQueueUseCase(repo, nil, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionComplete})
	assert.NoError(t, err)
	assert.Equal(t, entity.QueueStatusCompleted, res.Status)
}
