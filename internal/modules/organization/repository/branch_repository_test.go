package repository

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newBranchTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.Branch{}))
	return db
}

func TestBranchRepositoryInterfaceExists(t *testing.T) {
	var _ BranchRepository = (*branchRepository)(nil)
}

func TestBranchEntityTableName(t *testing.T) {
	if got := (entity.Branch{}).TableName(); got != "branches" {
		t.Fatalf("expected branches, got %s", got)
	}
}

func TestBranchRepository_CreateAndFindByID(t *testing.T) {
	db := newBranchTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	branch := &entity.Branch{
		ID:       "b-1",
		TenantID: "t-1",
		Code:     "B1",
		Name:     "Branch 1",
		Status:   entity.BranchStatusActive,
	}

	err := repo.Create(ctx, branch)
	require.NoError(t, err)

	found, err := repo.FindByID(ctx, "t-1", "b-1")
	require.NoError(t, err)
	assert.Equal(t, "Branch 1", found.Name)
	assert.Equal(t, "B1", found.Code)
}

func TestBranchRepository_FindAll(t *testing.T) {
	db := newBranchTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	repo.Create(ctx, &entity.Branch{ID: "b-1", TenantID: "t-1", Code: "B1", Name: "B1"})
	repo.Create(ctx, &entity.Branch{ID: "b-2", TenantID: "t-1", Code: "B2", Name: "B2"})
	repo.Create(ctx, &entity.Branch{ID: "b-3", TenantID: "t-2", Code: "B3", Name: "B3"})

	found, err := repo.FindAll(ctx, "t-1")
	require.NoError(t, err)
	assert.Len(t, found, 2)
}

func TestBranchRepository_UpdateAndDelete(t *testing.T) {
	db := newBranchTestDB(t)
	repo := NewBranchRepository(db)
	ctx := context.Background()

	branch := &entity.Branch{
		ID:       "b-1",
		TenantID: "t-1",
		Code:     "B1",
		Name:     "Branch 1",
		Status:   entity.BranchStatusActive,
	}
	repo.Create(ctx, branch)

	// Update
	now := time.Now().UnixMilli()
	err := repo.Update(ctx, &entity.Branch{
		ID:        "b-1",
		TenantID:  "t-1",
		Code:      "B1-NEW",
		Name:      "Branch One",
		Status:    entity.BranchStatusInactive,
		UpdatedAt: now,
	})
	require.NoError(t, err)

	updated, _ := repo.FindByID(ctx, "t-1", "b-1")
	assert.Equal(t, "B1-NEW", updated.Code)
	assert.Equal(t, "Branch One", updated.Name)
	assert.Equal(t, entity.BranchStatusInactive, updated.Status)
	assert.Equal(t, now, updated.UpdatedAt)

	// Update missing
	err = repo.Update(ctx, &entity.Branch{ID: "b-99", TenantID: "t-1"})
	assert.ErrorIs(t, err, exception.ErrNotFound)

	// Delete
	err = repo.Delete(ctx, "t-1", "b-1")
	require.NoError(t, err)

	_, err = repo.FindByID(ctx, "t-1", "b-1")
	assert.Error(t, err)

	// Delete missing
	err = repo.Delete(ctx, "t-1", "b-99")
	assert.ErrorIs(t, err, exception.ErrNotFound)
}
