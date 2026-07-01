package test_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type auditTestDeps struct {
	Repo        *mocks.MockAuditRepository
	MockWS      *mocks.MockWebSocketManager
	Distributor *mocks.MockTaskDistributor
}

func setupAuditTest() (*auditTestDeps, usecase.AuditUseCase) {
	deps := &auditTestDeps{
		Repo:        new(mocks.MockAuditRepository),
		MockWS:      new(mocks.MockWebSocketManager),
		Distributor: new(mocks.MockTaskDistributor),
	}
	logger := logrus.New()
	logger.SetOutput(io.Discard)

	// Default mock behavior
	deps.MockWS.On("BroadcastToChannel", mock.Anything, mock.Anything).Return()

	uc := usecase.NewAuditUseCase(deps.Repo, logger, deps.MockWS, deps.Distributor)
	return deps, uc
}

func TestLogActivity(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				req := model.CreateAuditLogRequest{
					UserID: "u1", Action: "CREATE", Entity: "User", EntityID: "u2",
					OldValues: map[string]string{"foo": "bar"},
				}

				deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AuditLog) bool {
					return log.UserID == "u1" && log.Action == "CREATE" && log.OldValues != ""
				})).Return(nil)

				err := uc.LogActivity(context.Background(), req)
				assert.NoError(t, err)
				deps.Repo.AssertExpectations(t)
			},
		},
		{
			name:     "Positive_TransactionalPath_WriteToOutbox",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				req := model.CreateAuditLogRequest{
					UserID: "u1", Action: "UPDATE", Entity: "Profile", EntityID: "u1",
				}

				db, mockSQL, _ := sqlmock.New()
				gormDB, _ := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
				tm := tx.NewTransactionManager(gormDB, logrus.New())

				mockSQL.ExpectBegin()
				mockSQL.ExpectCommit()

				err := tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
					deps.Repo.On("CreateOutbox", ctx, mock.MatchedBy(func(outbox *entity.AuditOutbox) bool {
						return outbox.UserID == "u1" && outbox.Action == "UPDATE"
					})).Return(nil)

					return uc.LogActivity(ctx, req)
				})

				assert.NoError(t, err)
				deps.Repo.AssertExpectations(t)
				deps.Repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			},
		},
		{
			name:     "Positive_OrganizationContext_CapturesOrgID",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				orgID := "org-999"
				ctx := database.SetOrganizationContext(context.Background(), orgID)
				req := model.CreateAuditLogRequest{
					UserID: "u1", Action: "LOGIN", Entity: "Auth", EntityID: "s1",
				}

				deps.Repo.On("Create", ctx, mock.MatchedBy(func(log *entity.AuditLog) bool {
					return log.OrganizationID != nil && *log.OrganizationID == orgID
				})).Return(nil)

				err := uc.LogActivity(ctx, req)
				assert.NoError(t, err)
				deps.Repo.AssertExpectations(t)
			},
		},
		{
			name:     "Edge_NilJSONValues",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				req := model.CreateAuditLogRequest{
					UserID: "u1", Action: "DELETE", Entity: "User", EntityID: "u2",
					OldValues: nil,
					NewValues: nil,
				}

				deps.Repo.On("Create", mock.Anything, mock.MatchedBy(func(log *entity.AuditLog) bool {
					return log.OldValues == "null" && log.NewValues == "null"
				})).Return(nil)

				err := uc.LogActivity(context.Background(), req)
				assert.NoError(t, err)
			},
		},
		{
			name:     "Negative_RepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				req := model.CreateAuditLogRequest{UserID: "u1"}
				deps.Repo.On("Create", mock.Anything, mock.Anything).Return(errors.New("db error"))

				err := uc.LogActivity(context.Background(), req)
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestGetLogsDynamic(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Positive_Success",
			category: "positive",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				now := time.Now().UnixMilli()
				entities := []*entity.AuditLog{
					{ID: "1", UserID: "u1", OldValues: `{"a":1}`, NewValues: `{"a":2}`, CreatedAt: now},
				}

				filter := &querybuilder.DynamicFilter{}
				deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(entities, int64(1), nil)

				res, total, err := uc.GetLogsDynamic(context.Background(), filter)
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, int64(1), total)
				assert.Equal(t, "u1", res[0].UserID)

				oldVal := res[0].OldValues.(map[string]interface{})
				assert.Equal(t, float64(1), oldVal["a"])
			},
		},
		{
			name:     "Edge_MalformedJSONInDB",
			category: "edge",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				entities := []*entity.AuditLog{
					{ID: "1", UserID: "u1", OldValues: `{broken_json`, NewValues: `null`},
				}

				filter := &querybuilder.DynamicFilter{}
				deps.Repo.On("FindAllDynamic", mock.Anything, filter).Return(entities, int64(1), nil)

				res, total, err := uc.GetLogsDynamic(context.Background(), filter)
				assert.NoError(t, err)
				assert.Len(t, res, 1)
				assert.Equal(t, int64(1), total)
				assert.Nil(t, res[0].OldValues)
			},
		},
		{
			name:     "Negative_RepoError",
			category: "negative",
			run: func(t *testing.T) {
				deps, uc := setupAuditTest()
				deps.Repo.On("FindAllDynamic", mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db fail"))

				res, total, err := uc.GetLogsDynamic(context.Background(), nil)
				assert.Error(t, err)
				assert.Nil(t, res)
				assert.Equal(t, int64(0), total)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
