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
