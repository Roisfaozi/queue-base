package usecase

import (
	"context"
	"testing"
	"time"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	authModel "github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/caller/model"
	queueModel "github.com/Roisfaozi/queue-base/internal/modules/queue/model"
	queueUseCasePkg "github.com/Roisfaozi/queue-base/internal/modules/queue/usecase"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/jwt"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubQueueUC struct {
	res     *queueModel.QueueResponse
	err     error
	ctx     context.Context
	queueID string
	action  string
}

func (s *stubQueueUC) SetEventBroadcaster(events queueUseCasePkg.EventBroadcaster) {}

func (s *stubQueueUC) SetWSBroadcaster(ws queueUseCasePkg.WSBroadcaster) {}
func (s *stubQueueUC) RegisterQueue(ctx context.Context, req *queueModel.RegisterQueueRequest) (*queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) ListQueues(ctx context.Context, req queueModel.ListQueuesRequest) ([]queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) GetQueueByID(ctx context.Context, queueID string) (*queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) ForwardQueue(ctx context.Context, queueID string, req *queueModel.ForwardQueueRequest) (*queueModel.QueueResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) TransitionQueue(ctx context.Context, queueID string, req *queueModel.QueueTransitionRequest) (*queueModel.QueueResponse, error) {
	s.ctx = ctx
	s.queueID = queueID
	s.action = req.Action
	return s.res, s.err
}
func (s *stubQueueUC) ListActiveJourneys(ctx context.Context, req queueModel.QueueJourneyListRequest) ([]queueModel.QueueJourneyResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) GetVisitJourneys(ctx context.Context, queueID string) ([]queueModel.VisitJourneyResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) GetQueueStats(ctx context.Context) (*queueModel.QueueStatsResponse, error) {
	return nil, nil
}
func (s *stubQueueUC) ResolveQueueBranchID(ctx context.Context, queueID string) (string, error) {
	return "", nil
}

type stubAuthUC struct {
	loginRes     *authModel.LoginResponse
	refreshToken string
	loginErr     error
	loginRequest authModel.LoginRequest
}

func (s *stubAuthUC) GenerateAccessToken(user *userEntity.User) (string, error)       { return "", nil }
func (s *stubAuthUC) GenerateRefreshToken(user *userEntity.User) (string, error)      { return "", nil }
func (s *stubAuthUC) ValidateAccessToken(token string) (*jwt.Claims, error)           { return nil, nil }
func (s *stubAuthUC) ValidateRefreshToken(token string) (*jwt.Claims, error)          { return nil, nil }
func (s *stubAuthUC) RevokeToken(ctx context.Context, userID, sessionID string) error { return nil }
func (s *stubAuthUC) Register(ctx context.Context, request authModel.RegisterRequest) (*authModel.LoginResponse, string, error) {
	return nil, "", nil
}
func (s *stubAuthUC) Login(ctx context.Context, request authModel.LoginRequest) (*authModel.LoginResponse, string, error) {
	s.loginRequest = request
	return s.loginRes, s.refreshToken, s.loginErr
}
func (s *stubAuthUC) RefreshToken(ctx context.Context, refreshToken string) (*authModel.TokenResponse, string, error) {
	return nil, "", nil
}
func (s *stubAuthUC) Verify(ctx context.Context, userID string, sessionID string) (*authModel.Auth, error) {
	return nil, nil
}
func (s *stubAuthUC) GetUserSessions(ctx context.Context, userID string) ([]*authModel.Auth, error) {
	return nil, nil
}
func (s *stubAuthUC) RevokeAllSessions(ctx context.Context, userID string) error         { return nil }
func (s *stubAuthUC) ForgotPassword(ctx context.Context, email string) error             { return nil }
func (s *stubAuthUC) ResetPassword(ctx context.Context, token, newPassword string) error { return nil }
func (s *stubAuthUC) RequestVerification(ctx context.Context, userID string) error       { return nil }
func (s *stubAuthUC) VerifyEmail(ctx context.Context, token string) error                { return nil }
func (s *stubAuthUC) GetTicket(ctx context.Context, userContext authModel.UserSessionContext) (string, error) {
	return "", nil
}
func (s *stubAuthUC) GetSSORedirectURL(ctx context.Context, provider string, state string) (string, error) {
	return "", nil
}
func (s *stubAuthUC) HandleSSOCallback(ctx context.Context, provider string, code string) (*authModel.LoginResponse, string, error) {
	return nil, "", nil
}

type stubAuditUC struct {
	requests []auditModel.CreateAuditLogRequest
	err      error
}

