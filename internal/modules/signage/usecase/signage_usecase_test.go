package usecase

import (
	"context"
	"testing"

	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	"github.com/Roisfaozi/queue-base/internal/modules/signage/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newSignageTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE organizations (
			id TEXT PRIMARY KEY,
			running_text TEXT,
			logo_asset_id TEXT
		);
		CREATE TABLE qms_clients (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			branch_id TEXT,
			branch_service_id TEXT,
			counter_id TEXT,
			client_type TEXT,
			name TEXT,
			is_active BOOLEAN
		);
		CREATE TABLE branches (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			name TEXT,
			running_text TEXT,
			logo_asset_id TEXT
		);
		CREATE TABLE branch_services (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			service_id TEXT
		);
		CREATE TABLE services (
			id TEXT PRIMARY KEY,
			name TEXT,
			type TEXT,
			audio_id TEXT,
			audio_en TEXT
		);
		CREATE TABLE counters (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			display_name TEXT
		);
		CREATE TABLE queues (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			branch_id TEXT,
			queue_date TEXT,
			ticket_no TEXT,
			queue_no INTEGER,
			status TEXT,
			created_at INTEGER,
			updated_at INTEGER
		);
		CREATE TABLE queue_journeys (
			id TEXT PRIMARY KEY,
			queue_id TEXT,
			tenant_id TEXT,
			branch_id TEXT,
			service_id TEXT,
			counter_id TEXT,
			status TEXT,
			created_at INTEGER
		);
	`).Error)
	return db
}

func seedSignageTestData(t *testing.T, db *gorm.DB) {
	t.Helper()
	stmts := []string{
		`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, name, is_active) VALUES ('signage-1','t-1','b-1','bs-1','c-1','signage','Main Signage',1)`,
		`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, name, is_active) VALUES ('signage-2','t-1','b-1','bs-2',NULL,'signage','Service Signage',1)`,
		`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, name, is_active) VALUES ('signage-3','t-1','b-1',NULL,NULL,'signage','Lobby Signage',1)`,
		`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, name, is_active) VALUES ('signage-4','t-1','b-2',NULL,NULL,'signage','Fallback Signage',1)`,
		`INSERT INTO organizations (id, running_text, logo_asset_id) VALUES ('t-1','Tenant Welcome','tenant-logo')`,
		`INSERT INTO branches (id, tenant_id, name, running_text, logo_asset_id) VALUES ('b-1','t-1','Branch One','Welcome','logo-1')`,
		`INSERT INTO branches (id, tenant_id, name, running_text, logo_asset_id) VALUES ('b-2','t-1','Branch Two',NULL,NULL)`,
		`INSERT INTO branch_services (id, tenant_id, service_id) VALUES ('bs-1','t-1','svc-1')`,
		`INSERT INTO branch_services (id, tenant_id, service_id) VALUES ('bs-2','t-1','svc-2')`,
		`INSERT INTO services (id, name, type) VALUES ('svc-1','Customer Service','vip')`,
		`INSERT INTO services (id, name, type) VALUES ('svc-2','Pharmacy','general')`,
		`UPDATE services SET audio_id='audio-1', audio_en='audio-1-en' WHERE id='svc-1'`,
		`INSERT INTO counters (id, tenant_id, display_name) VALUES ('c-1','t-1','Counter 1')`,
		`INSERT INTO counters (id, tenant_id, display_name) VALUES ('c-2','t-1','Counter 2')`,
		`INSERT INTO queues (id, tenant_id, branch_id, queue_date, ticket_no, queue_no, status, created_at, updated_at) VALUES ('q-1','t-1','b-1','2026-07-03','A001',1,'calling',100,110)`,
		`INSERT INTO queues (id, tenant_id, branch_id, queue_date, ticket_no, queue_no, status, created_at, updated_at) VALUES ('q-2','t-1','b-1','2026-07-03','A002',2,'waiting',120,130)`,
		`INSERT INTO queues (id, tenant_id, branch_id, queue_date, ticket_no, queue_no, status, created_at, updated_at) VALUES ('q-3','t-1','b-1','2026-07-03','B001',3,'calling',140,150)`,
		`INSERT INTO queue_journeys (id, queue_id, tenant_id, branch_id, service_id, counter_id, status, created_at) VALUES ('j-1','q-1','t-1','b-1','svc-1','c-1','calling',100)`,
		`INSERT INTO queue_journeys (id, queue_id, tenant_id, branch_id, service_id, counter_id, status, created_at) VALUES ('j-2','q-2','t-1','b-1','svc-1','c-1','pending',120)`,
		`INSERT INTO queue_journeys (id, queue_id, tenant_id, branch_id, service_id, counter_id, status, created_at) VALUES ('j-3','q-3','t-1','b-1','svc-2','c-2','calling',140)`,
	}
	for _, stmt := range stmts {
		require.NoError(t, db.Exec(stmt).Error)
	}
}

func newSignageCtx() context.Context {
	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	return database.SetBranchContext(ctx, "b-1")
}

func newSignageCtxFor(tenantID, branchID string) context.Context {
	ctx := database.SetOrganizationContext(context.Background(), tenantID)
	return database.SetBranchContext(ctx, branchID)
}

func TestSignageUseCase_GetMe(t *testing.T) {
	db := newSignageTestDB(t)
	seedSignageTestData(t, db)
	uc := NewSignageUseCase(db, nil)

	tests := []struct {
		name          string
		clientID      string
		expectedErr   error
		expectedModel *model.SignageMeResponse
	}{
		{
			name:        "Negative_unauthorized when client missing",
			clientID:    "",
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:        "Negative_not found when client inactive or absent",
			clientID:    "missing",
			expectedErr: exception.ErrNotFound,
		},
		{
			name:     "Positive_returns full bound signage profile",
			clientID: "signage-1",
			expectedModel: &model.SignageMeResponse{
				ClientID:           "signage-1",
				TenantID:           "t-1",
				BranchID:           "b-1",
				BranchServiceID:    "bs-1",
				CounterID:          "c-1",
				ClientType:         "signage",
				Name:               "Main Signage",
				RunningText:        "Welcome",
				LogoAssetID:        "logo-1",
				BranchName:         "Branch One",
				ServiceName:        "Customer Service",
				CounterDisplayName: "Counter 1",
				AudioID:            "audio-1",
				AudioEN:            "audio-1-en",
			},
		},
		{
			name:     "Edge_falls back to tenant branding when branch empty",
			clientID: "signage-4",
			expectedModel: &model.SignageMeResponse{
				ClientID:    "signage-4",
				TenantID:    "t-1",
				BranchID:    "b-2",
				ClientType:  "signage",
				Name:        "Fallback Signage",
				RunningText: "Tenant Welcome",
				LogoAssetID: "tenant-logo",
				BranchName:  "Branch Two",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.GetMe(context.Background(), tt.clientID)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.expectedModel, res)
		})
	}
}

func TestSignageUseCase_GetCurrentCalls(t *testing.T) {
	db := newSignageTestDB(t)
	seedSignageTestData(t, db)
	uc := NewSignageUseCase(db, nil)

	tests := []struct {
		name         string
		ctx          context.Context
		clientID     string
		expectedErr  error
		expectedRows []model.SignageCurrentCallResponse
	}{
		{
			name:        "Negative_unauthorized when client missing",
			ctx:         newSignageCtx(),
			clientID:    "",
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:        "Negative_bad request when tenant or branch missing",
			ctx:         context.Background(),
			clientID:    "signage-1",
			expectedErr: exception.ErrBadRequest,
		},
		{
			name:        "Vulnerability_forbidden when context crosses client binding",
			ctx:         newSignageCtxFor("t-1", "b-2"),
			clientID:    "signage-1",
			expectedErr: exception.ErrForbidden,
		},
		{
			name:     "Positive_filters by branch service and counter binding",
			ctx:      newSignageCtx(),
			clientID: "signage-1",
			expectedRows: []model.SignageCurrentCallResponse{
				{
					QueueID:            "q-1",
					TicketNo:           "A001",
					CounterID:          "c-1",
					CounterDisplayName: "Counter 1",
					ServiceID:          "svc-1",
					ServiceType:        "vip",
					AudioID:            "audio-1",
					AudioEN:            "audio-1-en",
				},
			},
		},
		{
			name:     "Edge_filters only by branch service when counter absent",
			ctx:      newSignageCtx(),
			clientID: "signage-2",
			expectedRows: []model.SignageCurrentCallResponse{
				{
					QueueID:            "q-3",
					TicketNo:           "B001",
					CounterID:          "c-2",
					CounterDisplayName: "Counter 2",
					ServiceID:          "svc-2",
					ServiceType:        "general",
				},
			},
		},
		{
			name:     "Edge_returns all current calls when no service binding",
			ctx:      newSignageCtx(),
			clientID: "signage-3",
			expectedRows: []model.SignageCurrentCallResponse{
				{
					QueueID:            "q-3",
					TicketNo:           "B001",
					CounterID:          "c-2",
					CounterDisplayName: "Counter 2",
					ServiceID:          "svc-2",
					ServiceType:        "general",
				},
				{
					QueueID:            "q-1",
					TicketNo:           "A001",
					CounterID:          "c-1",
					CounterDisplayName: "Counter 1",
					ServiceID:          "svc-1",
					ServiceType:        "vip",
					AudioID:            "audio-1",
					AudioEN:            "audio-1-en",
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.GetCurrentCalls(tt.ctx, tt.clientID)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRows, res)
		})
	}
}

func TestSignageUseCase_GetQueues(t *testing.T) {
	db := newSignageTestDB(t)
	seedSignageTestData(t, db)
	uc := NewSignageUseCase(db, nil)

	tests := []struct {
		name         string
		ctx          context.Context
		clientID     string
		expectedErr  error
		expectedRows []queueModel.QueueResponse
	}{
		{
			name:        "Negative_unauthorized when client missing",
			ctx:         newSignageCtx(),
			clientID:    "",
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:        "Negative_bad request when tenant or branch missing",
			ctx:         context.Background(),
			clientID:    "signage-1",
			expectedErr: exception.ErrBadRequest,
		},
		{
			name:        "Vulnerability_forbidden when context crosses client binding",
			ctx:         newSignageCtxFor("t-1", "b-2"),
			clientID:    "signage-1",
			expectedErr: exception.ErrForbidden,
		},
		{
			name:     "Positive_filters pending queues by branch service",
			ctx:      newSignageCtx(),
			clientID: "signage-1",
			expectedRows: []queueModel.QueueResponse{
				{
					ID:        "q-2",
					TenantID:  "t-1",
					BranchID:  "b-1",
					QueueDate: "2026-07-03",
					TicketNo:  "A002",
					QueueNo:   2,
					Status:    "waiting",
					CreatedAt: 120,
					UpdatedAt: 130,
				},
			},
		},
		{
			name:         "Edge_returns empty when no pending queues for service binding",
			ctx:          newSignageCtx(),
			clientID:     "signage-2",
			expectedRows: []queueModel.QueueResponse{},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			res, err := uc.GetQueues(tt.ctx, tt.clientID)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expectedRows, res)
		})
	}
}
