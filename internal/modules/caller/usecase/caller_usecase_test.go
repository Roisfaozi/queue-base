package usecase

import (
	"context"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	auditMocks "github.com/Roisfaozi/queue-base/internal/modules/audit/test/mocks"
	authModel "github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	authMocks "github.com/Roisfaozi/queue-base/internal/modules/auth/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/caller/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueMocks "github.com/Roisfaozi/queue-base/internal/modules/queue/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newCallerTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.Exec(`
		CREATE TABLE organizations (id TEXT PRIMARY KEY, name TEXT);
		CREATE TABLE branches (id TEXT PRIMARY KEY, tenant_id TEXT, name TEXT, running_text TEXT, logo_asset_id TEXT);
		CREATE TABLE services (id TEXT PRIMARY KEY, name TEXT, type TEXT);
		CREATE TABLE branch_services (id TEXT PRIMARY KEY, tenant_id TEXT, service_id TEXT);
		CREATE TABLE counters (id TEXT PRIMARY KEY, tenant_id TEXT, branch_id TEXT, name TEXT, display_name TEXT);
		CREATE TABLE queue_journeys (
			id TEXT PRIMARY KEY,
			queue_id TEXT,
			tenant_id TEXT,
			branch_id TEXT,
			counter_id TEXT
		);
		CREATE TABLE qms_clients (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			branch_id TEXT,
			branch_service_id TEXT,
			counter_id TEXT,
			client_type TEXT,
			is_active BOOLEAN
		);
		CREATE TABLE operator_counter_assignments (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			branch_id TEXT,
			user_id TEXT,
			counter_id TEXT,
			unassigned_at INTEGER
		);
		CREATE TABLE organization_members (
			organization_id TEXT,
			user_id TEXT,
			status TEXT,
			deleted_at INTEGER DEFAULT 0
		);
	`).Error)
	require.NoError(t, db.Exec(`INSERT INTO organizations (id, name) VALUES ('t-1','Tenant One')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO branches (id, tenant_id, name) VALUES ('b-1','t-1','Branch One')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO services (id, name, type) VALUES ('s-1','Registration','registration')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO branch_services (id, tenant_id, service_id) VALUES ('bs-1','t-1','s-1')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO counters (id, tenant_id, branch_id, name, display_name) VALUES ('c-1','t-1','b-1','Counter 1','Loket 1')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO counters (id, tenant_id, branch_id, name, display_name) VALUES ('c-2','t-1','b-1','Counter 2','Loket 2')`).Error)
	return db
}

func TestCallerUseCase_Login(t *testing.T) {
	db := newCallerTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, is_active) VALUES ('client-1','t-1','b-1','bs-1','c-1','caller',1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-1','active',0)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO operator_counter_assignments (id, tenant_id, branch_id, user_id, counter_id, unassigned_at) VALUES ('a-1','t-1','b-1','u-1','c-1',NULL)`).Error)

	now := time.Now().Add(time.Hour)
	tests := []struct {
		name         string
		category     string
		clientID     string
		request      model.CallerLoginRequest
		setupAuth    func(t *testing.T) *authMocks.MockAuthUseCase
		seed         func(t *testing.T, db *gorm.DB)
		expectedErr  error
		assertResult func(t *testing.T, res *model.CallerLoginResponse)
	}{
		{
			name:     "PositiveReturnsCallerContext",
			category: "positive",
			clientID: "client-1",
			request:  model.CallerLoginRequest{Username: "caller1", Password: "secret123"},
			setupAuth: func(t *testing.T) *authMocks.MockAuthUseCase {
				m := authMocks.NewMockAuthUseCase(t)
				m.On("Login", mock.Anything, authModel.LoginRequest{Username: "caller1", Password: "secret123"}).
					Return(&authModel.LoginResponse{AccessToken: "jwt-token", ExpiresAt: now, User: authModel.UserInfo{ID: "u-1", Username: "caller1"}}, "refresh-token", nil).Once()
				return m
			},
			assertResult: func(t *testing.T, res *model.CallerLoginResponse) {
				require.NotNil(t, res)
				assert.Equal(t, "jwt-token", res.AccessToken)
				assert.Equal(t, "t-1", res.Context.TenantID)
				assert.Equal(t, "Branch One", res.Context.BranchName)
				assert.Equal(t, "Registration", res.Context.ServiceName)
				assert.Equal(t, "c-1", res.Context.CounterID)
			},
		},
		{
			name:     "NegativeRejectsMissingClient",
			category: "negative",
			clientID: "",
			request:  model.CallerLoginRequest{Username: "caller1", Password: "secret123"},
			setupAuth: func(t *testing.T) *authMocks.MockAuthUseCase {
				return authMocks.NewMockAuthUseCase(t)
			},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "EdgeAllowsClientWithoutCounterBinding",
			category: "edge",
			clientID: "client-2",
			request:  model.CallerLoginRequest{Username: "caller2", Password: "secret123"},
			setupAuth: func(t *testing.T) *authMocks.MockAuthUseCase {
				m := authMocks.NewMockAuthUseCase(t)
				m.On("Login", mock.Anything, authModel.LoginRequest{Username: "caller2", Password: "secret123"}).
					Return(&authModel.LoginResponse{AccessToken: "jwt-token", ExpiresAt: now, User: authModel.UserInfo{ID: "u-2", Username: "caller2"}}, "refresh-token", nil).Once()
				return m
			},
			seed: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, is_active) VALUES ('client-2','t-1','b-1','bs-1',NULL,'caller',1)`).Error)
				require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-2','active',0)`).Error)
			},
			assertResult: func(t *testing.T, res *model.CallerLoginResponse) {
				assert.Equal(t, "", res.Context.CounterID)
			},
		},
		{
			name:     "VulnerabilityRejectsUserWithoutAssignment",
			category: "vulnerability",
			clientID: "client-1",
			request:  model.CallerLoginRequest{Username: "caller3", Password: "secret123"},
			setupAuth: func(t *testing.T) *authMocks.MockAuthUseCase {
				m := authMocks.NewMockAuthUseCase(t)
				m.On("Login", mock.Anything, authModel.LoginRequest{Username: "caller3", Password: "secret123"}).
					Return(&authModel.LoginResponse{AccessToken: "jwt-token", ExpiresAt: now, User: authModel.UserInfo{ID: "u-3", Username: "caller3"}}, "refresh-token", nil).Once()
				return m
			},
			seed: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-3','active',0)`).Error)
			},
			expectedErr: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			if tt.seed != nil {
				tt.seed(t, db)
			}
			authUC := tt.setupAuth(t)
			queueUC := queueMocks.NewMockQueueUseCase(t)
			auditUC := auditMocks.NewMockAuditUseCase(t)

			uc := NewCallerUseCase(db, queueUC, authUC, auditUC)
			res, _, err := uc.Login(context.Background(), tt.clientID, tt.request)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			tt.assertResult(t, res)
			authUC.AssertExpectations(t)
		})
	}
}

