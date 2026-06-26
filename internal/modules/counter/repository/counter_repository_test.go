package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newCounterTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Counter{}))
	return db
}

func TestCounterRepository_CreateAndFindByID(t *testing.T) {
	db := newCounterTestDB(t)
	repo := NewCounterRepository(db)
	ctx := context.Background()

	counter := &entity.Counter{
		ID:       "c-1",
		TenantID: "t-1",
		BranchID: "b-1",
		Code:     "C1",
		Name:     "Counter 1",
		Status:   entity.CounterStatusActive,
	}

	err := repo.Create(ctx, counter)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, "t-1", "c-1")
	require.NoError(t, err)
	assert.Equal(t, "Counter 1", found.Name)
	assert.Equal(t, "C1", found.Code)
}

func TestCounterRepository_FindAll(t *testing.T) {
	db := newCounterTestDB(t)
	repo := NewCounterRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entity.Counter{ID: "c-1", TenantID: "t-1", BranchID: "b-1", Code: "C1", Name: "C1"})
	repo.Create(ctx, &entity.Counter{ID: "c-2", TenantID: "t-1", BranchID: "b-1", Code: "C2", Name: "C2"})
	repo.Create(ctx, &entity.Counter{ID: "c-3", TenantID: "t-2", BranchID: "b-1", Code: "C3", Name: "C3"})

	found, err := repo.FindAll(ctx, "t-1")
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestCounterRepository_UpdateAndDelete(t *testing.T) {
	db := newCounterTestDB(t)
	repo := NewCounterRepository(db)
	ctx := context.Background()

	counter := &entity.Counter{
		ID:       "c-1",
		TenantID: "t-1",
		BranchID: "b-1",
		Code:     "C1",
		Name:     "Counter 1",
		Status:   entity.CounterStatusActive,
	}
	repo.Create(ctx, counter)

	// Update
	now := time.Now().UnixMilli()
	err := repo.Update(ctx, &entity.Counter{
		ID:        "c-1",
		TenantID:  "t-1",
		Code:      "C1-NEW",
		Name:      "Counter One",
		Status:    entity.CounterStatusInactive,
		UpdatedAt: now,
	})
	require.NoError(t, err)

	updated, _ := repo.FindByID(ctx, "t-1", "c-1")
	assert.Equal(t, "C1-NEW", updated.Code)
	assert.Equal(t, "Counter One", updated.Name)
	assert.Equal(t, entity.CounterStatusInactive, updated.Status)
	assert.Equal(t, now, updated.UpdatedAt)

	// Update missing
	err = repo.Update(ctx, &entity.Counter{ID: "c-99", TenantID: "t-1"})
	assert.ErrorIs(t, err, exception.ErrNotFound)

	// Delete
	err = repo.Delete(ctx, "t-1", "c-1")
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, "t-1", "c-1")
	assert.Error(t, err)

	// Delete missing
	err = repo.Delete(ctx, "t-1", "c-99")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}
