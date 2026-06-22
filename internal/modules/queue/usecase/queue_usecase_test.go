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
	j       *entity.QueueJourney
	err     error
	seenNum int
}

func (s *stubQueueRepo) NextQueueNumber(ctx context.Context, tenantID, branchID string, date time.Time, prefix string) (int, error) {
	s.seenNum++
	if s.err != nil {
		return 0, s.err
	}
	return s.seenNum, nil
}
func (s *stubQueueRepo) CreateRegistration(ctx context.Context, queue *entity.Queue, journey *entity.QueueJourney, visit *entity.VisitJourney) error {
	s.q = queue
	s.j = journey
	return s.err
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
	return s.err
}

func TestRegisterQueue_Success(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo)
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

func TestRegisterQueue_NoTenantOrBranch(t *testing.T) {
	repo := &stubQueueRepo{}
	uc := NewQueueUseCase(repo)

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
	uc := NewQueueUseCase(repo)
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
	uc := NewQueueUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.NoError(t, err)
	assert.Equal(t, "q-1", res.ID)
	assert.NotEqual(t, "j-1", res.CurrentJourneyID)
}

func TestForwardQueue_Negative_NoTenant(t *testing.T) {
	uc := NewQueueUseCase(&stubQueueRepo{})
	_, err := uc.ForwardQueue(context.Background(), "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.ErrorIs(t, err, exception.ErrBadRequest)
}

func TestForwardQueue_Edge_SameServiceStillCreatesJourney(t *testing.T) {
	repo := &stubQueueRepo{
		q: &entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", CurrentJourneyID: "j-1"},
		j: &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending},
	}
	uc := NewQueueUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	res, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-1"})
	assert.NoError(t, err)
	assert.NotEmpty(t, res.CurrentJourneyID)
}

func TestForwardQueue_Security_CrossTenantRejected(t *testing.T) {
	repo := &stubQueueRepo{err: exception.ErrNotFound}
	uc := NewQueueUseCase(repo)
	ctx := database.SetOrganizationContext(context.Background(), "other-tenant")

	_, err := uc.ForwardQueue(ctx, "q-1", &model.ForwardQueueRequest{DestinationServiceID: "s-2"})
	assert.Error(t, err)
}
