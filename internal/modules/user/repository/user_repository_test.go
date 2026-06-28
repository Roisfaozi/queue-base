package repository_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/user/model"
	"github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/plugin/soft_delete"
)

func setupUserRepo(t *testing.T) (repository.UserRepository, *gorm.DB) {
	dbName := uuid.New().String()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)

	logrusLogger := logrus.New()
	logrusLogger.SetOutput(io.Discard)
	logrusLogger.SetLevel(logrus.FatalLevel)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	err = db.AutoMigrate(&entity.User{}, &entity.UserSSOIdentity{})
	require.NoError(t, err)
	err = db.Exec(`CREATE TABLE IF NOT EXISTS organization_members (organization_id TEXT, user_id TEXT, deleted_at INTEGER DEFAULT 0)`).Error
	require.NoError(t, err)

	repo := repository.NewUserRepository(db, logrusLogger)
	return repo, db
}

func simulateDBError(db *gorm.DB) {
	sqlDB, _ := db.DB()
	sqlDB.Close()
}

func TestUserRepository(t *testing.T) {
	ctx := context.Background()

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			setup    func(*gorm.DB) *entity.User
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(db *gorm.DB) *entity.User {
					return &entity.User{ID: "1", Username: "u1", Email: "u1@test.com"}
				},
			},
			{
				name:     "Negative_Duplicate",
				category: "negative",
				setup: func(db *gorm.DB) *entity.User {
					db.Create(&entity.User{ID: "2", Username: "dup", Email: "dup@test.com"})
					return &entity.User{ID: "3", Username: "dup", Email: "other@test.com"}
				},
				wantErr: true,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				setup: func(db *gorm.DB) *entity.User {
					simulateDBError(db)
					return &entity.User{ID: "4", Username: "u4", Email: "u4@test.com"}
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				user := tt.setup(db)

				err := repo.Create(ctx, user)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					saved, _ := repo.FindByID(ctx, user.ID)
					assert.Equal(t, user.Username, saved.Username)
				}
			})
		}
	})

	t.Run("FindByUsername", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			username string
			setup    func(*gorm.DB)
			wantErr  bool
		}{
			{
				name:     "Positive_Found",
				category: "positive",
				username: "findme",
				setup:    func(db *gorm.DB) { db.Create(&entity.User{ID: "1", Username: "findme", Email: "findme@t.com"}) },
			},
			{
				name:     "Negative_NotFound",
				category: "negative",
				username: "unknown",
				setup:    func(db *gorm.DB) {},
				wantErr:  true,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				username: "erruser",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				res, err := repo.FindByUsername(ctx, tt.username)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, res)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.username, res.Username)
				}
			})
		}
	})

	t.Run("FindByEmail", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			email    string
			setup    func(*gorm.DB)
			wantErr  bool
		}{
			{
				name:     "Positive_Found",
				category: "positive",
				email:    "f@t.com",
				setup:    func(db *gorm.DB) { db.Create(&entity.User{ID: "1", Email: "f@t.com", Username: "f"}) },
			},
			{
				name:     "Negative_NotFound",
				category: "negative",
				email:    "u@t.com",
				setup:    func(db *gorm.DB) {},
				wantErr:  true,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				email:    "e@t.com",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				res, err := repo.FindByEmail(ctx, tt.email)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, res)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.email, res.Email)
				}
			})
		}
	})

	t.Run("FindByToken", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			token    string
			setup    func(*gorm.DB)
			wantErr  bool
		}{
			{
				name:     "Positive_Found",
				category: "positive",
				token:    "tok",
				setup:    func(db *gorm.DB) { db.Create(&entity.User{ID: "1", Token: "tok", Email: "a@a.com", Username: "u1"}) },
			},
			{
				name:     "Negative_NotFound",
				category: "negative",
				token:    "inv",
				setup:    func(db *gorm.DB) {},
				wantErr:  true,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				token:    "err",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				res, err := repo.FindByToken(ctx, tt.token)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, res)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.token, res.Token)
				}
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			setup    func(*gorm.DB) *entity.User
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(db *gorm.DB) *entity.User {
					user := &entity.User{ID: "1", Name: "Old", Username: "u1", Email: "e1@a.com"}
					db.Create(user)
					user.Name = "New"
					return user
				},
			},
			{
				name:     "Negative_DuplicateUsername",
				category: "negative",
				setup: func(db *gorm.DB) *entity.User {
					db.Create(&entity.User{ID: "1", Username: "u1", Email: "e1@a.com"})
					db.Create(&entity.User{ID: "2", Username: "u2", Email: "e2@a.com"})
					return &entity.User{ID: "2", Username: "u1"}
				},
				wantErr: true,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				setup: func(db *gorm.DB) *entity.User {
					user := &entity.User{ID: "1", Username: "u1", Email: "e1@a.com"}
					simulateDBError(db)
					return user
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				user := tt.setup(db)
				err := repo.Update(ctx, user)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					saved, _ := repo.FindByID(ctx, user.ID)
					assert.Equal(t, user.Name, saved.Name)
				}
			})
		}
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			setup    func(*gorm.DB)
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(db *gorm.DB) {
					db.Create(&entity.User{ID: "1", Status: entity.UserStatusActive, Username: "u1", Email: "e1@a.com"})
				},
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				err := repo.UpdateStatus(ctx, "1", entity.UserStatusBanned)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					saved, _ := repo.FindByID(ctx, "1")
					assert.Equal(t, entity.UserStatusBanned, saved.Status)
				}
			})
		}
	})

	t.Run("Delete", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			setup    func(*gorm.DB)
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup:    func(db *gorm.DB) { db.Create(&entity.User{ID: "1", Username: "u1", Email: "e1@a.com"}) },
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				err := repo.Delete(ctx, "1")
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					_, err = repo.FindByID(ctx, "1")
					assert.Error(t, err) // soft deleted
					var cnt int64
					db.Unscoped().Model(&entity.User{}).Where("id = '1'").Count(&cnt)
					assert.Equal(t, int64(1), cnt)
				}
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			req      *model.GetUserListRequest
			setup    func(*gorm.DB, context.Context) context.Context
			wantLen  int
			wantTot  int64
			wantErr  bool
		}{
			{
				name:     "Positive_NoFilter",
				category: "positive",
				req:      &model.GetUserListRequest{Page: 1, Limit: 10},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					db.Create(&entity.User{ID: "1", Username: "a", Name: "a", Email: "a@a.com"})
					db.Create(&entity.User{ID: "2", Username: "b", Name: "b", Email: "b@a.com"})
					return c
				},
				wantLen: 2, wantTot: 2,
			},
			{
				name:     "Positive_FilterAndPagination",
				category: "positive",
				req:      &model.GetUserListRequest{Username: "a", Page: 1, Limit: 1},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					db.Create(&entity.User{ID: "1", Username: "a1", Name: "a1", Email: "a1@a.com"})
					db.Create(&entity.User{ID: "2", Username: "a2", Name: "a2", Email: "a2@a.com"})
					return c
				},
				wantLen: 1, wantTot: 2,
			},
			{
				name:     "Positive_OrgContext",
				category: "positive",
				req:      &model.GetUserListRequest{Page: 1, Limit: 10},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					db.Create(&entity.User{ID: "1", Username: "o1", Name: "o1", Email: "o1@a.com"})
					db.Create(&entity.User{ID: "2", Username: "o2", Name: "o2", Email: "o2@a.com"})
					db.Exec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('o1', '1')`)
					return database.SetOrganizationContext(c, "o1")
				},
				wantLen: 1, wantTot: 1,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				req:      &model.GetUserListRequest{Page: 1, Limit: 10},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					simulateDBError(db)
					return c
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				reqCtx := tt.setup(db, ctx)
				res, tot, err := repo.FindAll(reqCtx, tt.req)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, res, tt.wantLen)
					assert.Equal(t, tt.wantTot, tot)
				}
			})
		}
	})

	t.Run("FindAllDynamic", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			filter   *querybuilder.DynamicFilter
			setup    func(*gorm.DB, context.Context) context.Context
			wantLen  int
			wantTot  int64
			wantErr  bool
		}{
			{
				name:     "Positive_Contains",
				category: "positive",
				filter: &querybuilder.DynamicFilter{
					Filter: map[string]querybuilder.Filter{"Name": {Type: "contains", From: "a"}},
				},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					db.Create(&entity.User{ID: "1", Name: "a", Username: "a", Email: "a@a.com"})
					db.Create(&entity.User{ID: "2", Name: "b", Username: "b", Email: "b@a.com"})
					return c
				},
				wantLen: 1, wantTot: 1,
			},
			{
				name:     "Positive_OrgContext_SkipCount",
				category: "positive",
				filter:   &querybuilder.DynamicFilter{SkipCount: true, Page: 1, PageSize: 10},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					db.Create(&entity.User{ID: "1", Username: "1", Email: "1@a.com"})
					db.Create(&entity.User{ID: "2", Username: "2", Email: "2@a.com"})
					db.Exec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('o1', '1')`)
					return database.SetOrganizationContext(c, "o1")
				},
				wantLen: 1, wantTot: -1,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				filter:   &querybuilder.DynamicFilter{},
				setup: func(db *gorm.DB, c context.Context) context.Context {
					simulateDBError(db)
					return c
				},
				wantErr: true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				reqCtx := tt.setup(db, ctx)
				res, tot, err := repo.FindAllDynamic(reqCtx, tt.filter)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, res, tt.wantLen)
					assert.Equal(t, tt.wantTot, tot)
				}
			})
		}
	})

	t.Run("HardDeleteSoftDeletedUsers", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			setup    func(*gorm.DB)
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(db *gorm.DB) {
					old := time.Now().Add(-31 * 24 * time.Hour).UnixMilli()
					u := entity.User{ID: "1", Username: "u1", Email: "e1@a.com", DeletedAt: soft_delete.DeletedAt(old)}
					db.Create(&u)
				},
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				err := repo.HardDeleteSoftDeletedUsers(ctx, 30)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					var cnt int64
					db.Unscoped().Model(&entity.User{}).Count(&cnt)
					assert.Equal(t, int64(0), cnt)
				}
			})
		}
	})

	t.Run("GetByOrganization", func(t *testing.T) {
		tests := []struct {
			name     string
			category string
			org      string
			setup    func(*gorm.DB)
			wantLen  int
			wantErr  bool
		}{
			{
				name:     "Positive_Found",
				category: "positive",
				org:      "o1",
				setup: func(db *gorm.DB) {
					db.Create(&entity.User{ID: "1", Username: "1", Email: "1@a.com"})
					db.Exec(`INSERT INTO organization_members (organization_id, user_id) VALUES ('o1', '1')`)
				},
				wantLen: 1,
			},
			{
				name:     "Positive_Empty",
				category: "positive",
				org:      "o2",
				setup:    func(db *gorm.DB) {},
				wantLen:  0,
			},
			{
				name:     "Negative_DBError",
				category: "negative",
				org:      "o1",
				setup:    func(db *gorm.DB) { simulateDBError(db) },
				wantErr:  true,
			},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				repo, db := setupUserRepo(t)
				tt.setup(db)
				res, err := repo.GetByOrganization(ctx, tt.org)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, res, tt.wantLen)
				}
			})
		}
	})

	t.Run("SSOIdentity", func(t *testing.T) {
		t.Run("FindBySSOIdentity", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				setup    func(*gorm.DB)
				wantErr  bool
			}{
				{
					name:     "Positive_Found",
					category: "positive",
					setup: func(db *gorm.DB) {
						db.Create(&entity.UserSSOIdentity{ID: "1", Provider: "p", ProviderID: "pid"})
					},
				},
				{
					name:     "Negative_NotFound",
					category: "negative",
					setup:    func(db *gorm.DB) {},
					wantErr:  true,
				},
				{
					name:     "Negative_DBError",
					category: "negative",
					setup:    func(db *gorm.DB) { simulateDBError(db) },
					wantErr:  true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					repo, db := setupUserRepo(t)
					tt.setup(db)
					res, err := repo.FindBySSOIdentity(ctx, "p", "pid")
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.NotNil(t, res)
					}
				})
			}
		})

		t.Run("CreateSSOIdentity", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				setup    func(*gorm.DB)
				wantErr  bool
			}{
				{
					name:     "Positive_Success",
					category: "positive",
					setup:    func(db *gorm.DB) {},
				},
				{
					name:     "Negative_DBError",
					category: "negative",
					setup:    func(db *gorm.DB) { simulateDBError(db) },
					wantErr:  true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					repo, db := setupUserRepo(t)
					tt.setup(db)
					err := repo.CreateSSOIdentity(ctx, &entity.UserSSOIdentity{ID: "1", Provider: "p"})
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	})

	t.Run("TransactionContext", func(t *testing.T) {
		repo, db := setupUserRepo(t)
		tm := tx.NewTransactionManager(db, logrus.New())

		err := tm.WithinTransaction(ctx, func(txCtx context.Context) error {
			return repo.Create(txCtx, &entity.User{ID: "tx", Username: "tx", Email: "tx@a.com"})
		})
		require.NoError(t, err)
		saved, _ := repo.FindByID(ctx, "tx")
		assert.Equal(t, "tx", saved.Username)
	})
}
