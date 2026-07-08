//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	queueModule "github.com/Roisfaozi/queue-base/internal/modules/queue"
	queueEntity "github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue_config"
	settingsEntity "github.com/Roisfaozi/queue-base/internal/modules/queue_config/entity"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_QMSJourneyLifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	db := env.DB
	v := validator.New()
	log := env.Logger

	// Migrate settings typed config to ensure they exist
	db.AutoMigrate(
		&settingsEntity.TenantQueueSetting{},
		&settingsEntity.BranchQueueSetting{},
		&settingsEntity.BranchServiceQueueSetting{},
		&settingsEntity.CounterQueueSetting{},
	)

	settingsMod := settings.NewQueueConfigModule(db, v, log, nil)
	queueMod := queueModule.NewQueueModule(db, v, settingsMod.QueueConfigResolver, log, nil)

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	regServiceID := uuid.New().String()
	regCounterID := uuid.New().String()

	require.NoError(t, db.Create(&branchEntity.Branch{ID: branchID, TenantID: tenantID, Code: "BR-LIFECYCLE", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, db.Create(&serviceEntity.Service{ID: regServiceID, TenantID: tenantID, Code: "REG", Name: "Registration", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, db.Create(&serviceEntity.BranchService{ID: uuid.New().String(), TenantID: tenantID, BranchID: branchID, ServiceID: regServiceID, IsActive: true}).Error)
	require.NoError(t, db.Create(&counterEntity.Counter{ID: regCounterID, TenantID: tenantID, BranchID: branchID, Code: "C-REG", Status: counterEntity.CounterStatusActive}).Error)

	ctx := database.SetOrganizationContext(context.Background(), tenantID)
	ctx = database.SetBranchContext(ctx, branchID)

	// Settings setup: allow_recall = true
	require.NoError(t, db.Create(&settingsEntity.TenantQueueSetting{ID: uuid.New().String(), TenantID: tenantID, AllowRecall: true}).Error)

	t.Run("Full Lifecycle: Register -> Call -> Recall -> Serve -> Complete", func(t *testing.T) {
		// 1. Register
		q, err := queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: regServiceID, PatientName: "Lifecycle Test"})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusWaiting, q.Status)
		assert.NotEmpty(t, q.CurrentJourneyID)

		// Wait a bit to ensure timestamps differ if needed
		time.Sleep(10 * time.Millisecond)

		// 2. Call
		qCall, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusCalling, qCall.Status)
		assert.Equal(t, q.CurrentJourneyID, qCall.CurrentJourneyID)

		// 3. Recall (handled as repeated Call)
		qRecall, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusCalling, qRecall.Status) // remains calling

		// Verify Visit Journey has the "recall" event
		visits, err := queueMod.QueueUseCase.GetVisitJourneys(ctx, q.ID)
		require.NoError(t, err)
		assert.Len(t, visits, 3) // created, call, recall
		assert.Equal(t, "recall", visits[2].EventType)

		// 4. Serve
		qServe, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionServe})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusServing, qServe.Status)

		// 5. Complete
		qComplete, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionComplete})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusCompleted, qComplete.Status)
	})

	t.Run("Short Lifecycle: Register -> Skip -> Call -> Skip", func(t *testing.T) {
		// Settings setup: allow_skip = true
		require.NoError(t, db.Model(&settingsEntity.TenantQueueSetting{}).Where("tenant_id = ?", tenantID).Update("allow_skip", true).Error)

		// 1. Register
		q, err := queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: regServiceID, PatientName: "Skip Test"})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusWaiting, q.Status)

		// 2. Skip from waiting
		qSkip1, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionSkip})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusSkipped, qSkip1.Status)

		// 3. Call a skipped journey
		qCall, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusCalling, qCall.Status)

		// 4. Skip from calling
		qSkip2, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionSkip})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusSkipped, qSkip2.Status)
	})
}
