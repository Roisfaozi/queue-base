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
	"gorm.io/gorm"
)

type qmsDeps struct {
	db                *gorm.DB
	queueMod          *queueModule.QueueModule
	tenantID          string
	branchID          string
	regServiceID      string
	pharmacyServiceID string
	counterID         string
}

func setupQMSIntegration(t *testing.T) *qmsDeps {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	v := validator.New()
	settingsMod := settingsModule.NewSettingsModule(env.DB, v)
	queueMod := queueModule.NewQueueModule(env.DB, v, settingsMod.QueueSettingsResolver)

	deps := &qmsDeps{
		db:                env.DB,
		queueMod:          queueMod,
		tenantID:          uuid.New().String(),
		branchID:          uuid.New().String(),
		regServiceID:      uuid.New().String(),
		pharmacyServiceID: uuid.New().String(),
		counterID:         uuid.New().String(),
	}

	require.NoError(t, deps.db.Create(&branchEntity.Branch{ID: deps.branchID, TenantID: deps.tenantID, Code: "BR1", Name: "Main Branch", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, deps.db.Create(&serviceEntity.Service{ID: deps.regServiceID, TenantID: deps.tenantID, Code: "RG", Name: "Registration", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, deps.db.Create(&serviceEntity.Service{ID: deps.pharmacyServiceID, TenantID: deps.tenantID, Code: "PH", Name: "Pharmacy", Status: serviceEntity.ServiceStatusActive, IsPharmacy: true}).Error)
	require.NoError(t, deps.db.Create(&counterEntity.Counter{ID: deps.counterID, TenantID: deps.tenantID, BranchID: deps.branchID, Code: "C1", Name: "Counter 1", Status: counterEntity.CounterStatusActive}).Error)
	require.NoError(t, deps.db.Create(&settingsEntity.Setting{ID: uuid.New().String(), TenantID: deps.tenantID, ScopeType: settingsEntity.ScopeTypeService, ScopeID: deps.pharmacyServiceID, Key: settingsModel.SettingKeyPharmacyFlowEnabled, Value: "true", ValueType: "boolean", IsActive: true}).Error)
	require.NoError(t, deps.db.Create(&settingsEntity.Setting{ID: uuid.New().String(), TenantID: deps.tenantID, ScopeType: settingsEntity.ScopeTypeService, ScopeID: deps.pharmacyServiceID, Key: settingsModel.SettingKeyRequireCounterForService, Value: "true", ValueType: "boolean", IsActive: true}).Error)

	return deps
}

func TestQMSQueueIntegration(t *testing.T) {
	deps := setupQMSIntegration(t)

	ctx := database.SetOrganizationContext(context.Background(), deps.tenantID)
	ctx = database.SetBranchContext(ctx, deps.branchID)

	// Step 1: Initial state check
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, deps *qmsDeps, ctx context.Context)
	}{
		{
			name:     "Positive_RegisterAndForwardQueue",
			category: "positive",
			run: func(t *testing.T, deps *qmsDeps, ctx context.Context) {
				// Register
				queueRes, err := deps.queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: deps.regServiceID, PatientName: "John Queue"})
				require.NoError(t, err)
				require.NotNil(t, queueRes)
				assert.Equal(t, queueEntity.QueueStatusWaiting, queueRes.Status)

				// Forward
				forwarded, err := deps.queueMod.QueueUseCase.ForwardQueue(ctx, queueRes.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: deps.pharmacyServiceID, DestinationCounterID: deps.counterID})
				require.NoError(t, err)
				assert.Equal(t, queueRes.ID, forwarded.ID)
				assert.NotEqual(t, queueRes.CurrentJourneyID, forwarded.CurrentJourneyID)
			},
		},
		{
			name:     "Negative_ForwardWithoutRequiredCounterFails",
			category: "negative",
			run: func(t *testing.T, deps *qmsDeps, ctx context.Context) {
				queueRes, err := deps.queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: deps.regServiceID, PatientName: "Jane Queue"})
				require.NoError(t, err)

				_, err = deps.queueMod.QueueUseCase.ForwardQueue(ctx, queueRes.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: deps.pharmacyServiceID})
				assert.ErrorIs(t, err, exception.ErrForbidden) // Fails due to settings check requiring counter
			},
		},
		{
			name:     "Positive_TransitionsAndVisits",
			category: "positive",
			run: func(t *testing.T, deps *qmsDeps, ctx context.Context) {
				queueRes, err := deps.queueMod.QueueUseCase.RegisterQueue(ctx, &queueModel.RegisterQueueRequest{ServiceID: deps.regServiceID, PatientName: "Alice Queue"})
				require.NoError(t, err)

				_, err = deps.queueMod.QueueUseCase.ForwardQueue(ctx, queueRes.ID, &queueModel.ForwardQueueRequest{DestinationServiceID: deps.pharmacyServiceID, DestinationCounterID: deps.counterID})
				require.NoError(t, err)

				serving, err := deps.queueMod.QueueUseCase.TransitionQueue(ctx, queueRes.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
				require.NoError(t, err)
				assert.Equal(t, queueEntity.QueueStatusCalling, serving.Status)

				// Invalid transition
				_, err = deps.queueMod.QueueUseCase.TransitionQueue(ctx, queueRes.ID, &queueModel.QueueTransitionRequest{Action: queueModel.QueueActionCall})
				assert.ErrorIs(t, err, exception.ErrBadRequest)

				visits, err := deps.queueMod.QueueUseCase.GetVisitJourneys(ctx, queueRes.ID)
				require.NoError(t, err)
				assert.Len(t, visits, 3)
				assert.Equal(t, "registration", visits[0].EventType)
				assert.Equal(t, "forward", visits[1].EventType)
				assert.Equal(t, "call", visits[2].EventType)
			},
		},
		{
			name:     "Positive_StatsCountActiveJourneys",
			category: "positive",
			run: func(t *testing.T, deps *qmsDeps, ctx context.Context) {
				// Relies on data created by previous tests. 3 queues created total.
				stats, err := deps.queueMod.QueueUseCase.GetQueueStats(ctx)
				require.NoError(t, err)
				assert.GreaterOrEqual(t, stats.TotalQueuesToday, int64(3))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, deps, ctx)
		})
	}
}
