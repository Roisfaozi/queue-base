package repository_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/auth/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/model"
	"github.com/Roisfaozi/queue-base/internal/modules/auth/repository"
	"github.com/Roisfaozi/queue-base/pkg/util"
	"github.com/glebarez/sqlite"
	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func getSessionKey(userID, sessionID string) string {
	return fmt.Sprintf("session:%s:%s", userID, sessionID)
}

func getSessionIndexKey(userID string) string {
	return fmt.Sprintf("session_index:%s", userID)
}

func getAttemptsKey(username string) string {
	return fmt.Sprintf("auth:attempts:%s", username)
}

func getLockedKey(username string) string {
	return fmt.Sprintf("auth:locked:%s", username)
}

func setupGormDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	assert.NoError(t, err)
	err = db.AutoMigrate(&entity.PasswordResetToken{}, &entity.EmailVerificationToken{})
	assert.NoError(t, err)
	return db
}

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) { return 0, nil }
func (w *NoOpWriter) Levels() []logrus.Level    { return logrus.AllLevels }

func TestTokenRepository(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(&NoOpWriter{})
	now := time.Now().Round(time.Second)
	mockClock := &util.MockClock{CurrentTime: now}

	t.Run("StoreToken", func(t *testing.T) {
		authData := &model.Auth{
			ID:           "session123",
			UserID:       "user456",
			RefreshToken: "some_refresh_token",
			ExpiresAt:    now.Add(time.Hour),
			CreatedAt:    now,
			UpdatedAt:    now,
		}
		key := getSessionKey(authData.UserID, authData.ID)
		val, _ := json.Marshal(authData)
		redisErr := errors.New("redis connection failed")

		tests := []struct {
			name     string
			category string
			setup    func(mock redismock.ClientMock)
			wantErr  error
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSet(key, val, time.Hour).SetVal("OK")
					mock.ExpectSAdd(getSessionIndexKey(authData.UserID), key).SetVal(1)
				},
			},
			{
				name:     "Negative_RedisError",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSet(key, val, time.Hour).SetErr(redisErr)
				},
				wantErr: redisErr,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db, mock := redismock.NewClientMock()
				repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
				tt.setup(mock)

				err := repo.StoreToken(context.Background(), authData)

				if tt.wantErr != nil {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.wantErr.Error())
				} else {
					assert.NoError(t, err)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("PasswordResetToken_GORM", func(t *testing.T) {
		token := &entity.PasswordResetToken{
			Email:     "test@example.com",
			Token:     "token123",
			ExpiresAt: time.Now().Add(time.Hour),
		}
		expiredToken := &entity.PasswordResetToken{
			Email:     "expired@example.com",
			Token:     "expired123",
			ExpiresAt: time.Now().Add(-1 * time.Hour),
		}

		t.Run("Save", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				token    *entity.PasswordResetToken
			}{
				{
					name:     "Positive_SaveToken",
					category: "positive",
					token:    token,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})
					err := repo.Save(context.Background(), tt.token)
					assert.NoError(t, err)

					var stored entity.PasswordResetToken
					err = db.First(&stored, "email = ?", tt.token.Email).Error
					assert.NoError(t, err)
					assert.Equal(t, tt.token.Token, stored.Token)
				})
			}
		})

		t.Run("FindByToken", func(t *testing.T) {
			tests := []struct {
				name        string
				category    string
				searchToken string
				setupDB     func(*gorm.DB)
				wantErr     bool
			}{
				{
					name:        "Positive_Found",
					category:    "positive",
					searchToken: token.Token,
					setupDB: func(db *gorm.DB) {
						db.Create(token)
					},
					wantErr: false,
				},
				{
					name:        "Negative_NotFound",
					category:    "negative",
					searchToken: "invalid",
					setupDB:     func(db *gorm.DB) {},
					wantErr:     true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					tt.setupDB(db)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

					result, err := repo.FindByToken(context.Background(), tt.searchToken)
					if tt.wantErr {
						assert.Error(t, err)
						assert.Nil(t, result)
					} else {
						assert.NoError(t, err)
						assert.NotNil(t, result)
					}
				})
			}
		})

		t.Run("DeleteByEmail", func(t *testing.T) {
			tests := []struct {
				name        string
				category    string
				searchEmail string
				setupDB     func(*gorm.DB)
				wantErr     bool
			}{
				{
					name:        "Positive_Deleted",
					category:    "positive",
					searchEmail: token.Email,
					setupDB: func(db *gorm.DB) {
						db.Create(token)
					},
					wantErr: false,
				},
				{
					name:        "Negative_DBError",
					category:    "negative",
					searchEmail: token.Email,
					setupDB: func(db *gorm.DB) {
						db.Exec("DROP TABLE IF EXISTS password_reset_tokens;")
					},
					wantErr: true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					tt.setupDB(db)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

					err := repo.DeleteByEmail(context.Background(), tt.searchEmail)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})

		t.Run("DeleteExpiredResetTokens", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				setupDB  func(*gorm.DB)
				wantErr  bool
			}{
				{
					name:     "Negative_SQLiteNoNOW",
					category: "negative",
					setupDB: func(db *gorm.DB) {
						db.Create(token)
						db.Create(expiredToken)
					},
					wantErr: true, // SQLite NOW() syntax error
				},
				{
					name:     "Negative_TableDrop",
					category: "negative",
					setupDB: func(db *gorm.DB) {
						db.Exec("DROP TABLE IF EXISTS password_reset_tokens;")
					},
					wantErr: true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					tt.setupDB(db)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

					err := repo.DeleteExpiredResetTokens(context.Background())
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	})

	t.Run("GetUserSessions", func(t *testing.T) {
		userID := "user123"
		indexKey := getSessionIndexKey(userID)
		keys := []string{getSessionKey(userID, "s1"), getSessionKey(userID, "s2")}
		s1 := model.Auth{ID: "s1", UserID: userID}
		s2 := model.Auth{ID: "s2", UserID: userID}
		json1, _ := json.Marshal(s1)
		json2, _ := json.Marshal(s2)

		tests := []struct {
			name     string
			category string
			setup    func(mock redismock.ClientMock)
			wantLen  int
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetVal(keys)
					mock.ExpectGet(keys[0]).SetVal(string(json1))
					mock.ExpectGet(keys[1]).SetVal(string(json2))
				},
				wantLen: 2,
			},
			{
				name:     "Negative_SMembersError",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetErr(errors.New("redis error"))
				},
				wantErr: true,
			},
			{
				name:     "Edge_GetErrorSkipsItem",
				category: "edge",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetVal(keys)
					mock.ExpectGet(keys[0]).SetErr(errors.New("get error"))
					mock.ExpectGet(keys[1]).SetVal(string(json2))
				},
				wantLen: 1,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db, mock := redismock.NewClientMock()
				repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
				tt.setup(mock)

				sessions, err := repo.GetUserSessions(context.Background(), userID)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
					assert.Len(t, sessions, tt.wantLen)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("RevokeAllSessions", func(t *testing.T) {
		userID := "user123"
		indexKey := getSessionIndexKey(userID)
		keys := []string{"k1", "k2"}

		tests := []struct {
			name     string
			category string
			setup    func(mock redismock.ClientMock)
			wantErr  bool
		}{
			{
				name:     "Positive_KeysFound",
				category: "positive",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetVal(keys)
					mock.ExpectDel(append(keys, indexKey)...).SetVal(3)
				},
			},
			{
				name:     "Positive_EmptyKeys",
				category: "positive",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetVal([]string{})
					mock.ExpectDel(indexKey).SetVal(1)
				},
			},
			{
				name:     "Negative_SMembersError",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetErr(errors.New("redis error"))
				},
				wantErr: true,
			},
			{
				name:     "Negative_DelError",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectSMembers(indexKey).SetVal(keys)
					mock.ExpectDel(append(keys, indexKey)...).SetErr(errors.New("del error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db, mock := redismock.NewClientMock()
				repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
				tt.setup(mock)

				err := repo.RevokeAllSessions(context.Background(), userID)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("GetToken", func(t *testing.T) {
		userID := "user456"
		sessionID := "session123"
		key := getSessionKey(userID, sessionID)
		expectedAuth := model.Auth{ID: sessionID, UserID: userID, RefreshToken: "expected_refresh_token"}
		jsonVal, _ := json.Marshal(expectedAuth)

		tests := []struct {
			name     string
			category string
			setup    func(mock redismock.ClientMock)
			wantRes  bool
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectGet(key).SetVal(string(jsonVal))
				},
				wantRes: true,
			},
			{
				name:     "Negative_NotFound",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectGet(key).SetErr(redis.Nil)
				},
			},
			{
				name:     "Negative_RedisError",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectGet(key).SetErr(errors.New("redis error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db, mock := redismock.NewClientMock()
				repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
				tt.setup(mock)

				res, err := repo.GetToken(context.Background(), userID, sessionID)
				if tt.wantErr {
					assert.Error(t, err)
					assert.Nil(t, res)
				} else {
					assert.NoError(t, err)
					if tt.wantRes {
						assert.NotNil(t, res)
					} else {
						assert.Nil(t, res)
					}
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("DeleteToken", func(t *testing.T) {
		userID := "user456"
		sessionID := "session123"
		key := getSessionKey(userID, sessionID)
		indexKey := getSessionIndexKey(userID)

		tests := []struct {
			name     string
			category string
			setup    func(mock redismock.ClientMock)
			wantErr  bool
		}{
			{
				name:     "Positive_Success",
				category: "positive",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectDel(key).SetVal(1)
					mock.ExpectSRem(indexKey, key).SetVal(1)
				},
			},
			{
				name:     "Negative_RedisError",
				category: "negative",
				setup: func(mock redismock.ClientMock) {
					mock.ExpectDel(key).SetErr(errors.New("redis error"))
				},
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				db, mock := redismock.NewClientMock()
				repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
				tt.setup(mock)

				err := repo.DeleteToken(context.Background(), userID, sessionID)
				if tt.wantErr {
					assert.Error(t, err)
				} else {
					assert.NoError(t, err)
				}
				assert.NoError(t, mock.ExpectationsWereMet())
			})
		}
	})

	t.Run("AccountLockout", func(t *testing.T) {
		username := "testuser"
		attemptsKey := getAttemptsKey(username)
		lockedKey := getLockedKey(username)

		t.Run("GetLoginAttempts", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				setup    func(mock redismock.ClientMock)
				want     int
				wantErr  bool
			}{
				{
					name:     "Positive_Found",
					category: "positive",
					setup:    func(mock redismock.ClientMock) { mock.ExpectGet(attemptsKey).SetVal("3") },
					want:     3,
				},
				{
					name:     "Edge_Nil",
					category: "edge",
					setup:    func(mock redismock.ClientMock) { mock.ExpectGet(attemptsKey).SetErr(redis.Nil) },
					want:     0,
				},
				{
					name:     "Negative_Error",
					category: "negative",
					setup:    func(mock redismock.ClientMock) { mock.ExpectGet(attemptsKey).SetErr(errors.New("err")) },
					want:     0,
					wantErr:  true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db, mock := redismock.NewClientMock()
					repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
					tt.setup(mock)
					res, err := repo.GetLoginAttempts(context.Background(), username)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tt.want, res)
					}
				})
			}
		})

		t.Run("IncrementLoginAttempts", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				setup    func(mock redismock.ClientMock)
				want     int
				wantErr  bool
			}{
				{
					name:     "Positive_Success",
					category: "positive",
					setup: func(mock redismock.ClientMock) {
						mock.ExpectIncr(attemptsKey).SetVal(1)
						mock.ExpectExpire(attemptsKey, time.Hour).SetVal(true)
					},
					want: 1,
				},
				{
					name:     "Negative_Error",
					category: "negative",
					setup: func(mock redismock.ClientMock) {
						mock.ExpectIncr(attemptsKey).SetErr(errors.New("err"))
					},
					want:    0,
					wantErr: true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db, mock := redismock.NewClientMock()
					repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
					tt.setup(mock)
					res, err := repo.IncrementLoginAttempts(context.Background(), username)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tt.want, res)
					}
				})
			}
		})

		t.Run("ResetLoginAttempts", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				setup    func(mock redismock.ClientMock)
				wantErr  bool
			}{
				{
					name:     "Positive_Success",
					category: "positive",
					setup:    func(mock redismock.ClientMock) { mock.ExpectDel(attemptsKey).SetVal(1) },
				},
				{
					name:     "Negative_Error",
					category: "negative",
					setup:    func(mock redismock.ClientMock) { mock.ExpectDel(attemptsKey).SetErr(errors.New("err")) },
					wantErr:  true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db, mock := redismock.NewClientMock()
					repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
					tt.setup(mock)
					err := repo.ResetLoginAttempts(context.Background(), username)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})

		t.Run("LockAccount", func(t *testing.T) {
			duration := 30 * time.Minute
			tests := []struct {
				name     string
				category string
				setup    func(mock redismock.ClientMock)
				wantErr  bool
			}{
				{
					name:     "Positive_Success",
					category: "positive",
					setup:    func(mock redismock.ClientMock) { mock.ExpectSet(lockedKey, "locked", duration).SetVal("OK") },
				},
				{
					name:     "Negative_Error",
					category: "negative",
					setup: func(mock redismock.ClientMock) {
						mock.ExpectSet(lockedKey, "locked", duration).SetErr(errors.New("err"))
					},
					wantErr: true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db, mock := redismock.NewClientMock()
					repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
					tt.setup(mock)
					err := repo.LockAccount(context.Background(), username, duration)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})

		t.Run("IsAccountLocked", func(t *testing.T) {
			tests := []struct {
				name       string
				category   string
				setup      func(mock redismock.ClientMock)
				wantLocked bool
				wantTTL    time.Duration
				wantErr    bool
			}{
				{
					name:       "Positive_Locked",
					category:   "positive",
					setup:      func(mock redismock.ClientMock) { mock.ExpectTTL(lockedKey).SetVal(time.Minute) },
					wantLocked: true,
					wantTTL:    time.Minute,
				},
				{
					name:       "Positive_NotLocked",
					category:   "positive",
					setup:      func(mock redismock.ClientMock) { mock.ExpectTTL(lockedKey).SetVal(-2 * time.Second) },
					wantLocked: false,
					wantTTL:    0,
				},
				{
					name:       "Negative_Error",
					category:   "negative",
					setup:      func(mock redismock.ClientMock) { mock.ExpectTTL(lockedKey).SetErr(errors.New("err")) },
					wantLocked: false,
					wantTTL:    0,
					wantErr:    true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db, mock := redismock.NewClientMock()
					repo := repository.NewTokenRepositoryRedis(db, logger, nil, mockClock)
					tt.setup(mock)
					locked, ttl, err := repo.IsAccountLocked(context.Background(), username)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, tt.wantLocked, locked)
						assert.Equal(t, tt.wantTTL, ttl)
					}
				})
			}
		})
	})

	t.Run("EmailVerificationToken_GORM", func(t *testing.T) {
		token := &entity.EmailVerificationToken{
			Email:     "verify@example.com",
			Token:     "verify123",
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		}

		t.Run("SaveVerificationToken", func(t *testing.T) {
			tests := []struct {
				name     string
				category string
				token    *entity.EmailVerificationToken
			}{
				{
					name:     "Positive_Save",
					category: "positive",
					token:    token,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})
					err := repo.SaveVerificationToken(context.Background(), tt.token)
					assert.NoError(t, err)

					var stored entity.EmailVerificationToken
					err = db.First(&stored, "email = ?", tt.token.Email).Error
					assert.NoError(t, err)
					assert.Equal(t, tt.token.Token, stored.Token)
				})
			}
		})

		t.Run("FindVerificationToken", func(t *testing.T) {
			tests := []struct {
				name        string
				category    string
				searchToken string
				setupDB     func(*gorm.DB)
				wantErr     bool
			}{
				{
					name:        "Positive_Found",
					category:    "positive",
					searchToken: token.Token,
					setupDB: func(db *gorm.DB) {
						db.Create(token)
					},
					wantErr: false,
				},
				{
					name:        "Negative_NotFound",
					category:    "negative",
					searchToken: "invalid",
					setupDB:     func(db *gorm.DB) {},
					wantErr:     true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					tt.setupDB(db)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

					result, err := repo.FindVerificationToken(context.Background(), tt.searchToken)
					if tt.wantErr {
						assert.Error(t, err)
						assert.Nil(t, result)
					} else {
						assert.NoError(t, err)
						assert.NotNil(t, result)
					}
				})
			}
		})

		t.Run("DeleteVerificationTokenByEmail", func(t *testing.T) {
			tests := []struct {
				name        string
				category    string
				searchEmail string
				setupDB     func(*gorm.DB)
				wantErr     bool
			}{
				{
					name:        "Positive_Deleted",
					category:    "positive",
					searchEmail: token.Email,
					setupDB: func(db *gorm.DB) {
						db.Create(token)
					},
					wantErr: false,
				},
				{
					name:        "Negative_DBError",
					category:    "negative",
					searchEmail: token.Email,
					setupDB: func(db *gorm.DB) {
						db.Exec("DROP TABLE IF EXISTS email_verification_tokens;")
					},
					wantErr: true,
				},
			}
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					db := setupGormDB(t)
					tt.setupDB(db)
					repo := repository.NewTokenRepositoryRedis(nil, logger, db, &util.RealClock{})

					err := repo.DeleteVerificationTokenByEmail(context.Background(), tt.searchEmail)
					if tt.wantErr {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
					}
				})
			}
		})
	})
}
