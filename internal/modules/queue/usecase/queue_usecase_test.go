package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/database"
	"github.com/Roisfaozi/go-clean-boilerplate/pkg/exception"
	"github.com/stretchr/testify/assert"
)

type stubQueueRepo struct {
	q       *entity.Queue
	queues  []*entity.Queue
	j       *entity.QueueJourney
	visit   *entity.VisitJourney
	err     error
	seenNum int
	exists  bool
}

func (s *stubQueueRepo) NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error) {
	s.seenNum++
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
	return s.err
}

func (s *stubQueueRepo) ListQueues(ctx context.Context, tenantID, branchID string) ([]*entity.Queue, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.queues, nil
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
	uc := NewQueueUseCase(repo, nil)
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

func TestRegisterQueue_DuplicateReturnsConflict(t *testing.T) {
	repo := &stubQueueRepo{exists: true}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	_, err := uc.RegisterQueue(ctx, &model.RegisterQueueRequest{ServiceID: "s-1", PatientName: "John Doe"})
	assert.ErrorIs(t, err, exception.ErrConflict)
}

func TestRegisterQueue_NoTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo, nil)

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
	uc := NewQueueUseCase(repo, nil)
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
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.NoError(t, err)
	assert.Equal(t, "q-1", res.ID)
	assert.NotEqual(t, "j-1", res.CurrentJourneyID)
}

func TestForwardQueue_Negative_NoTenant(t *testing.T) {
	uc := NewQueueUseCase(&stubQueueRepo{}, nil)
	_, err := uc.ForwardQueue(context.Background(), "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestForwardQueue_Edge_SameServiceStillCreatesJourney(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, res.CurrentJourneyID)
}

func TestForwardQueue_Security_CrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.Error(t, err)
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
	uc := NewQueueUseCase(repo, nil)
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
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCall})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestTransitionQueue_EdgeServingFromCalling(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusCalling},
	}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionServe})
	assert.NoError(t, err)
	assert.Equal(t, entity.QueueStatusServing, res.Status)
}

func TestListQueues_Success(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	ctx = database.SetBranchContext(ctx, "b-1")

	res, err := uc.ListQueues(ctx)
	assert.NoError(t, err)
	assert.Len(t, res, 1)
	assert.Equal(t, "q-1", res[0].ID)
}

func TestListQueues_NegativeMissingTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{queues: []*entity.Queue{{ID: "q-1"}}}
	uc := NewQueueUseCase(repo, nil)

	_, err := uc.ListQueues(context.Background())
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestGetQueueByID_Success(t *testing.T) {
	repo := &stubQueueRepo{q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting}}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.GetQueueByID(ctx, "q-1")
	assert.NoError(t, err)
	assert.Equal(t, "q-1", res.ID)
}

func TestGetQueueByID_NegativeCrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.GetQueueByID(ctx, "q-1")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestTransitionQueue_SecurityCrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCancel})
	assert.ErrorIs(t, err, exception.ErrNotFound)
}

func TestTransitionQueue_NegativeEmptyAction(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: ""})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestTransitionQueue_EdgeCancelTerminalStateRejected(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", Status: entity.QueueStatusCanceled, CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", Status: entity.JourneyStatusCanceled},
	}
	uc := NewQueueUseCase(repo, nil)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	_, err := uc.TransitionQueue(ctx, "q-1", &model.QueueTransitionRequest{Action: model.QueueActionCancel})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}
