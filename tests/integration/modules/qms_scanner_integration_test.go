//go:build integration
// +build integration

package modules

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	apiKeyModulePkg "github.com/Roisfaozi/queue-base/internal/modules/api_key"
	counterModulePkg "github.com/Roisfaozi/queue-base/internal/modules/counter"
	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	branchModulePkg "github.com/Roisfaozi/queue-base/internal/modules/organization"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	queueModulePkg "github.com/Roisfaozi/queue-base/internal/modules/queue"
	scannerModulePkg "github.com/Roisfaozi/queue-base/internal/modules/scanner"
	scannerModel "github.com/Roisfaozi/queue-base/internal/modules/scanner/model"
	scannerUsecase "github.com/Roisfaozi/queue-base/internal/modules/scanner/usecase"
	serviceModulePkg "github.com/Roisfaozi/queue-base/internal/modules/service"
	serviceEntity "github.com/Roisfaozi/queue-base/internal/modules/service/entity"
	settingsModulePkg "github.com/Roisfaozi/queue-base/internal/modules/settings"
	settingsEntity "github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type scannerDeps struct {
	db                *gorm.DB
	scannerMod        *scannerModulePkg.ScannerModule
	tenantID          string
	branchID          string
	regServiceID      string
	pharmacyServiceID string
	counterID         string
	apiKey            string
	clientID          string
	otherTenantID     string
}

