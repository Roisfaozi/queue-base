//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	qmsClientModulePkg "github.com/Roisfaozi/queue-base/internal/modules/qms_client"
	qmsClientModel "github.com/Roisfaozi/queue-base/internal/modules/qms_client/model"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type qmsClientDeps struct {
	db              *gorm.DB
	qmsClientMod    *qmsClientModulePkg.QMSClientModule
	tenantID        string
	branchID        string
	regServiceID    string
	branchServiceID string
	counterID       string
}

func setupQMSClientIntegration(t *testing.T) *qmsClientDeps {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	v := validator.New()
	log := env.Logger

	qmsClientMod := qmsClientModulePkg.NewQMSClientModule(env.DB, v, log)

	deps := &qmsClientDeps{
		db:              env.DB,
		qmsClientMod:    qmsClientMod,
		tenantID:        uuid.New().String(),
		branchID:        uuid.New().String(),
		regServiceID:    uuid.New().String(),
		branchServiceID: uuid.New().String(),
		counterID:       uuid.New().String(),
	}

	require.NoError(t, deps.db.Create(&orgEntity.Organization{ID: deps.tenantID, Code: "test-tenant-c-" + deps.tenantID[:6], Name: "TestTenantC", Slug: "test-tenant-c-" + deps.tenantID[:6], OwnerID: "system", Status: orgEntity.OrgStatusActive}).Error)
	require.NoError(t, deps.db.Create(&branchEntity.Branch{ID: deps.branchID, TenantID: deps.tenantID, Code: "BR1", Name: "Main Branch", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, deps.db.Create(&serviceEntity.Service{ID: deps.regServiceID, TenantID: deps.tenantID, Code: "RG", Name: "Registration", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, deps.db.Create(&serviceEntity.BranchService{ID: deps.branchServiceID, TenantID: deps.tenantID, BranchID: deps.branchID, ServiceID: deps.regServiceID, IsActive: true}).Error)
	require.NoError(t, deps.db.Create(&counterEntity.Counter{ID: deps.counterID, TenantID: deps.tenantID, BranchID: deps.branchID, Code: "C1", Name: "Counter 1", Status: counterEntity.CounterStatusActive}).Error)

	return deps
}

func TestQMSClientAdminIntegration(t *testing.T) {
	deps := setupQMSClientIntegration(t)

	ctx := database.SetOrganizationContext(context.Background(), deps.tenantID)
	ctx = database.SetBranchContext(ctx, deps.branchID)
	ctx = context.WithValue(ctx, "user_id", "admin-user-id")

	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, deps *qmsClientDeps, ctx context.Context)
	}{
		{
			name:     "Positive_CreateCallerClient",
			category: "positive",
			run: func(t *testing.T, deps *qmsClientDeps, ctx context.Context) {
				req := &qmsClientModel.QMSClientRequest{
					Name:       "Caller Int Client",
					ClientType: "caller",
					BranchID:   deps.branchID,
				}

				res, err := deps.qmsClientMod.AdminUseCase.CreateClient(ctx, req)
				require.NoError(t, err)
				assert.NotEmpty(t, res.ID)
				assert.Equal(t, "caller", res.ClientType)
				assert.Equal(t, "Caller Int Client", res.Name)
			},
		},
		{
			name:     "Positive_CreateSignageClient",
			category: "positive",
			run: func(t *testing.T, deps *qmsClientDeps, ctx context.Context) {
				req := &qmsClientModel.QMSClientRequest{
					Name:       "Signage Int Client",
					ClientType: "signage",
					BranchID:   deps.branchID,
				}

				res, err := deps.qmsClientMod.AdminUseCase.CreateClient(ctx, req)
				require.NoError(t, err)
				assert.NotEmpty(t, res.ID)
				assert.Equal(t, "signage", res.ClientType)
			},
		},
		{
			name:     "Negative_CreateWithInvalidClientType",
			category: "negative",
			run: func(t *testing.T, deps *qmsClientDeps, ctx context.Context) {
				req := &qmsClientModel.QMSClientRequest{
					Name:       "Bad Client",
					ClientType: "invalid",
					BranchID:   deps.branchID,
				}

				_, err := deps.qmsClientMod.AdminUseCase.CreateClient(ctx, req)
				require.Error(t, err)
			},
		},
		{
			name:     "Positive_CreateCredential",
			category: "positive",
			run: func(t *testing.T, deps *qmsClientDeps, ctx context.Context) {
				clientReq := &qmsClientModel.QMSClientRequest{
					Name:       "Cred Int Client",
					ClientType: "signage",
					BranchID:   deps.branchID,
				}

				clientRes, err := deps.qmsClientMod.AdminUseCase.CreateClient(ctx, clientReq)
				require.NoError(t, err)

				credReq := &qmsClientModel.QMSClientCredentialRequest{
					ClientID: clientRes.ID,
					APIKey:   "secretkey123",
				}

				credRes, err := deps.qmsClientMod.AdminUseCase.CreateCredential(ctx, credReq)
				require.NoError(t, err)
				assert.NotEmpty(t, credRes.ID)
				assert.Equal(t, clientRes.ID, credRes.ClientID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t, deps, ctx)
		})
	}
}
