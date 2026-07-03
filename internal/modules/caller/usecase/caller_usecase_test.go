package usecase

import (
	"context"
	"testing"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubQueueUC struct {
	res     *queueModel.QueueResponse
	err     error
	ctx     context.Context
	queueID string
	action  string
}

func (s *stubQueueUC) RegisterQueue(ctx context.Context, req *queueModel.RegisterQueueRequest) (*queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) ListQueues(ctx context.Context, req queueModel.ListQueuesRequest) ([]queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) GetQueueByID(ctx context.Context, queueID string) (*queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) ForwardQueue(ctx context.Context, queueID string, req *queueModel.ForwardQueueRequest) (*queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) TransitionQueue(ctx context.Context, queueID string, req *queueModel.QueueTransitionRequest) (*queueModel.QueueResponse, error) {
	s.ctx = ctx
	s.queueID = queueID
	s.action = req.Action
	return s.res, s.err
}
func (s *stubQueueUC) ListActiveJourneys(ctx context.Context, req queueModel.QueueJourneyListRequest) ([]queueModel.QueueJourneyResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) GetVisitJourneys(ctx context.Context, queueID string) ([]queueModel.VisitJourneyResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) GetQueueStats(ctx context.Context) (*queueModel.QueueStatsResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) ResolveQueueBranchID(ctx context.Context, queueID string) (string, error) {
	return "", nil
}

func newCallerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE queue_journeys (
			id TEXT PRIMARY KEY,
			queue_id TEXT,
			tenant_id TEXT,
			branch_id TEXT,
			counter_id TEXT
		);
		CREATE TABLE qms_clients (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			branch_id TEXT,
			branch_service_id TEXT,
			counter_id TEXT,
			is_active BOOLEAN
		);
		CREATE TABLE operator_counter_assignments (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			branch_id TEXT,
			user_id TEXT,
			counter_id TEXT,
			unassigned_at INTEGER
		);
	`).Error)
	return db
}

func TestCallerUseCase_ExecuteAction(t *testing.T) {
	db := newCallerTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO queue_journeys (id, queue_id, tenant_id, branch_id, counter_id) VALUES ('j-1','q-1','t-1','b-1','c-1')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, is_active) VALUES ('client-1','t-1','b-1','bs-1','c-1',1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, is_active) VALUES ('client-2','t-1','b-1','bs-1','c-2',1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO operator_counter_assignments (id, tenant_id, branch_id, user_id, counter_id, unassigned_at) VALUES ('a-1','t-1','b-1','u-1','c-1',NULL)`).Error)

	t.Run("Success", func(t *testing.T) {
		queueUC := &stubQueueUC{res: &queueModel.QueueResponse{ID: "q-1", TicketNo: "A001", QueueNo: 1, Status: "calling"}}
		uc := NewCallerUseCase(db, queueUC)
		ctx := authcontext.WithUserID(database.SetOrganizationContext(context.Background(), "t-1"), "u-1")
		ctx = database.SetBranchContext(ctx, "b-1")

		res, err := uc.ExecuteAction(ctx, "client-1", "j-1", queueModel.QueueActionCall)
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, "j-1", res.JourneyID)
		assert.Equal(t, "q-1", queueUC.queueID)
		assert.Equal(t, queueModel.QueueActionCall, queueUC.action)
	})

	t.Run("ForbiddenWhenClientCounterMismatch", func(t *testing.T) {
		queueUC := &stubQueueUC{res: &queueModel.QueueResponse{ID: "q-1"}}
		uc := NewCallerUseCase(db, queueUC)
		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		_, err := uc.ExecuteAction(ctx, "client-2", "j-1", queueModel.QueueActionCall)
		require.ErrorIs(t, err, exception.ErrForbidden)
	})

	t.Run("UnauthorizedWhenClientMissing", func(t *testing.T) {
		queueUC := &stubQueueUC{res: &queueModel.QueueResponse{ID: "q-1"}}
		uc := NewCallerUseCase(db, queueUC)
		ctx := database.SetOrganizationContext(context.Background(), "t-1")
		ctx = database.SetBranchContext(ctx, "b-1")
		_, err := uc.ExecuteAction(ctx, "", "j-1", queueModel.QueueActionCall)
		require.ErrorIs(t, err, exception.ErrUnauthorized)
	})

	t.Run("ForbiddenWhenOperatorAssignmentMissing", func(t *testing.T) {
		queueUC := &stubQueueUC{res: &queueModel.QueueResponse{ID: "q-1"}}
		uc := NewCallerUseCase(db, queueUC)
		ctx := authcontext.WithUserID(database.SetOrganizationContext(context.Background(), "t-1"), "u-2")
		ctx = database.SetBranchContext(ctx, "b-1")
		_, err := uc.ExecuteAction(ctx, "client-1", "j-1", queueModel.QueueActionCall)
		require.ErrorIs(t, err, exception.ErrForbidden)
	})
}