func (s *stubAuditUC) LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error {
	s.requests = append(s.requests, req)
	return s.err
}
func (s *stubAuditUC) GetLogsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]auditModel.AuditLogResponse, int64, error) {
	return nil, 0, nil
}
func (s *stubAuditUC) ExportLogs(ctx context.Context, fromDate, toDate string, process func([]auditModel.AuditLogResponse) error) error {
	return nil
}
func (s *stubAuditUC) ExportLogsAsync(ctx context.Context, userID, orgID, fromDate, toDate, format string) error {
	return nil
}

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
		clientID     string
		request      model.CallerLoginRequest
		authUC       *stubAuthUC
		seed         func(t *testing.T, db *gorm.DB)
		expectedErr  error
		assertResult func(t *testing.T, res *model.CallerLoginResponse, auth *stubAuthUC)
	}{
		{
			name:     "PositiveReturnsCallerContext",
			clientID: "client-1",
			request:  model.CallerLoginRequest{Username: "caller1", Password: "secret123"},
			authUC:   &stubAuthUC{loginRes: &authModel.LoginResponse{AccessToken: "jwt-token", ExpiresAt: now, User: authModel.UserInfo{ID: "u-1", Username: "caller1"}}, refreshToken: "refresh-token"},
			assertResult: func(t *testing.T, res *model.CallerLoginResponse, auth *stubAuthUC) {
				require.NotNil(t, res)
				assert.Equal(t, "jwt-token", res.AccessToken)
				assert.Equal(t, "t-1", res.Context.TenantID)
				assert.Equal(t, "Branch One", res.Context.BranchName)
				assert.Equal(t, "Registration", res.Context.ServiceName)
				assert.Equal(t, "c-1", res.Context.CounterID)
				assert.Equal(t, "caller1", auth.loginRequest.Username)
			},
		},
		{
			name:        "NegativeRejectsMissingClient",
			clientID:    "",
			request:     model.CallerLoginRequest{Username: "caller1", Password: "secret123"},
			authUC:      &stubAuthUC{},
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:     "EdgeAllowsClientWithoutCounterBinding",
			clientID: "client-2",
			request:  model.CallerLoginRequest{Username: "caller2", Password: "secret123"},
			authUC:   &stubAuthUC{loginRes: &authModel.LoginResponse{AccessToken: "jwt-token", ExpiresAt: now, User: authModel.UserInfo{ID: "u-2", Username: "caller2"}}, refreshToken: "refresh-token"},
			seed: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(`INSERT INTO qms_clients (id, tenant_id, branch_id, branch_service_id, counter_id, client_type, is_active) VALUES ('client-2','t-1','b-1','bs-1',NULL,'caller',1)`).Error)
				require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-2','active',0)`).Error)
			},
			assertResult: func(t *testing.T, res *model.CallerLoginResponse, auth *stubAuthUC) {
				assert.Equal(t, "", res.Context.CounterID)
			},
		},
		{
			name:     "VulnerabilityRejectsUserWithoutAssignment",
			clientID: "client-1",
			request:  model.CallerLoginRequest{Username: "caller3", Password: "secret123"},
			authUC:   &stubAuthUC{loginRes: &authModel.LoginResponse{AccessToken: "jwt-token", ExpiresAt: now, User: authModel.UserInfo{ID: "u-3", Username: "caller3"}}, refreshToken: "refresh-token"},
			seed: func(t *testing.T, db *gorm.DB) {
				require.NoError(t, db.Exec(`INSERT INTO organization_members (organization_id, user_id, status, deleted_at) VALUES ('t-1','u-3','active',0)`).Error)
			},
			expectedErr: exception.ErrForbidden,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.seed != nil {
				tt.seed(t, db)
			}
			uc := NewCallerUseCase(db, &stubQueueUC{}, tt.authUC, &stubAuditUC{})
			res, _, err := uc.Login(context.Background(), tt.clientID, tt.request)
			if tt.expectedErr != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}
			require.NoError(t, err)
			tt.assertResult(t, res, tt.authUC)
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
		clientID    string
		ctx         context.Context
		expectedErr error
	}{
		{name: "PositiveReturnsCallerContext", clientID: "client-1", ctx: authcontext.WithUserID(context.Background(), "u-1")},
		{name: "NegativeRejectsMissingUser", clientID: "client-1", ctx: context.Background(), expectedErr: exception.ErrUnauthorized},
		{name: "EdgeRejectsMissingClient", clientID: "", ctx: authcontext.WithUserID(context.Background(), "u-1"), expectedErr: exception.ErrUnauthorized},
		{name: "VulnerabilityRejectsCrossUser", clientID: "client-1", ctx: authcontext.WithUserID(context.Background(), "u-2"), expectedErr: exception.ErrForbidden},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			uc := NewCallerUseCase(db, &stubQueueUC{}, &stubAuthUC{}, &stubAuditUC{})
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
		clientID    string
		userID      string
		journeyID   string
		action      string
		expectedErr error
		assertAudit func(t *testing.T, audit *stubAuditUC)
	}{
		{
			name:      "PositiveSuccess",
			clientID:  "client-1",
			userID:    "u-1",
			journeyID: "j-1",
			action:    queueModel.QueueActionCall,
			assertAudit: func(t *testing.T, audit *stubAuditUC) {
				require.Len(t, audit.requests, 1)
				assert.Equal(t, "CALLER_CALL", audit.requests[0].Action)
			},
		},
		{
			name:        "NegativeUnauthorizedWhenClientMissing",
			clientID:    "",
			journeyID:   "j-1",
			action:      queueModel.QueueActionCall,
			expectedErr: exception.ErrUnauthorized,
		},
		{
			name:        "EdgeForbiddenWhenCounterMismatch",
			clientID:    "client-2",
			journeyID:   "j-1",
			action:      queueModel.QueueActionCall,
			expectedErr: exception.ErrForbidden,
		},
		{
			name:        "VulnerabilityForbiddenWhenOperatorAssignmentMissing",
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
			queueUC := &stubQueueUC{res: &queueModel.QueueResponse{ID: "q-1", TicketNo: "A001", QueueNo: 1, Status: "calling"}}
			auditUC := &stubAuditUC{}
			uc := NewCallerUseCase(db, queueUC, &stubAuthUC{}, auditUC)
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
			assert.Equal(t, "q-1", queueUC.queueID)
			assert.Equal(t, tt.action, queueUC.action)
			if tt.assertAudit != nil {
				tt.assertAudit(t, auditUC)
			}
		})
	}
}
