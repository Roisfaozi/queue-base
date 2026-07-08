//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	callerUsecase "github.com/Roisfaozi/queue-base/internal/modules/caller/usecase"
	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	qmsClientEntity "github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	queueModule "github.com/Roisfaozi/queue-base/internal/modules/queue"
	queueEntity "github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue_config"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCallerActionsIntegration(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	serviceID := uuid.New().String()
	branchServiceID := uuid.New().String()
	counterID := uuid.New().String()
	clientID := uuid.New().String()
	queueID := uuid.New().String()
	journeyID := uuid.New().String()
	now := time.Now().UnixMilli()

	require.NoError(t, env.DB.Create(&branchEntity.Organization{ID: tenantID, Code: "caller-it", Name: "Caller IT", Slug: "caller-it-" + tenantID, OwnerID: uuid.New().String(), Status: branchEntity.OrgStatusActive}).Error)
	require.NoError(t, env.DB.Create(&branchEntity.Branch{ID: branchID, TenantID: tenantID, Code: "BR-CALL", Name: "Caller Branch", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, env.DB.Create(&serviceEntity.Service{ID: serviceID, TenantID: tenantID, Code: "CALL", Name: "Caller Service", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, env.DB.Create(&serviceEntity.BranchService{ID: branchServiceID, TenantID: tenantID, BranchID: branchID, ServiceID: serviceID, IsActive: true}).Error)
	require.NoError(t, env.DB.Create(&counterEntity.Counter{ID: counterID, TenantID: tenantID, BranchID: branchID, BranchServiceID: branchServiceID, Code: "C-CALL", Name: "Caller Counter", Status: counterEntity.CounterStatusActive}).Error)
	require.NoError(t, env.DB.Create(&qmsClientEntity.QMSClient{ID: clientID, TenantID: tenantID, BranchID: branchID, BranchServiceID: &branchServiceID, CounterID: &counterID, ClientType: qmsClientEntity.ClientTypeCaller, Name: "Caller Client", IsActive: true}).Error)
	require.NoError(t, env.DB.Create(&queueEntity.Queue{ID: queueID, TenantID: tenantID, BranchID: branchID, QueueDate: time.Now().Format("2006-01-02"), TicketNo: "CALL001", QueueNo: 1, PatientName: "Caller Patient", Status: queueEntity.QueueStatusWaiting, CurrentJourneyID: journeyID, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, env.DB.Create(&queueEntity.QueueJourney{ID: journeyID, QueueID: queueID, TenantID: tenantID, BranchID: branchID, ServiceID: serviceID, CounterID: counterID, SeqNo: 1, Status: queueEntity.JourneyStatusPending, CreatedAt: now, UpdatedAt: now}).Error)

	v := validator.New()
	settingsMod := settings.NewQueueConfigModule(env.DB, v, env.Logger)
	queueMod := queueModule.NewQueueModule(env.DB, v, settingsMod.QueueConfigResolver, env.Logger)
	callerUC := callerUsecase.NewCallerUseCase(env.DB, queueMod.QueueUseCase, nil, nil)
	ctx := database.SetBranchContext(database.SetOrganizationContext(context.Background(), tenantID), branchID)

	tests := []struct {
		name       string
		action     string
		wantStatus string
	}{
		{name: "Call Ticket", action: queueModel.QueueActionCall, wantStatus: queueEntity.QueueStatusCalling},
		{name: "Serve Ticket", action: queueModel.QueueActionServe, wantStatus: queueEntity.QueueStatusServing},
		{name: "Complete Ticket", action: queueModel.QueueActionComplete, wantStatus: queueEntity.QueueStatusCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := callerUC.ExecuteAction(ctx, clientID, journeyID, tt.action)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, journeyID, res.JourneyID)
			assert.Equal(t, tt.wantStatus, res.Status)

			var queue queueEntity.Queue
			require.NoError(t, env.DB.First(&queue, "id = ?", queueID).Error)
			assert.Equal(t, tt.wantStatus, queue.Status)

			var journey queueEntity.QueueJourney
			require.NoError(t, env.DB.First(&journey, "id = ?", journeyID).Error)
			assert.Equal(t, tt.wantStatus, journey.Status)
		})
	}
}
