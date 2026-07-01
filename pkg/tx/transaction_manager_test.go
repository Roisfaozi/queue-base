package tx_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/pkg/tx"
	"github.com/glebarez/sqlite" // Pure Go SQLite
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type NoOpWriter struct{}

func (w *NoOpWriter) Write([]byte) (int, error) {
	return 0, nil
}

func (w *NoOpWriter) Levels() []logrus.Level {
	return logrus.AllLevels
}

type User struct {
	ID   uint `gorm:"primaryKey"`
	Name string
}

func setupTestDB(t *testing.T) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&User{})
	require.NoError(t, err)
	return db, nil
}

func TestTransactionManager_WithinTransaction_Commit(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Within Transaction Commit",
			category: "positive",
			run: func(t *testing.T) {
				db, err := setupTestDB(t)
				assert.NoError(t, err)

				logger := logrus.New()
				logger.SetOutput(&NoOpWriter{})

				tm := tx.NewTransactionManager(db, logger)

				err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
					txDB, ok := tx.DBFromContext(ctx)
					assert.True(t, ok)

					if err := txDB.Create(&User{Name: "Alice"}).Error; err != nil {
						return err
					}
					return nil
				})

				assert.NoError(t, err)

				var count int64
				db.Model(&User{}).Count(&count)
				assert.Equal(t, int64(1), count)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTransactionManager_WithinTransaction_Rollback(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Within Transaction Rollback",
			category: "negative",
			run: func(t *testing.T) {
				db, err := setupTestDB(t)
				assert.NoError(t, err)

				logger := logrus.New()
				logger.SetOutput(&NoOpWriter{})

				tm := tx.NewTransactionManager(db, logger)

				err = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
					txDB, ok := tx.DBFromContext(ctx)
					assert.True(t, ok)

					txDB.Create(&User{Name: "Bob"})

					return errors.New("simulated error")
				})

				assert.Error(t, err)

				var count int64
				db.Model(&User{}).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestTransactionManager_WithinTransaction_PanicRollback(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Within Transaction Panic Rollback",
			category: "edge",
			run: func(t *testing.T) {
				db, err := setupTestDB(t)
				assert.NoError(t, err)

				logger := logrus.New()
				logger.SetOutput(&NoOpWriter{})

				tm := tx.NewTransactionManager(db, logger)

				assert.PanicsWithValue(t, "panic inside transaction", func() {
					_ = tm.WithinTransaction(context.Background(), func(ctx context.Context) error {
						panic("panic inside transaction")
					})
				})

				var count int64
				db.Model(&User{}).Count(&count)
				assert.Equal(t, int64(0), count)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
