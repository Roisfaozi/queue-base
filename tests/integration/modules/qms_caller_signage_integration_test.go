//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	qmsClientEntity "github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	queueModule "github.com/Roisfaozi/queue-base/internal/modules/queue"
	queueEntity "github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueConfigModule "github.com/Roisfaozi/queue-base/internal/modules/queue_config"
	settingsEntity "github.com/Roisfaozi/queue-base/internal/modules/queue_config/entity"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/signage/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntegration_CallerSignageLifecycle(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	db := env.DB
	v := validator.New()
	log := env.Logger

	db.AutoMigrate(
		&settingsEntity.TenantQueueSetting{},
		&settingsEntity.BranchQueueSetting{},
		&settingsEntity.BranchServiceQueueSetting{},
		&settingsEntity.CounterQueueSetting{},
	)

	settingsMod := queueConfigModule.NewQueueConfigModule(db, v, log)
	queueMod := queueModule.NewQueueModule(db, v, settingsMod.QueueConfigResolver, log)
	signageUC := usecase.NewSignageUseCase(db, queueMod.QueueUseCase, log)

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	serviceID := uuid.New().String()
	branchServiceID := uuid.New().String()
	counterID := uuid.New().String()
	clientID := uuid.New().String()

	require.NoError(t, db.Create(&branchEntity.Branch{ID: branchID, TenantID: tenantID, Code: "BR-CSR", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, db.Create(&serviceEntity.Service{ID: serviceID, TenantID: tenantID, Code: "RG", Name: "Registration", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, db.Create(&serviceEntity.BranchService{ID: branchServiceID, TenantID: tenantID, BranchID: branchID, ServiceID: serviceID, IsActive: true}).Error)
	require.NoError(t, db.Create(&counterEntity.Counter{ID: counterID, TenantID: tenantID, BranchID: branchID, BranchServiceID: branchServiceID, Code: "CSR", Name: "CSR Counter", Status: counterEntity.CounterStatusActive}).Error)
	require.NoError(t, db.Create(&qmsClientEntity.QMSClient{
		ID: clientID, TenantID: tenantID, BranchID: branchID, BranchServiceID: &branchServiceID, CounterID: &counterID,
		ClientType: "signage", Name: "Signage CSR", IsActive: true,
	}).Error)

	ctx := database.SetOrganizationContext(context.Background(), tenantID)
	ctx = database.SetBranchContext(ctx, branchID)

	t.Run("Register, Call, then Verify Signage State", func(t *testing.T) {
		// Register queue
		q, err := queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: serviceID, PatientName: "Signage Bound"})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusWaiting, q.Status)

		// Forward to counter so journey has counter_id
		qForward, err := queueMod.QueueUseCase.ForwardQueue(ctx, q.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: serviceID, DestinationCounterID: counterID})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusWaiting, qForward.Status)

		// Call — applies for service under signage branch
		qCall, err := queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
		require.NoError(t, err)
		assert.Equal(t, queueEntity.QueueStatusCalling, qCall.Status)

		// Signage GetCurrentCalls bound to same branch_service + counter
		time.Sleep(10 * time.Millisecond)
		calls, err := signageUC.GetCurrentCalls(ctx, clientID)
		require.NoError(t, err)
		require.NotEmpty(t, calls)
		assert.Equal(t, q.TicketNo, calls[0].TicketNo)
		assert.Equal(t, counterID, calls[0].CounterID)
		assert.Equal(t, serviceID, calls[0].ServiceID)
	})

	t.Run("Complete queue; Signage calls hide completed", func(t *testing.T) {
		// Serve
		q, err := queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: serviceID, PatientName: "Signage Hide"})
		require.NoError(t, err)

		_, err = queueMod.QueueUseCase.ForwardQueue(ctx, q.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: serviceID, DestinationCounterID: counterID})
		require.NoError(t, err)

		_, err = queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
		require.NoError(t, err)

		_, err = queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionServe})
		require.NoError(t, err)

		_, err = queueMod.QueueUseCase.TransitionQueue(ctx, q.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionComplete})
		require.NoError(t, err)

		// After complete, signage should only show the first calling ticket
		calls, err := signageUC.GetCurrentCalls(ctx, clientID)
		require.NoError(t, err)
		if len(calls) > 0 {
			// The completed ticket might also be hidden by different service filter
			for _, c := range calls {
				assert.NotEqual(t, q.ID, c.QueueID, "completed ticket should not appear in signage current calls")
			}
		}
	})
}
