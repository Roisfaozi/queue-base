//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	auditRepo "github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUsecase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	counterModule "github.com/Roisfaozi/queue-base/internal/modules/counter"
	counterModel "github.com/Roisfaozi/queue-base/internal/modules/counter/model"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	branchRepoPkg "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	serviceModule "github.com/Roisfaozi/queue-base/internal/modules/service"
	serviceModel "github.com/Roisfaozi/queue-base/internal/modules/service/model"
	settingsModule "github.com/Roisfaozi/queue-base/internal/modules/settings"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQMSAuditIntegration_CRUDVisibility(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}
	defer env.Cleanup()

	setup.CleanupDatabase(t, env.DB)

	validate := validator.New()
	auditUC := auditUsecase.NewAuditUseCase(auditRepo.NewAuditRepository(env.DB, env.Logger), env.Logger, nil, nil)
	branchRepo := branchRepoPkg.NewBranchRepository(env.DB)
	serviceMod := serviceModule.NewServiceModule(env.DB, validate, branchRepo, env.Logger, auditUC)
	counterMod := counterModule.NewCounterModule(env.DB, validate, branchRepo, serviceMod.BranchServiceRepo, env.Logger, auditUC)
	settingsMod := settingsModule.NewSettingsModule(env.DB, validate, env.Logger, auditUC)

	tenantID := uuid.New().String()
	branchID := uuid.New().String()
	ctx := database.SetOrganizationContext(context.Background(), tenantID)

	require.NoError(t, env.DB.Create(&branchEntity.Branch{ID: branchID, TenantID: tenantID, Code: "BRA", Name: "Branch Audit", Status: branchEntity.BranchStatusActive}).Error)

	svc, err := serviceMod.ServiceUseCase.CreateService(ctx, &serviceModel.CreateServiceRequest{Code: "SVA", Name: "Service Audit"})
	require.NoError(t, err)

	counter, err := counterMod.CounterUseCase.CreateCounter(ctx, &counterModel.CreateCounterRequest{BranchID: branchID, Code: "CTA", Name: "Counter Audit"})
	require.NoError(t, err)

	branchService, err := serviceMod.BranchServiceUseCase.CreateBranchService(ctx, branchID, &serviceModel.CreateBranchServiceRequest{ServiceID: svc.ID})
	require.NoError(t, err)

	_, err = serviceMod.BranchServiceUseCase.UpdateBranchService(ctx, branchID, branchService.ID, &serviceModel.UpdateBranchServiceRequest{CustomName: strPtr("Front Desk")})
	require.NoError(t, err)

	err = serviceMod.BranchServiceUseCase.DeleteBranchService(ctx, branchID, branchService.ID)
	require.NoError(t, err)

	setting, err := settingsMod.SettingsUseCase.CreateSetting(ctx, &settingsModel.CreateSettingRequest{ScopeType: "tenant", Key: "queue_prefix", Value: "A", ValueType: "string"})
	require.NoError(t, err)
	_, err = settingsMod.SettingsUseCase.UpdateSetting(ctx, setting.ID, &settingsModel.UpdateSettingRequest{Value: strPtr("B")})
	require.NoError(t, err)
	err = settingsMod.SettingsUseCase.DeleteSetting(ctx, setting.ID)
	require.NoError(t, err)

	logs, _, err := auditUC.GetLogsDynamic(ctx, &querybuilder.DynamicFilter{Sort: &[]querybuilder.SortModel{{ColId: "CreatedAt", Sort: "asc"}}})
	require.NoError(t, err)

	body := make([]string, 0, len(logs))
	for _, item := range logs {
		body = append(body, item.Action)
	}

	assert.Contains(t, body, "SERVICE_CREATE")
	assert.Contains(t, body, "COUNTER_CREATE")
	assert.Contains(t, body, "BRANCH_SERVICE_CREATE")
	assert.Contains(t, body, "BRANCH_SERVICE_UPDATE")
	assert.Contains(t, body, "BRANCH_SERVICE_DELETE")
	assert.Contains(t, body, "SETTING_CREATE")
	assert.Contains(t, body, "SETTING_UPDATE")
	assert.Contains(t, body, "SETTING_DELETE")
	assert.NotEmpty(t, counter.ID)
}

func strPtr(v string) *string { return &v }
