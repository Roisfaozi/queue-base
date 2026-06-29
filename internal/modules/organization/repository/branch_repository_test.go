package repository

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func TestBranchRepositoryInterfaceExists(t *testing.T) {
	var _ BranchRepository = (*branchRepository)(nil)
}

func TestBranchEntityTableName(t *testing.T) {
	if got := (entity.Branch{}).TableName(); got != "branches" {
		t.Fatalf("expected branches, got %s", got)
	}
}

func TestBranchRepositoryCreateAndUpdateMatchBranchesSchema(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	if err := db.Exec(`
		CREATE TABLE branches (
			id TEXT PRIMARY KEY,
			tenant_id TEXT NOT NULL,
			code TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT,
			created_at INTEGER,
			updated_at INTEGER,
			deleted_at INTEGER
		)
	`).Error; err != nil {
		t.Fatalf("create branches table: %v", err)
	}

	repo := NewBranchRepository(db)
	branch := &entity.Branch{
		ID:        "branch-1",
		TenantID:  "tenant-1",
		Code:      "MAIN",
		Name:      "Main Branch",
		Status:    entity.BranchStatusActive,
		CreatedAt: 1,
		UpdatedAt: 1,
	}

	if err := repo.Create(context.Background(), branch); err != nil {
		t.Fatalf("create branch: %v", err)
	}

	branch.Name = "Main Branch Updated"
	branch.UpdatedAt = 2
	if err := repo.Update(context.Background(), branch); err != nil {
		t.Fatalf("update branch: %v", err)
	}

	stored, err := repo.FindByID(context.Background(), "tenant-1", "branch-1")
	if err != nil {
		t.Fatalf("find branch: %v", err)
	}
	if stored.Name != "Main Branch Updated" {
		t.Fatalf("expected updated name, got %+v", stored)
	}
}
