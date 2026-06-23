package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/entity"
	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/queue/model"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newQueueTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Queue{}, &entity.QueueJourney{}, &entity.VisitJourney{}, &queueCounterRow{}))
	return db
}

func TestQueueRepository_NextQueueNumber_Increments(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()
	date := time.Date(2026, 6, 24, 4, 0, 0, 0, time.UTC)

	first, err := repo.NextQueueNumber(ctx, "tenant-1", "branch-1", date, "A")
	require.NoError(t, err)
	require.Equal(t, 1, first)

	second, err := repo.NextQueueNumber(ctx, "tenant-1", "branch-1", date, "A")
	require.NoError(t, err)
	require.Equal(t, 2, second)
}

func TestQueueRepository_ExistsRegistration_DuplicateDetects(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", PatientID: "p-1", PatientName: "John", Status: entity.QueueStatusWaiting}).Error)

	exists, err := repo.ExistsRegistration(ctx, "tenant-1", "branch-1", "2026-06-24", "p-1", "John")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestQueueRepository_CreateRegistration_WritesQueueAndJourney(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	queue := &entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", TicketNo: "A001", QueueNo: 1, Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"}
	journey := &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending}
	visit := &entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", EventType: "registration"}

	require.NoError(t, repo.CreateRegistration(ctx, queue, journey, visit))

	var saved entity.Queue
	require.NoError(t, db.First(&saved, "id = ?", "q-1").Error)
	require.Equal(t, "j-1", saved.CurrentJourneyID)
}

func TestQueueRepository_UpdateQueueState_WritesStateAndVisit(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", TicketNo: "A001", QueueNo: 1, Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusCalling}).Error)

	visit := &entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", EventType: "call"}
	queue := &entity.Queue{ID: "q-1", Status: entity.QueueStatusServing, UpdatedAt: 123}
	journey := &entity.QueueJourney{ID: "j-1", Status: entity.JourneyStatusServing, UpdatedAt: 123}

	require.NoError(t, repo.UpdateQueueState(ctx, queue, journey, visit))

	var savedQueue entity.Queue
	require.NoError(t, db.First(&savedQueue, "id = ?", "q-1").Error)
	require.Equal(t, entity.QueueStatusServing, savedQueue.Status)

	var savedJourney entity.QueueJourney
	require.NoError(t, db.First(&savedJourney, "id = ?", "j-1").Error)
	require.Equal(t, entity.JourneyStatusServing, savedJourney.Status)

	var savedVisit entity.VisitJourney
	require.NoError(t, db.First(&savedVisit, "id = ?", "v-1").Error)
	require.Equal(t, "call", savedVisit.EventType)
}

func TestQueueRepository_UpdateQueueState_RollsBackOnJourneyError(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", TicketNo: "A001", QueueNo: 1, Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1"}).Error)

	visit := &entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", EventType: "call"}
	queue := &entity.Queue{ID: "q-1", Status: entity.QueueStatusServing, UpdatedAt: 123}
	journey := &entity.QueueJourney{ID: "missing", Status: entity.JourneyStatusServing, UpdatedAt: 123}

	err := repo.UpdateQueueState(ctx, queue, journey, visit)
	require.Error(t, err)

	var savedQueue entity.Queue
	require.NoError(t, db.First(&savedQueue, "id = ?", "q-1").Error)
	require.Equal(t, entity.QueueStatusCalling, savedQueue.Status)

	var savedVisit entity.VisitJourney
	require.Error(t, db.First(&savedVisit, "id = ?", "v-1").Error)
}

