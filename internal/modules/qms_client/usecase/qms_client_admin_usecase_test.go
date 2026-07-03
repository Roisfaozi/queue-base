package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/qms_client/model"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubAuditLogger struct {
	requests []auditModel.CreateAuditLogRequest
}

func (l *stubAuditLogger) LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error {
	l.requests = append(l.requests, req)
	return nil
}

func newAdminTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&entity.QMSClient{},
		&entity.QMSClientCredential{},
	))
	return db
}

func TestQMSClientAdminUseCase_CreateClient(t *testing.T) {
	db := newAdminTestDB(t)
	audit := &stubAuditLogger{}
	uc := NewQMSClientAdminUseCase(db, audit)

	ctx := database.SetOrganizationContext(context.Background(), "t-1")

	tests := []struct {
		name      string
		req       *model.QMSClientRequest
		tenantID  string
		wantError error
		wantAudit string
	}{
		{
			name:      "Positive_CreateCallerClient",
			req:       &model.QMSClientRequest{BranchID: "b-1", ClientType: "caller", Name: "Test Caller"},
			tenantID:  "t-1",
			wantError: nil,
			wantAudit: "QMS_CLIENT_CREATE",
		},
		{
			name:      "Negative_MissingTenant",
			req:       &model.QMSClientRequest{BranchID: "b-1", ClientType: "caller", Name: "Test Caller"},
			tenantID:  "",
			wantError: exception.ErrBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit.requests = nil
			testCtx := ctx
			if tt.tenantID == "" {
				testCtx = context.Background()
			}

			res, err := uc.CreateClient(testCtx, tt.req)

			if tt.wantError != nil {
				require.ErrorIs(t, err, tt.wantError)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.req.Name, res.Name)

			if tt.wantAudit != "" {
				require.Len(t, audit.requests, 1)
				assert.Equal(t, tt.wantAudit, audit.requests[0].Action)
				assert.Equal(t, res.ID, audit.requests[0].EntityID)
			}
		})
	}
}

func TestQMSClientAdminUseCase_CreateCredential(t *testing.T) {
	db := newAdminTestDB(t)
	audit := &stubAuditLogger{}
	uc := NewQMSClientAdminUseCase(db, audit)

	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	require.NoError(t, db.Create(&entity.QMSClient{ID: "c-1", TenantID: "t-1", BranchID: "b-1", ClientType: entity.ClientTypeCaller, Name: "A"}).Error)

	tests := []struct {
		name      string
		req       *model.QMSClientCredentialRequest
		tenantID  string
		wantError error
		wantAudit string
	}{
		{
			name:      "Positive_CreateCredential",
			req:       &model.QMSClientCredentialRequest{ClientID: "c-1", APIKey: "secret123"},
			tenantID:  "t-1",
			wantError: nil,
			wantAudit: "QMS_CLIENT_CREDENTIAL_CREATE",
		},
		{
			name:      "Negative_ClientNotFound",
			req:       &model.QMSClientCredentialRequest{ClientID: "c-missing", APIKey: "secret123"},
			tenantID:  "t-1",
			wantError: exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			audit.requests = nil
			testCtx := ctx
			if tt.tenantID == "" {
				testCtx = context.Background()
			}

			res, err := uc.CreateCredential(testCtx, tt.req)

			if tt.wantError != nil {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)

			// verify hash exists but differs from plain key
			var cred entity.QMSClientCredential
			require.NoError(t, db.First(&cred, "id = ?", res.ID).Error)
			assert.NotEqual(t, tt.req.APIKey, cred.ClientSecretHash)
			assert.True(t, pkg.CheckPasswordHash(tt.req.APIKey, cred.ClientSecretHash))

			if tt.wantAudit != "" {
				require.Len(t, audit.requests, 1)
				assert.Equal(t, tt.wantAudit, audit.requests[0].Action)
				// vulnerability check: audit should not log the secret
				if newVals, ok := audit.requests[0].NewValues.(map[string]string); ok {
					assert.NotContains(t, newVals, "client_secret_hash")
					assert.NotContains(t, newVals, "api_key")
				}
			}
		})
	}
}

func TestQMSClientAdminUseCase_UpdateAndDelete(t *testing.T) {
	db := newAdminTestDB(t)
	audit := &stubAuditLogger{}
	uc := NewQMSClientAdminUseCase(db, audit)
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	otherCtx := database.SetOrganizationContext(context.Background(), "t-2")
	require.NoError(t, db.Create(&entity.QMSClient{ID: "c-1", TenantID: "t-1", BranchID: "b-1", ClientType: entity.ClientTypeCaller, Name: "Old", IsActive: true}).Error)

	tests := []struct {
		name string
		run  func(t *testing.T)
	}{
		{
			name: "Positive_UpdateAndDeactivate",
			run: func(t *testing.T) {
				newName := "New"
				active := true
				res, err := uc.Update(ctx, "c-1", &model.QMSClientUpdateRequest{Name: newName, IsActive: &active})
				require.NoError(t, err)
				assert.Equal(t, newName, res.Name)

				require.NoError(t, uc.Delete(ctx, "c-1"))
				res, err = uc.GetByID(ctx, "c-1")
				require.NoError(t, err)
				assert.False(t, res.IsActive)
			},
		},
		{
			name: "Vulnerability_CrossTenantUpdateRejected",
			run: func(t *testing.T) {
				_, err := uc.Update(otherCtx, "c-1", &model.QMSClientUpdateRequest{Name: "Owned"})
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.run)
	}
}
