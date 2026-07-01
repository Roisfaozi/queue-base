package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/queue/model"
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

func TestQueueRepository(t *testing.T) {
	ctx := context.Background()
	date := time.Date(2026, 6, 24, 4, 0, 0, 0, time.UTC)

	t.Run("NextQueueNumber", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			tenantID string
			branchID string
			prefix   string
			want     int
			wantErr  bool
		}{
			{
				name:     "Positive_Increments",
				category: "positive",
				tenantID: "tenant-1",
				branchID: "branch-1",
				prefix:   "A",
				want:     1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				first, err := repo.NextQueueNumber(ctx, tt.tenantID, tt.branchID, date, tt.prefix)
				require.NoError(t, err)
				require.Equal(t, tt.want, first)

				second, err := repo.NextQueueNumber(ctx, tt.tenantID, tt.branchID, date, tt.prefix)
				require.NoError(t, err)
				require.Equal(t, tt.want+1, second)
			})
		}
	})

	t.Run("ExistsRegistration", func(t *testing.T) {
		tests := []struct {
			name        string
			category    string
			tenantID    string
			branchID    string
			queueID     string
			patientID   string
			patientName string
			want        bool
			wantErr     bool
		}{
			{
				name:        "Positive_DuplicateDetects",
				category:    "positive",
				tenantID:    "tenant-1",
				branchID:    "branch-1",
				queueID:     "q-1",
				patientID:   "p-1",
				patientName: "John",
				want:        true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				require.NoError(t, db.Create(&entity.Queue{
					ID: tt.queueID, TenantID: tt.tenantID, BranchID: tt.branchID,
					QueueDate: "2026-06-24", PatientID: tt.patientID,
					PatientName: tt.patientName, Status: entity.QueueStatusWaiting,
				}).Error)

				exists, err := repo.ExistsRegistration(ctx, tt.tenantID, tt.branchID, "2026-06-24", tt.patientID, tt.patientName)
				require.NoError(t, err)
				require.Equal(t, tt.want, exists)
			})
		}
	})

	t.Run("CreateRegistration", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
		}{
			{
				name:     "Positive_WritesQueueAndJourney",
				category: "positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				queue := &entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", TicketNo: "A001", QueueNo: 1, Status: entity.QueueStatusWaiting, CurrentJourneyID: "j-1"}
				journey := &entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusPending}
				visit := &entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", EventType: "registration"}

				require.NoError(t, repo.CreateRegistration(ctx, queue, journey, visit))

				var saved entity.Queue
				require.NoError(t, db.First(&saved, "id = ?", "q-1").Error)
				require.Equal(t, "j-1", saved.CurrentJourneyID)
			})
		}
	})

	t.Run("UpdateQueueState", func(t *testing.T) {
		tests := []struct {
			name      string
			category  string
			queueID   string
			journeyID string
			visitID   string
			wantErr   bool
		}{
			{
				name:      "Positive_WritesStateAndVisit",
				category:  "positive",
				queueID:   "q-1",
				journeyID: "j-1",
				visitID:   "v-1",
			},
			{
				name:      "Negative_RollsBackOnJourneyError",
				category:  "negative",
				queueID:   "q-1",
				journeyID: "missing",
				visitID:   "v-1",
				wantErr:   true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				require.NoError(t, db.Create(&entity.Queue{
					ID: tt.queueID, TenantID: "tenant-1", BranchID: "branch-1",
					QueueDate: "2026-06-24", TicketNo: "A001", QueueNo: 1,
					Status: entity.QueueStatusCalling, CurrentJourneyID: "j-1",
				}).Error)
				if tt.journeyID == "j-1" {
					require.NoError(t, db.Create(&entity.QueueJourney{
						ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1",
						ServiceID: "s-1", SeqNo: 1, Status: entity.JourneyStatusCalling,
					}).Error)
				}

				visit := &entity.VisitJourney{ID: tt.visitID, QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", EventType: "call"}
				queue := &entity.Queue{ID: tt.queueID, TenantID: "tenant-1", BranchID: "branch-1", Status: entity.QueueStatusServing, UpdatedAt: 123}
				journey := &entity.QueueJourney{ID: tt.journeyID, TenantID: "tenant-1", BranchID: "branch-1", Status: entity.JourneyStatusServing, UpdatedAt: 123}

				err := repo.UpdateQueueState(ctx, queue, journey, visit)

				if tt.wantErr {
					require.Error(t, err)
					var savedQueue entity.Queue
					require.NoError(t, db.First(&savedQueue, "id = ?", tt.queueID).Error)
					require.Equal(t, entity.QueueStatusCalling, savedQueue.Status)
					var savedVisit entity.VisitJourney
					require.Error(t, db.First(&savedVisit, "id = ?", tt.visitID).Error)
				} else {
					require.NoError(t, err)
					var savedQueue entity.Queue
					require.NoError(t, db.First(&savedQueue, "id = ?", tt.queueID).Error)
					require.Equal(t, entity.QueueStatusServing, savedQueue.Status)
					var savedJourney entity.QueueJourney
					require.NoError(t, db.First(&savedJourney, "id = ?", tt.journeyID).Error)
					require.Equal(t, entity.JourneyStatusServing, savedJourney.Status)
					var savedVisit entity.VisitJourney
					require.NoError(t, db.First(&savedVisit, "id = ?", tt.visitID).Error)
					require.Equal(t, "call", savedVisit.EventType)
				}
			})
		}
	})

	t.Run("ListQueues", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
		}{
			{
				name:     "Positive_FiltersByDateAndTenantBranch",
				category: "positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1, CurrentJourneyID: "j-1"}).Error)
				require.NoError(t, db.Create(&entity.Queue{ID: "q-2", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A002", QueueNo: 2, CurrentJourneyID: "j-2"}).Error)
				require.NoError(t, db.Create(&entity.Queue{ID: "q-3", TenantID: "tenant-2", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A003", QueueNo: 3, CurrentJourneyID: "j-3"}).Error)
				require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
				require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-2", QueueID: "q-2", TenantID: "tenant-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
				require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-3", QueueID: "q-3", TenantID: "tenant-2", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)

				queues, err := repo.ListQueues(ctx, "tenant-1", "branch-1", model.ListQueuesRequest{Status: entity.QueueStatusWaiting, QueueDate: "2026-06-24", ServiceID: "svc-1"})
				require.NoError(t, err)
				require.Len(t, queues, 2)
				assert.Contains(t, []string{queues[0].ID, queues[1].ID}, "q-1")
				assert.Contains(t, []string{queues[0].ID, queues[1].ID}, "q-2")
			})
		}
	})

	t.Run("ListActiveJourneys", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
		}{
			{
				name:     "Positive_FiltersByTenantBranchServiceCounter",
				category: "positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

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
			})
		}
	})

	t.Run("FindVisitJourneysByQueueID", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
		}{
			{
				name:     "Positive_TenantScoped",
				category: "positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1}).Error)
				require.NoError(t, db.Create(&entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", EventType: "registration", CreatedAt: 100}).Error)
				require.NoError(t, db.Create(&entity.VisitJourney{ID: "v-2", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", EventType: "call", CreatedAt: 200}).Error)
				require.NoError(t, db.Create(&entity.VisitJourney{ID: "v-3", QueueID: "q-1", TenantID: "tenant-2", BranchID: "branch-2", EventType: "registration", CreatedAt: 50}).Error)

				visits, err := repo.FindVisitJourneysByQueueID(ctx, "tenant-1", "branch-1", "q-1")
				require.NoError(t, err)
				require.Len(t, visits, 2)
				assert.Equal(t, "v-1", visits[0].ID)
				assert.Equal(t, "v-2", visits[1].ID)
			})
		}
	})

	t.Run("CreateForwarding", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			prepare  func(t *testing.T, db *gorm.DB)
			wantErr  error
		}{
			{
				name:     "Positive_AutoIncrementsSequence",
				category: "positive",
			},
			{
				name:     "Negative_RejectsStaleCurrentJourney",
				category: "negative",
				prepare: func(t *testing.T, db *gorm.DB) {
					require.NoError(t, db.Table("queues").Where("id = ?", "q-1").Update("current_journey_id", "j-other").Error)
				},
				wantErr: gorm.ErrRecordNotFound,
			},
			{
				name:     "Negative_RejectsExistingSecondActiveJourney",
				category: "negative",
				prepare: func(t *testing.T, db *gorm.DB) {
					require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-active", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", ServiceID: "svc-extra", SeqNo: 2, Status: entity.JourneyStatusPending}).Error)
				},
				wantErr: gorm.ErrDuplicatedKey,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)

				require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", QueueDate: "2026-06-24", TicketNo: "A001", QueueNo: 1, CurrentJourneyID: "j-1"}).Error)
				require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
				if tt.prepare != nil {
					tt.prepare(t, db)
				}

				queue := &entity.Queue{ID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", CurrentJourneyID: "j-2", UpdatedAt: 123}
				currentJourney := &entity.QueueJourney{ID: "j-1", TenantID: "tenant-1", BranchID: "branch-1", Status: entity.JourneyStatusForwarded, UpdatedAt: 123}
				nextJourney := &entity.QueueJourney{ID: "j-2", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", ServiceID: "svc-2", Status: entity.JourneyStatusPending}
				visit := &entity.VisitJourney{ID: "v-1", QueueID: "q-1", TenantID: "tenant-1", BranchID: "branch-1", EventType: "forward"}

				err := repo.CreateForwarding(ctx, queue, currentJourney, nextJourney, visit)
				if tt.wantErr != nil {
					require.ErrorIs(t, err, tt.wantErr)
					var notSaved entity.QueueJourney
					require.Error(t, db.First(&notSaved, "id = ?", "j-2").Error)
					return
				}
				require.NoError(t, err)

				var savedJourney entity.QueueJourney
				require.NoError(t, db.First(&savedJourney, "id = ?", "j-2").Error)
				assert.Equal(t, 2, savedJourney.SeqNo)
			})
		}
	})

	t.Run("GetQueueStats", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
		}{
			{
				name:     "Positive_AggregatesCorrectly",
				category: "positive",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db := newQueueTestDB(t)
				repo := NewQueueRepository(db)
				dateStr := "2026-06-24"

				require.NoError(t, db.Create(&entity.Queue{ID: "q-1", TenantID: "t-1", BranchID: "b-1", QueueDate: dateStr, Status: entity.QueueStatusWaiting, TicketNo: "A001", QueueNo: 1}).Error)
				require.NoError(t, db.Create(&entity.Queue{ID: "q-2", TenantID: "t-1", BranchID: "b-1", QueueDate: dateStr, Status: entity.QueueStatusCompleted, TicketNo: "A002", QueueNo: 2}).Error)
				require.NoError(t, db.Create(&entity.Queue{ID: "q-3", TenantID: "t-1", BranchID: "b-1", QueueDate: "2026-06-23", Status: entity.QueueStatusWaiting, TicketNo: "A003", QueueNo: 3}).Error)

				require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-1", QueueID: "q-1", TenantID: "t-1", ServiceID: "svc-1", SeqNo: 1, Status: entity.JourneyStatusPending}).Error)
				require.NoError(t, db.Create(&entity.QueueJourney{ID: "j-2", QueueID: "q-2", TenantID: "t-1", ServiceID: "svc-2", SeqNo: 1, Status: entity.JourneyStatusCompleted}).Error)

				stats, err := repo.GetQueueStats(ctx, "t-1", "b-1", dateStr)
				require.NoError(t, err)
				assert.Equal(t, int64(2), stats.TotalQueuesToday)
				assert.Equal(t, int64(1), stats.TotalActiveJourneys)
				assert.Equal(t, int64(1), stats.TotalCompletedVisits)
				assert.Equal(t, int64(1), stats.WaitingByService["svc-1"])
				assert.Equal(t, int64(0), stats.WaitingByService["svc-2"])
			})
		}
	})
}