func setupScannerIntegration(t *testing.T) *scannerDeps {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}

	v := validator.New()
	settingsMod := settingsModulePkg.NewSettingsModule(env.DB, v)
	queueMod := queueModulePkg.NewQueueModule(env.DB, v, settingsMod.QueueSettingsResolver, env.Logger)
	branchMod := branchModulePkg.NewBranchModule(env.DB, v, env.Logger)
	serviceMod := serviceModulePkg.NewServiceModule(env.DB, v)
	counterMod := counterModulePkg.NewCounterModule(env.DB, v, branchMod.BranchRepo)

	deps := &scannerDeps{
		db:                env.DB,
		tenantID:          uuid.New().String(),
		branchID:          uuid.New().String(),
		regServiceID:      uuid.New().String(),
		pharmacyServiceID: uuid.New().String(),
		counterID:         uuid.New().String(),
		apiKey:            "test-api-key-" + uuid.New().String()[:8],
		clientID:          "scanner-client-" + uuid.New().String()[:8],
		otherTenantID:     uuid.New().String(),
	}

	// Create tenant organizations and branches
	require.NoError(t, deps.db.Create(&orgEntity.Organization{ID: deps.tenantID, Name: "TestTenant", Slug: "test-tenant-" + deps.tenantID[:6], OwnerID: "system", Status: orgEntity.OrgStatusActive}).Error)
	require.NoError(t, deps.db.Create(&branchEntity.Branch{ID: deps.branchID, TenantID: deps.tenantID, Code: "BR1", Name: "Main Branch", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, deps.db.Create(&orgEntity.Organization{ID: deps.otherTenantID, Name: "OtherTenant", Slug: "other-tenant-" + deps.otherTenantID[:6], OwnerID: "system", Status: orgEntity.OrgStatusActive}).Error)

	// Create services
	require.NoError(t, deps.db.Create(&serviceEntity.Service{ID: deps.regServiceID, TenantID: deps.tenantID, Code: "RG", Name: "Registration", Status: serviceEntity.ServiceStatusActive}).Error)
	require.NoError(t, deps.db.Create(&serviceEntity.Service{ID: deps.pharmacyServiceID, TenantID: deps.tenantID, Code: "PH", Name: "Pharmacy", Status: serviceEntity.ServiceStatusActive, IsPharmacy: true}).Error)

	// Create counter
	require.NoError(t, deps.db.Create(&counterEntity.Counter{ID: deps.counterID, TenantID: deps.tenantID, BranchID: deps.branchID, Code: "C1", Name: "Counter 1", Status: counterEntity.CounterStatusActive}).Error)

	// Create pharmacy settings
	require.NoError(t, deps.db.Create(&settingsEntity.Setting{ID: uuid.New().String(), TenantID: deps.tenantID, ScopeType: settingsEntity.ScopeTypeService, ScopeID: deps.pharmacyServiceID, Key: settingsModel.SettingKeyPharmacyFlowEnabled, Value: "true", ValueType: "boolean", IsActive: true}).Error)
	require.NoError(t, deps.db.Create(&settingsEntity.Setting{ID: uuid.New().String(), TenantID: deps.tenantID, ScopeType: settingsEntity.ScopeTypeService, ScopeID: deps.pharmacyServiceID, Key: settingsModel.SettingKeyRequireCounterForService, Value: "true", ValueType: "boolean", IsActive: true}).Error)

	// Create API key for scanner auth
	apiKeyHash := sha256.Sum256([]byte(deps.apiKey))
	apiKeyHashHex := hex.EncodeToString(apiKeyHash[:])

	// First we need a user to own the API key
	require.NoError(t, deps.db.Create(&userEntity.User{
		ID:       deps.clientID,
		Username: "scanner-user",
		Email:    deps.clientID + "@example.com",
		Password: "password",
		Status:   userEntity.UserStatusActive,
	}).Error)

	require.NoError(t, deps.db.Exec("INSERT INTO api_keys (id, key_hash, organization_id, user_id, name, scopes, is_active) VALUES (?, ?, ?, ?, ?, ?, ?)",
		uuid.New().String(), apiKeyHashHex, deps.tenantID, deps.clientID, "Scanner Test Key", `["*"]`, true).Error)

	// Wire authenticator
	apiKeyMod := apiKeyModulePkg.NewApiKeyModule(env.DB, userRepository.NewUserRepository(env.DB, env.Logger), env.Redis, env.Logger, v)
	authenticator := scannerModulePkg.NewAPIKeyAuthenticator(apiKeyMod.UseCase)

	deps.scannerMod = scannerModulePkg.NewScannerModule(queueMod, branchMod, serviceMod, counterMod, settingsMod, v, authenticator, env.Logger)

	return deps
}

func TestQMSScannerIntegration(t *testing.T) {
	deps := setupScannerIntegration(t)

	ctx := database.SetOrganizationContext(context.Background(), deps.tenantID)
	ctx = database.SetBranchContext(ctx, deps.branchID)

	tests := []struct {
		name     string
		category string
		prepare  func(t *testing.T, ctx context.Context) context.Context
		req      scannerModel.CheckInRequest
		wantErr  error
		wantRes  func(t *testing.T, res *scannerUsecase.CheckInResponse)
	}{
		{
			name:     "Positive_RegisterViaScanner",
			category: "positive",
			req: scannerModel.CheckInRequest{
				Action:      scannerUsecase.ActionRegister,
				BranchID:    deps.branchID,
				ServiceID:   deps.regServiceID,
				PatientName: "John Scanner",
			},
			wantRes: func(t *testing.T, res *scannerUsecase.CheckInResponse) {
				assert.Equal(t, scannerUsecase.ActionRegister, res.Action)
				assert.NotNil(t, res.Queue)
				assert.NotEmpty(t, res.Queue.ID)
				assert.Equal(t, "John Scanner", res.Queue.PatientName)
			},
		},
		{
			name:     "Positive_ForwardViaScanner",
			category: "positive",
			prepare: func(t *testing.T, ctx context.Context) context.Context {
				// First register a queue, then forward it
				res, err := deps.scannerMod.ScannerUseCase.CheckIn(ctx, &scannerUsecase.CheckInRequest{
					Action:      scannerUsecase.ActionRegister,
					ClientID:    deps.clientID,
					APIKey:      deps.apiKey,
					BranchID:    deps.branchID,
					ServiceID:   deps.regServiceID,
					PatientName: "Forward Me",
				})
				require.NoError(t, err)
				// Store queue ID on context for the forward test
				return context.WithValue(ctx, "queueID", res.Queue.ID)
			},
			req: scannerModel.CheckInRequest{
				Action:               scannerUsecase.ActionForward,
				BranchID:             deps.branchID,
				DestinationServiceID: deps.pharmacyServiceID,
				DestinationCounterID: deps.counterID,
			},
			wantRes: func(t *testing.T, res *scannerUsecase.CheckInResponse) {
				assert.Equal(t, scannerUsecase.ActionForward, res.Action)
				assert.NotNil(t, res.Queue)
				assert.NotEmpty(t, res.Queue.ID)
			},
		},
		{
			name:     "Negative_InvalidAPIKey",
			category: "negative",
			req: scannerModel.CheckInRequest{
				Action:      scannerUsecase.ActionRegister,
				BranchID:    deps.branchID,
				ServiceID:   deps.regServiceID,
				PatientName: "Bad Auth",
			},
			prepare: func(t *testing.T, ctx context.Context) context.Context {
				// Override API key to a bad one for this test
				return context.WithValue(ctx, "apiKey", "invalid-key")
			},
			wantErr: exception.ErrUnauthorized,
		},
		{
			name:     "Negative_MissingBranchID",
			category: "negative",
			req: scannerModel.CheckInRequest{
				Action:      scannerUsecase.ActionRegister,
				ServiceID:   deps.regServiceID,
				PatientName: "No Branch",
			},
			prepare: func(t *testing.T, ctx context.Context) context.Context {
				return database.SetBranchContext(ctx, "")
			},
			wantErr: exception.ErrBadRequest,
		},
		{
			name:     "Vulnerability_CrossTenantAccessRejected",
			category: "vulnerability",
			prepare: func(t *testing.T, ctx context.Context) context.Context {
				// Switch to a different tenant's context but use same credentials
				return database.SetOrganizationContext(database.SetBranchContext(context.Background(), deps.branchID), deps.otherTenantID)
			},
			req: scannerModel.CheckInRequest{
				Action:      scannerUsecase.ActionRegister,
				BranchID:    deps.branchID,
				ServiceID:   deps.regServiceID,
				PatientName: "Cross Tenant",
			},
			wantErr: exception.ErrUnauthorized,
		},
		{
			name:     "Vulnerability_MismatchedBranchRequestRejected",
			category: "vulnerability",
			req: scannerModel.CheckInRequest{
				Action:      scannerUsecase.ActionRegister,
				BranchID:    uuid.New().String(),
				ServiceID:   deps.regServiceID,
				PatientName: "Wrong Branch",
			},
			wantErr: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testCtx := ctx
			if tt.prepare != nil {
				testCtx = tt.prepare(t, ctx)
			}

			// Determine API key and client ID for this test case
			apiKey := deps.apiKey
			clientID := deps.clientID
			if overrideKey, ok := testCtx.Value("apiKey").(string); ok {
				apiKey = overrideKey
			}
			// For the forward test, patch in the QueueID from prepare
			queueID := tt.req.QueueID
			if queueID == "" {
				if id, ok := testCtx.Value("queueID").(string); ok {
					queueID = id
				}
			}

			res, err := deps.scannerMod.ScannerUseCase.CheckIn(testCtx, &scannerUsecase.CheckInRequest{
				Action:               tt.req.Action,
				ClientID:             clientID,
				APIKey:               apiKey,
				BranchID:             tt.req.BranchID,
				ServiceID:            tt.req.ServiceID,
				PatientName:          tt.req.PatientName,
				QueueID:              queueID,
				DestinationServiceID: tt.req.DestinationServiceID,
				DestinationCounterID: tt.req.DestinationCounterID,
			})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err, "failed when calling checkIn %v", err)
			if tt.wantRes != nil {
				tt.wantRes(t, res)
			}
		})
	}
}