func TestCallerUseCase_Me(t *testing.T) {
	db := newCallerTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, is_active) VALUES ('client-1','t-1','b-1','bs-1','c-1','caller',1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-1','active',0)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO operator_counter_assignments (id, tenant_id, branch_id, user_id, counter_id, unassigned_at) VALUES ('a-1','t-1','b-1','u-1','c-1',NULL)`).Error)

	tests := []struct {
		name        string
		category    string
		clientID    string
		ctx         context.Context
		expectedErr error
	}{
		{name: "PositiveReturnsCallerContext", category: "positive", clientID: "client-1", ctx: authcontext.WithUserID(context.Background(), "u-1")},
		{name: "NegativeRejectsMissingUser", category: "negative", clientID: "client-1", ctx: context.Background(), expectedErr: exception.ErrUnauthorized},
		{name: "EdgeRejectsMissingClient", category: "edge", clientID: "", ctx: authcontext.WithUserID(context.Background(), "u-1"), expectedErr: exception.ErrUnauthorized},
		{name: "VulnerabilityRejectsCrossUser", category: "vulnerability", clientID: "client-1", ctx: authcontext.WithUserID(context.Background(), "u-2"), expectedErr: exception.ErrForbidden},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			queueUC := queueMocks.NewMockQueueUseCase(t)
			authUC := authMocks.NewMockAuthUseCase(t)
			auditUC := auditMocks.NewMockAuditUseCase(t)

			uc := NewCallerUseCase(db, queueUC, authUC, auditUC)
			res, err := uc.Me(tt.ctx, tt.clientID)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, "t-1", res.Context.TenantID)
			assert.Equal(t, "Loket 1", res.Context.CounterDisplayName)
		})
	}
}

