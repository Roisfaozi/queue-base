//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/user/model"
	"github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataIsolation_User_FindAll(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		return
	}

	repo := repository.NewUserRepository(env.DB, env.Logger)

	orgA := "org-a"
	orgB := "org-b"

	env.DB.Exec("DELETE FROM organization_members")
	env.DB.Exec("DELETE FROM organizations")
	env.DB.Exec("INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		orgA, "Org A", "org-a", "system", "active", time.Now().UnixMilli(), time.Now().UnixMilli())
	env.DB.Exec("INSERT INTO organizations (id, name, slug, owner_id, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		orgB, "Org B", "org-b", "system", "active", time.Now().UnixMilli(), time.Now().UnixMilli())

	userA := setup.CreateTestUser(t, env.DB, "usera_iso", "user.a.iso@example.com", "password", orgA)
	userB := setup.CreateTestUser(t, env.DB, "userb_iso", "user.b.iso@example.com", "password", orgB)

	// Verify memberships exist
	var count int64
	env.DB.Table("organization_members").Where("organization_id = ?", orgA).Count(&count)
	assert.Equal(t, int64(1), count, "Org A should have 1 member")
	env.DB.Table("organization_members").Where("organization_id = ?", orgB).Count(&count)
	assert.Equal(t, int64(1), count, "Org B should have 1 member")

	tests := []struct {
		name     string
		category string
		queryFn  func(ctx context.Context) ([]*entity.User, error)
		assertFn func(t *testing.T, users []*entity.User, err error)
	}{
		{
			name:     "Scope to Org A returns user A only",
			category: "positive",
			queryFn: func(ctx context.Context) ([]*entity.User, error) {
				users, _, err := repo.FindAll(ctx, &model.GetUserListRequest{})
				return users, err
			},
			assertFn: func(t *testing.T, users []*entity.User, err error) {
				require.NoError(t, err)
				foundA, foundB := false, false
				for _, u := range users {
					if u.ID == userA.ID {
						foundA = true
					}
					if u.ID == userB.ID {
						foundB = true
					}
				}
				assert.True(t, foundA, "Should find user A in Org A context")
				assert.False(t, foundB, "Should NOT find user B in Org A context")
			},
		},
		{
			name:     "Scope to Org B returns user B only",
			category: "positive",
			queryFn: func(ctx context.Context) ([]*entity.User, error) {
				users, _, err := repo.FindAll(ctx, &model.GetUserListRequest{})
				return users, err
			},
			assertFn: func(t *testing.T, users []*entity.User, err error) {
				require.NoError(t, err)
				foundA, foundB := false, false
				for _, u := range users {
					if u.ID == userA.ID {
						foundA = true
					}
					if u.ID == userB.ID {
						foundB = true
					}
				}
				assert.False(t, foundA, "Should NOT find user A in Org B context")
				assert.True(t, foundB, "Should find user B in Org B context")
			},
		},
		{
			name:     "GetByOrganization returns org A users",
			category: "edge",
			queryFn: func(ctx context.Context) ([]*entity.User, error) {
				return repo.GetByOrganization(ctx, orgA)
			},
			assertFn: func(t *testing.T, users []*entity.User, err error) {
				require.NoError(t, err)
				assert.NotEmpty(t, users)
				assert.Equal(t, userA.ID, users[0].ID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			switch tt.name {
			case "Scope to Org A returns user A only":
				ctx = database.SetOrganizationContext(context.Background(), orgA)
			case "Scope to Org B returns user B only":
				ctx = database.SetOrganizationContext(context.Background(), orgB)
			default:
				ctx = context.Background()
			}

			users, err := tt.queryFn(ctx)
			tt.assertFn(t, users, err)
		})
	}
}
