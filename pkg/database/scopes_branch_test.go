package database

import (
	"context"
	"testing"
)

func TestTenantAliasAndBranchContext(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				ctx := SetOrganizationContext(context.Background(), "tenant-1")
				ctx = SetBranchContext(ctx, "branch-1")

				if got := GetTenantID(ctx); got != "tenant-1" {
					t.Fatalf("expected tenant-1, got %s", got)
				}
				if got := GetBranchID(ctx); got != "branch-1" {
					t.Fatalf("expected branch-1, got %s", got)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
