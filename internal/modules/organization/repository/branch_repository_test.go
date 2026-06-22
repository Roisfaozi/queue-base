package repository

import (
	"testing"

	"github.com/Roisfaozi/go-clean-boilerplate/internal/modules/organization/entity"
)

func TestBranchRepositoryInterfaceExists(t *testing.T) {
	var _ BranchRepository = (*branchRepository)(nil)
}

func TestBranchEntityTableName(t *testing.T) {
	if got := (entity.Branch{}).TableName(); got != "branches" {
		t.Fatalf("expected branches, got %s", got)
	}
}
