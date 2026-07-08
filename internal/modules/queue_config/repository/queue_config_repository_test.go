package repository

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/queue_config/entity"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&entity.TenantQueueSetting{},
		&entity.BranchQueueSetting{},
		&entity.BranchServiceQueueSetting{},
		&entity.CounterQueueSetting{},
	))
	return db
}

func TestQueueConfigRepository_Upsert(t *testing.T) {
	db := newTestDB(t)
	repo := NewQueueConfigRepository(db)
	ctx := context.Background()

	t.Run("UpsertTenantQueueSetting", func(t *testing.T) {
		tenantID := "tenant-1"
		err := repo.UpsertTenantQueueSetting(ctx, tenantID, map[string]any{"default_estimated_duration": 10})
		require.NoError(t, err)

		var row entity.TenantQueueSetting
		require.NoError(t, db.First(&row, "tenant_id = ?", tenantID).Error)
		assert.Equal(t, 10, row.DefaultEstimatedDuration)

		// Update existing
		err = repo.UpsertTenantQueueSetting(ctx, tenantID, map[string]any{"default_estimated_duration": 15})
		require.NoError(t, err)
		require.NoError(t, db.First(&row, "tenant_id = ?", tenantID).Error)
		assert.Equal(t, 15, row.DefaultEstimatedDuration)
	})

	t.Run("UpsertBranchQueueSetting", func(t *testing.T) {
		tenantID := "tenant-2"
		branchID := "branch-1"
		err := repo.UpsertBranchQueueSetting(ctx, tenantID, branchID, map[string]any{"ticket_prefix": "B"})
		require.NoError(t, err)

		var row entity.BranchQueueSetting
		require.NoError(t, db.First(&row, "tenant_id = ? AND branch_id = ?", tenantID, branchID).Error)
		require.NotNil(t, row.TicketPrefix)
		assert.Equal(t, "B", *row.TicketPrefix)
	})

	t.Run("UpsertBranchServiceQueueSetting", func(t *testing.T) {
		tenantID := "tenant-3"
		branchID := "branch-2"
		branchServiceID := "bs-1"
		err := repo.UpsertBranchServiceQueueSetting(ctx, tenantID, branchID, branchServiceID, map[string]any{"allow_skip": true})
		require.NoError(t, err)

		var row entity.BranchServiceQueueSetting
		require.NoError(t, db.First(&row, "tenant_id = ? AND branch_id = ? AND branch_service_id = ?", tenantID, branchID, branchServiceID).Error)
		require.NotNil(t, row.AllowSkip)
		assert.True(t, *row.AllowSkip)
	})

	t.Run("UpsertCounterQueueSetting", func(t *testing.T) {
		tenantID := "tenant-4"
		counterID := "counter-1"
		err := repo.UpsertCounterQueueSetting(ctx, tenantID, counterID, map[string]any{"auto_call_next": false})
		require.NoError(t, err)

		var row entity.CounterQueueSetting
		require.NoError(t, db.First(&row, "tenant_id = ? AND counter_id = ?", tenantID, counterID).Error)
		require.NotNil(t, row.AutoCallNext)
		assert.False(t, *row.AutoCallNext)
	})
}
