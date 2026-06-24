//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	queueModule "github.com/Roisfaozi/queue-base/internal/modules/queue"
	queueEntity "github.com/Roisfaozi/queue-base/internal/modules/queue/entity"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	settingsModule "github.com/Roisfaozi/queue-base/internal/modules/settings"
	settingsEntity "github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQMSQueueIntegration_LifecycleAndSettingsGuard(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	v := validator.New()
	settingsMod := settingsModule.NewSettingsModule(env.DB, v)
	queueMod := queueModule.NewQueueModule(env.DB, v, settingsMod.QueueSettingsResolver)

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	regServiceID := uuid.New().String()
	pharmacyServiceID := uuid.New().String()
	counterID := uuid.New().String()

	require.NoError(t, env.DB.Create(&branchEntity.Branch{ID: branchID, TenantID: tenantID, Code: "BR1", Name: "Main Branch", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, env.DB.Create(&serviceEntity.Service{ID: regServiceID, TenantID: tenantID, Code: "RG", Name: "Registration", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, env.DB.Create(&serviceEntity.Service{ID: pharmacyServiceID, TenantID: tenantID, Code: "PH", Name: "Pharmacy", Status: serviceEntity.ServiceStatusActive, IsPharmacy: true}).Error)
	require.NoError(t, env.DB.Create(&counterEntity.Counter{ID: counterID, TenantID: tenantID, BranchID: branchID, Code: "C1", Name: "Counter 1", Status: counterEntity.CounterStatusActive}).Error)
	require.NoError(t, env.DB.Create(&settingsEntity.Setting{ID: uuid.New().String(), TenantID: tenantID, ScopeType: settingsEntity.ScopeTypeService, ScopeID: pharmacyServiceID, Key: settingsModel.SettingKeyPharmacyFlowEnabled, Value: "true", ValueType: "boolean", IsActive: true}).Error)
	require.NoError(t, env.DB.Create(&settingsEntity.Setting{ID: uuid.New().String(), TenantID: tenantID, ScopeType: settingsEntity.ScopeTypeService, ScopeID: pharmacyServiceID, Key: settingsModel.SettingKeyRequireCounterForService, Value: "true", ValueType: "boolean", IsActive: true}).Error)

	ctx := database.SetOrganizationContext(context.Background(), tenantID)
	ctx = database.SetBranchContext(ctx, branchID)

	emptyJourneys, err := queueMod.QueueUseCase.ListActiveJourneys(ctx, queueModel.QueueJourneyListRequest{ServiceID: regServiceID})
	require.NoError(t, err)
	assert.Empty(t, emptyJourneys)

	queueRes, err := queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: regServiceID, PatientName: "John Queue"})
	require.NoError(t, err)
	require.NotNil(t, queueRes)

	forwarded, err := queueMod.QueueUseCase.ForwardQueue(ctx, queueRes.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: pharmacyServiceID, DestinationCounterID: counterID})
	require.NoError(t, err)
	assert.Equal(t, queueRes.ID, forwarded.ID)
	assert.NotEqual(t, queueRes.CurrentJourneyID, forwarded.CurrentJourneyID)

	serving, err := queueMod.QueueUseCase.TransitionQueue(ctx, queueRes.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
	require.NoError(t, err)
	assert.Equal(t, queueEntity.QueueStatusCalling, serving.Status)

	_, err = queueMod.QueueUseCase.TransitionQueue(ctx, queueRes.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
	assert.ErrorIs(t, err, exception.ErrBadRequest)

	visits, err := queueMod.QueueUseCase.GetVisitJourneys(ctx, queueRes.ID)
	require.NoError(t, err)
	assert.Len(t, visits, 3)
	assert.Equal(t, "registration", visits[0].EventType)
	assert.Equal(t, "forward", visits[1].EventType)
	assert.Equal(t, "call", visits[2].EventType)

	stats, err := queueMod.QueueUseCase.GetQueueStats(ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats.TotalQueuesToday)
	assert.Equal(t, int64(1), stats.TotalActiveJourneys)

	_, err = queueMod.QueueUseCase.ForwardQueue(ctx, queueRes.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: pharmacyServiceID})
	assert.ErrorIs(t, err, exception.ErrForbidden)
}
