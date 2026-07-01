//go:build integration
// +build integration

package modules

import (
	"context"
	"fmt"
	"testing"

	orgEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	orgRepo "github.com/Roisfaozi/queue-base/internal/modules/organization/repository"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOrganizationRepository_Integration(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, env *setup.TestEnvironment)
	}{
		{
			name:     "Create Organization Success",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				user := setup.CreateTestUser(t, env.DB, "testowner", "owner@test.com", "password123")
				org := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Test Organization",
					Slug:    "test-org-" + uuid.New().String()[:8],
					OwnerID: user.ID,
					Status:  orgEntity.OrgStatusActive,
				}

				err := env.DB.Create(org).Error
				require.NoError(t, err)

				member := &orgEntity.OrganizationMember{
					ID:             uuid.New().String(),
					OrganizationID: org.ID,
					UserID:         user.ID,
					RoleID:         "role:admin",
					Status:         orgEntity.MemberStatusActive,
				}
				err = env.DB.Create(member).Error
				require.NoError(t, err)

				var foundOrg orgEntity.Organization
				err = env.DB.First(&foundOrg, "id = ?", org.ID).Error
				require.NoError(t, err)
				assert.Equal(t, org.Name, foundOrg.Name)
			},
		},
		{
			name:     "Slug Uniqueness",
			category: "negative",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				user := setup.CreateTestUser(t, env.DB, "sluguser", "slug@test.com", "password123")
				slug := "unique-slug-" + uuid.New().String()[:8]

				org1 := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Org 1",
					Slug:    slug,
					OwnerID: user.ID,
					Status:  orgEntity.OrgStatusActive,
				}
				err := env.DB.Create(org1).Error
				require.NoError(t, err)

				org2 := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Org 2",
					Slug:    slug,
					OwnerID: user.ID,
					Status:  orgEntity.OrgStatusActive,
				}
				err = env.DB.Create(org2).Error
				assert.Error(t, err)
			},
		},
		{
			name:     "Data Isolation",
			category: "edge",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				userA := setup.CreateTestUser(t, env.DB, "usera", "usera@test.com", "password123")
				userB := setup.CreateTestUser(t, env.DB, "userb", "userb@test.com", "password123")

				orgA := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Org A",
					Slug:    "org-a-" + uuid.New().String()[:8],
					OwnerID: userA.ID,
					Status:  orgEntity.OrgStatusActive,
				}
				err := env.DB.Create(orgA).Error
				require.NoError(t, err)

				orgB := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Org B",
					Slug:    "org-b-" + uuid.New().String()[:8],
					OwnerID: userB.ID,
					Status:  orgEntity.OrgStatusActive,
				}
				err = env.DB.Create(orgB).Error
				require.NoError(t, err)

				memberA := &orgEntity.OrganizationMember{
					ID:             uuid.New().String(),
					OrganizationID: orgA.ID,
					UserID:         userA.ID,
					RoleID:         "role:admin",
					Status:         orgEntity.MemberStatusActive,
				}
				env.DB.Create(memberA)

				memberB := &orgEntity.OrganizationMember{
					ID:             uuid.New().String(),
					OrganizationID: orgB.ID,
					UserID:         userB.ID,
					RoleID:         "role:admin",
					Status:         orgEntity.MemberStatusActive,
				}
				env.DB.Create(memberB)

				var membersOfA []*orgEntity.OrganizationMember
				err = env.DB.Where("organization_id = ?", orgA.ID).Find(&membersOfA).Error
				require.NoError(t, err)
				assert.Len(t, membersOfA, 1)
				assert.Equal(t, userA.ID, membersOfA[0].UserID)
			},
		},
		{
			name:     "Check Membership",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				user := setup.CreateTestUser(t, env.DB, "memberuser", "member@test.com", "password123")
				org := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Member Test Org",
					Slug:    "member-test-" + uuid.New().String()[:8],
					OwnerID: user.ID,
					Status:  orgEntity.OrgStatusActive,
				}
				env.DB.Create(org)

				member := &orgEntity.OrganizationMember{
					ID:             uuid.New().String(),
					OrganizationID: org.ID,
					UserID:         user.ID,
					RoleID:         "role:user",
					Status:         orgEntity.MemberStatusActive,
				}
				env.DB.Create(member)

				var count int64
				err := env.DB.Model(&orgEntity.OrganizationMember{}).
					Where("organization_id = ? AND user_id = ? AND status = ?", org.ID, user.ID, orgEntity.MemberStatusActive).
					Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(1), count)
			},
		},
		{
			name:     "Banned Member Denied",
			category: "negative",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				user := setup.CreateTestUser(t, env.DB, "banneduser", "banned@test.com", "password123")
				org := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Ban Test Org",
					Slug:    "ban-test-" + uuid.New().String()[:8],
					OwnerID: user.ID,
					Status:  orgEntity.OrgStatusActive,
				}
				env.DB.Create(org)

				member := &orgEntity.OrganizationMember{
					ID:             uuid.New().String(),
					OrganizationID: org.ID,
					UserID:         user.ID,
					RoleID:         "role:user",
					Status:         orgEntity.MemberStatusBanned,
				}
				env.DB.Create(member)

				var count int64
				err := env.DB.Model(&orgEntity.OrganizationMember{}).
					Where("organization_id = ? AND user_id = ? AND status = ?", org.ID, user.ID, orgEntity.MemberStatusActive).
					Count(&count).Error
				require.NoError(t, err)
				assert.Equal(t, int64(0), count)
			},
		},
		{
			name:     "Atomic Create",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				orgRepoImpl := orgRepo.NewOrganizationRepository(env.DB)
				user := setup.CreateTestUser(t, env.DB, "atomicowner", "atomic@test.com", "password123")
				org := &orgEntity.Organization{
					ID:      uuid.New().String(),
					Name:    "Atomic Test Org",
					Slug:    "atomic-test-" + uuid.New().String()[:8],
					OwnerID: user.ID,
					Status:  orgEntity.OrgStatusActive,
				}

				err := orgRepoImpl.Create(context.Background(), org, "role:admin")
				require.NoError(t, err)

				foundOrg, err := orgRepoImpl.FindByID(context.Background(), org.ID)
				require.NoError(t, err)
				assert.Equal(t, org.Name, foundOrg.Name)
			},
		},
		{
			name:     "Find User Organizations",
			category: "positive",
			run: func(t *testing.T, env *setup.TestEnvironment) {
				orgRepoImpl := orgRepo.NewOrganizationRepository(env.DB)
				user := setup.CreateTestUser(t, env.DB, "multiorguser", "multiorg@test.com", "password123")

				for i := 1; i <= 3; i++ {
					org := &orgEntity.Organization{
						ID:      uuid.New().String(),
						Name:    fmt.Sprintf("Org %d", i),
						Slug:    fmt.Sprintf("org-%d-%s", i, uuid.New().String()[:8]),
						OwnerID: user.ID,
						Status:  orgEntity.OrgStatusActive,
					}
					env.DB.Create(org)

					if i <= 2 {
						member := &orgEntity.OrganizationMember{
							ID:             uuid.New().String(),
							OrganizationID: org.ID,
							UserID:         user.ID,
							RoleID:         "role:member",
							Status:         orgEntity.MemberStatusActive,
						}
						env.DB.Create(member)
					}
				}

				orgs, err := orgRepoImpl.FindUserOrganizations(context.Background(), user.ID)
				require.NoError(t, err)
				assert.Len(t, orgs, 2)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			if env == nil {
				return
			}
			defer env.Cleanup()
			tt.run(t, env)
		})
	}
}