func TestQueueRepository_ListQueues_FiltersByDateAndTenantBranch(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1, CurrentJourneyID: "j-1"}).Error)
	require.NoError(t, db.Create(&entity.Queue{ID: "q-2", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A002", QueueNo: 2, CurrentJourneyID: "j-2"}).Error)
	require.NoError(t, db.Create(&entity.Queue{ID: "q-3", TenantID: "tenant-2", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A003", QueueNo: 3, CurrentJourneyID: "j-3"}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-2", QueueID: "q-2", TenantID: "tenant-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-3", QueueID: "q-3", TenantID: "tenant-2", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)

	queues, err := repo.ListQueues(ctx, "tenant-1", "branch-1", model.ListQueuesRequest{Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", ServiceID: "svc-1"})
	require.NoError(t, err)
	require.Len(t, queues, 2)
	assert.Contains(t, []string{queues[0].ID, queues[1].ID}, "q-1")
	assert.Contains(t, []string{queues[0].ID, queues[1].ID}, "q-2")
}

func TestQueueRepository_ListActiveJourneys_FiltersByTenantBranchServiceCounter(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1}).Error)
	require.NoError(t, db.Create(&entity.Queue{ID: "q-2", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A002", QueueNo: 2}).Error)
	require.NoError(t, db.Create(&entity.Queue{ID: "q-3", TenantID: "tenant-2", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A003", QueueNo: 3}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", ServiceID: "svc-1", CounterID: "counter-1", SeqNo: 1, Status: entity.JourneyStatusCalling}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-2", QueueID: "q-2", TenantID: "tenant-1", ServiceID: "svc-1", CounterID: "counter-2", SeqNo: 1, Status: entity.JourneyStatusCompleted}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-3", QueueID: "q-3", TenantID: "tenant-2", ServiceID: "svc-1", CounterID: "counter-1", SeqNo: 1, Status: entity.JourneyStatusCalling}).Error)

	journeys, err := repo.ListActiveJourneys(ctx, "tenant-1", "branch-1", model.QueueJourneyListRequest{QueueDate: "2026-06-24", ServiceID: "svc-1", CounterID: "counter-1"})
	require.NoError(t, err)
	require.Len(t, journeys, 1)
	assert.Equal(t, "j-1", journeys[0].ID)
}

func TestQueueRepository_FindVisitJourneysByQueueID_TenantScoped(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1}).Error)
	require.NoError(t, db.Create(&entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", EventType: "registration", CreatedAt: 100}).Error)
	require.NoError(t, db.Create(&entity.VisitJourney{ID: "v-2", QueueID: "q-1", TenantID: "tenant-1", EventType: "call", CreatedAt: 200}).Error)
	require.NoError(t, db.Create(&entity.VisitJourney{ID: "v-3", QueueID: "q-1", TenantID: "tenant-2", EventType: "registration", CreatedAt: 50}).Error)

	visits, err := repo.FindVisitJourneysByQueueID(ctx, "tenant-1", "q-1")
	require.NoError(t, err)
	require.Len(t, visits, 2)
	assert.Equal(t, "v-1", visits[0].ID)
	assert.Equal(t, "v-2", visits[1].ID)
}

func TestQueueRepository_GetQueueStats_AggregatesCorrectly(t *testing.T) {
	db := newQueueTestDB(t)
	repo := NewQueueRepository(db)
	ctx := context.Background()
	date := "2026-06-24"

	require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", QueueDate: date, Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1}).Error)
	require.NoError(t, db.Create(&entity.Queue{ID: "q-2", TenantID: "t-1", BranchID: "b-1", QueueDate: date, Status: entity.QueueStatusCompleted, TicketNo: "A002", QueueNo: 2}).Error)
	require.NoError(t, db.Create(&entity.Queue{ID: "q-3", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-23", Status: entity.QueueStatusWaiting, TicketNo: "A003", QueueNo: 3}).Error) // Other date

	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
	require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-2", QueueID: "q-2", TenantID: "t-1", ServiceID: "svc-2", SeqNo: 1, Status: entity.JourneyStatusCompleted}).Error)

	stats, err := repo.GetQueueStats(ctx, "t-1", "b-1", date)
	require.NoError(t, err)
	assert.Equal(t, int64(2), stats.TotalQueuesToday)
	assert.Equal(t, int64(1), stats.TotalActiveJourneys)
	assert.Equal(t, int64(1), stats.TotalCompletedVisits)
	assert.Equal(t, int64(1), stats.WaitingByService["svc-1"])
	assert.Equal(t, int64(0), stats.WaitingByService["svc-2"])
}
