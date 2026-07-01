package test_test

import (
	"context"
	"io"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/audit/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	"github.com/glebarez/sqlite"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)
	err = db.AutoMigrate(&entity.AuditLog{})
	require.NoError(t, err)
	return db
}

func TestAuditRepository_Create(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, repo auditUseCase.AuditRepository, db *gorm.DB)
	}{
		{
			name:     "Success - Create Full Audit Log",
			category: "unit",
			run: func(t *testing.T, repo auditUseCase.AuditRepository, db *gorm.DB) {
				ctx := context.Background()

				oldVal := `{"name": "Old Name"}`
				newVal := `{"name": "New Name"}`

				log := &entity.AuditLog{
					UserID:    "user-123",
					Action:    "UPDATE",
					Entity:    "User",
					EntityID:  "target-user-id",
					OldValues: oldVal,
					NewValues: newVal,
					IPAddress: "127.0.0.1",
					UserAgent: "Mozilla/5.0",
				}

				err := repo.Create(ctx, log)
				assert.NoError(t, err)
				assert.NotEmpty(t, log.ID, "ID should be auto-generated")
				assert.NotZero(t, log.CreatedAt, "CreatedAt should be set")

				var storedLog entity.AuditLog
				err = db.First(&storedLog, "id = ?", log.ID).Error
				assert.NoError(t, err)
				assert.Equal(t, "user-123", storedLog.UserID)
				assert.Equal(t, oldVal, storedLog.OldValues)
				assert.Equal(t, "UPDATE", storedLog.Action)
			},
		},
		{
			name:     "Success - Create Partial Log (No JSON values)",
			category: "unit",
			run: func(t *testing.T, repo auditUseCase.AuditRepository, db *gorm.DB) {
				ctx := context.Background()
				log := &entity.AuditLog{
					UserID:   "user-456",
					Action:   "LOGIN",
					Entity:   "Auth",
					EntityID: "session-id",
				}

				err := repo.Create(ctx, log)
				assert.NoError(t, err)

				var storedLog entity.AuditLog
				db.First(&storedLog, "id = ?", log.ID)
				assert.Empty(t, storedLog.OldValues)
				assert.Equal(t, "LOGIN", storedLog.Action)
			},
		},
		{
			name:     "Edge - JSON Injection Safety",
			category: "edge",
			run: func(t *testing.T, repo auditUseCase.AuditRepository, db *gorm.DB) {
				ctx := context.Background()
				malformedJSON := `{"key": "value" -- broken query'; DROP TABLE users; --`

				log := &entity.AuditLog{
					UserID:    "hacker",
					Action:    "HACK",
					Entity:    "System",
					EntityID:  "1",
					OldValues: malformedJSON,
				}

				err := repo.Create(ctx, log)
				assert.NoError(t, err)

				var storedLog entity.AuditLog
				db.First(&storedLog, "id = ?", log.ID)
				assert.Equal(t, malformedJSON, storedLog.OldValues)
			},
		},
		{
			name:     "Edge - Very Long UserAgent",
			category: "edge",
			run: func(t *testing.T, repo auditUseCase.AuditRepository, db *gorm.DB) {
				ctx := context.Background()
				longUA := ""
				for i := 0; i < 300; i++ {
					longUA += "a"
				}

				log := &entity.AuditLog{
					UserID:    "user-long",
					Action:    "TEST",
					Entity:    "Test",
					EntityID:  "1",
					UserAgent: longUA,
				}

				err := repo.Create(ctx, log)
				assert.NoError(t, err)

				var storedLog entity.AuditLog
				db.First(&storedLog, "id = ?", log.ID)
				assert.Equal(t, longUA, storedLog.UserAgent)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			log := logrus.New()
			log.SetOutput(io.Discard)
			repo := repository.NewAuditRepository(db, log)
			tt.run(t, repo, db)
		})
	}
}