func TestCallerUseCase_ExecuteAction(t *testing.T) {
	db := newCallerTestDB(t)
	require.NoError(t, db.Exec(`INSERT INTO queue_journeys (id, queue_id, tenant_id, branch_id, counter_id) VALUES ('j-1','q-1','t-1','b-1','c-1')`).Error)
	require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, is_active) VALUES ('client-1','t-1','b-1','bs-1','c-1','caller',1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, is_active) VALUES ('client-2','t-1','b-1','bs-1','c-2','caller',1)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-1','active',0)`).Error)
	require.NoError(t, db.Exec(`INSERT INTO operator_counter_assignments (id, tenant_id, branch_id, user_id, counter_id, unassigned_at) VALUES ('a-1','t-1','b-1','u-1','c-1',NULL)`).Error)

	tests := []struct {
		name        string
		category    string
		clientID    string
		userID      string
		journeyID   string
		action      string
		expectedErr error
		expectAudit bool
	}{
		{
			name:        "PositiveSuccess",
			category:    "positive",
			clientID:    "client-1",
			userID:      "u-1",
			journeyID:   "j-1",
			action:      queueModel.QueueActionCall,
			expectAudit: true,
		},
		{
			name:        "NegativeUnauthorizedWhenClientMissing",
			category:    "negative",
			clientID:    "",
			journeyID:   "j-1",
			action:      queueModel.QueueActionCall,
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:        "EdgeForbiddenWhenCounterMismatch",
			category:    "edge",
			clientID:    "client-2",
			journeyID:   "j-1",
			action:      queueModel.QueueActionCall,
			expectedErr: exception.ErrForbidden,
		},
		{
			name:        "VulnerabilityForbiddenWhenOperatorAssignmentMissing",
			category:    "vulnerability",
			clientID:    "client-1",
			userID:      "u-2",
			journeyID:   "j-1",
			action:      queueModel.QueueActionCall,
			expectedErr: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			queueUC := queueMocks.NewMockQueueUseCase(t)
			auditUC := auditMocks.NewMockAuditUseCase(t)
			authUC := authMocks.NewMockAuthUseCase(t)

			if tt.expectAudit {
				queueUC.On("TransitionQueue", mock.Anything, "q-1", mock.MatchedBy(func(req *queueModel.QueueTransitionRequest) bool {
					return req != nil && req.Action == tt.action
				})).Return(&queueModel.QueueResponse{ID: "q-1", TicketNo: "A001", QueueNo: 1, Status: "calling"}, nil).Once()

				auditUC.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
					return req.Action == "CALLER_CALL"
				})).Return(nil).Once()
			}

			uc := NewCallerUseCase(db, queueUC, authUC, auditUC)
			ctx := database.SetOrganizationContext(context.Background(), "t-1")
			ctx = database.SetBranchContext(ctx, "b-1")
			if tt.userID != "" {
				ctx = authcontext.WithUserID(ctx, tt.userID)
			}

			res, err := uc.ExecuteAction(ctx, tt.clientID, tt.journeyID, tt.action)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Equal(t, tt.journeyID, res.JourneyID)
			queueUC.AssertExpectations(t)
			auditUC.AssertExpectations(t)
		})
	}
}
