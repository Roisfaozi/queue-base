//go:build integration
// +build integration

package integration

import (
	"context"
	"testing"
	"time"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	qmsClientEntity "github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	queueEntity "github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	signageUsecase "github.com/Roisfaozi/queue-base/internal/modules/signage/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignageIntegration(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	tenantID := uuid.NewString()
	branchID := uuid.NewString()
	serviceID := uuid.NewString()
	branchServiceID := uuid.NewString()
	counterID := uuid.NewString()
	clientID := uuid.NewString()
	now := time.Now().UnixMilli()

	require.NoError(t, env.DB.Create(&branchEntity.Organization{ID: tenantID, Code: "sig-int", Name: "Signage Int", Slug: "sig-int-" + tenantID[:8], OwnerID: uuid.NewString(), Status: branchEntity.OrgStatusActive}).Error)
	require.NoError(t, env.DB.Create(&branchEntity.Branch{ID: branchID, TenantID: tenantID, Code: "BR-SIG", Name: "Branch S", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, env.DB.Create(&serviceEntity.Service{ID: serviceID, TenantID: tenantID, Code: "SS", Name: "Service S", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, env.DB.Create(&serviceEntity.BranchService{ID: branchServiceID, TenantID: tenantID, BranchID: branchID, ServiceID: serviceID, IsActive: true}).Error)
	require.NoError(t, env.DB.Create(&counterEntity.Counter{ID: counterID, TenantID: tenantID, BranchID: branchID, BranchServiceID: branchServiceID, Code: "CS", Name: "Counter S", Status: counterEntity.CounterStatusActive}).Error)
	require.NoError(t, env.DB.Create(&qmsClientEntity.QMSClient{ID: clientID, TenantID: tenantID, BranchID: branchID, BranchServiceID: &branchServiceID, CounterID: &counterID, ClientType: qmsClientEntity.ClientTypeSignage, Name: "Signage Client", IsActive: true}).Error)

	q1 := uuid.NewString()
	j1 := uuid.NewString()
	q2 := uuid.NewString()
	j2 := uuid.NewString()
	require.NoError(t, env.DB.Create(&queueEntity.Queue{ID: q1, TenantID: tenantID, BranchID: branchID, QueueDate: time.Now().Format("2006-01-02"), TicketNo: "SS001", QueueNo: 1, Status: queueEntity.QueueStatusWaiting, CurrentJourneyID: j1, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, env.DB.Create(&queueEntity.QueueJourney{ID: j1, QueueID: q1, TenantID: tenantID, BranchID: branchID, ServiceID: serviceID, CounterID: counterID, Status: queueEntity.JourneyStatusPending, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, env.DB.Create(&queueEntity.Queue{ID: q2, TenantID: tenantID, BranchID: branchID, QueueDate: time.Now().Format("2006-01-02"), TicketNo: "SS002", QueueNo: 2, Status: queueEntity.QueueStatusCalling, CurrentJourneyID: j2, CreatedAt: now, UpdatedAt: now}).Error)
	require.NoError(t, env.DB.Create(&queueEntity.QueueJourney{ID: j2, QueueID: q2, TenantID: tenantID, BranchID: branchID, ServiceID: serviceID, CounterID: counterID, Status: queueEntity.JourneyStatusCalling, CreatedAt: now, UpdatedAt: now}).Error)

	uc := signageUsecase.NewSignageUseCase(env.DB, nil, env.Logger)
	ctx := database.SetBranchContext(database.SetOrganizationContext(context.Background(), tenantID), branchID)

	calls, err := uc.GetCurrentCalls(ctx, clientID)
	require.NoError(t, err)
	require.Len(t, calls, 1)
	assert.Equal(t, q2, calls[0].QueueID)

	queues, err := uc.GetQueues(ctx, clientID)
	require.NoError(t, err)
	require.NotEmpty(t, queues)
	assert.Equal(t, q1, queues[0].ID)
}
