//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	projectEntity "github.com/Roisfaozi/queue-base/internal/modules/project/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/project/model"
	"github.com/Roisfaozi/queue-base/internal/modules/project/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/project/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupProjectIntegration(env *setup.TestEnvironment) usecase.ProjectUseCase {
	repo := repository.NewProjectRepository(env.DB)
	return usecase.NewProjectUseCase(repo)
}

func TestProjectIntegration_Create_Success(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_Create",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				uc := setupProjectIntegration(env)
				ctx := context.Background()
				owner := setup.CreateTestUser(t, env.DB, "project_owner_create", "project_owner_create@test.com", "Password123!")
				org := setup.CreateTestOrganization(t, env.DB, owner.ID, "Project Create Org", "project-create-org-"+uuid.NewString()[:8])

				req := model.CreateProjectRequest{
					Name:   "Integration Test Project",
					Domain: "integration.example.com",
				}

				result, err := uc.CreateProject(ctx, owner.ID, org.ID, req)

				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.NotEmpty(t, result.ID)
				assert.Equal(t, "Integration Test Project", result.Name)
				assert.Equal(t, "integration.example.com", result.Domain)
				assert.Equal(t, "active", result.Status)
				assert.Equal(t, org.ID, result.OrganizationID)
				assert.Equal(t, owner.ID, result.UserID)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProjectIntegration_CRUD_Lifecycle(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_CRUD",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				uc := setupProjectIntegration(env)
				ctx := context.Background()
				user := setup.CreateTestUser(t, env.DB, "project_lifecycle", "project_lifecycle@test.com", "Password123!")
				org := setup.CreateTestOrganization(t, env.DB, user.ID, "Project Lifecycle Org", "project-lifecycle-org-"+uuid.NewString()[:8])
				orgID := org.ID
				userID := user.ID

				req := model.CreateProjectRequest{
					Name:   "Lifecycle Project",
					Domain: "lifecycle.example.com",
				}
				created, err := uc.CreateProject(ctx, userID, orgID, req)
				require.NoError(t, err)
				require.NotNil(t, created)
				projectID := created.ID

				fetched, err := uc.GetProjectByID(ctx, projectID)
				require.NoError(t, err)
				assert.Equal(t, "Lifecycle Project", fetched.Name)
				assert.Equal(t, "lifecycle.example.com", fetched.Domain)

				name := "Updated Lifecycle"
				domain := "updated.example.com"
				status := "inactive"
				updateReq := model.UpdateProjectRequest{
					Name:   &name,
					Domain: &domain,
					Status: &status,
				}
				updated, err := uc.UpdateProject(ctx, projectID, updateReq)
				require.NoError(t, err)
				assert.Equal(t, "Updated Lifecycle", updated.Name)
				assert.Equal(t, "updated.example.com", updated.Domain)
				assert.Equal(t, "inactive", updated.Status)

				err = uc.DeleteProject(ctx, projectID)
				require.NoError(t, err)

				_, err = uc.GetProjectByID(ctx, projectID)
				assert.Error(t, err, "Should not find deleted project")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProjectIntegration_GetByOrgID(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_GetByOrgID",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				uc := setupProjectIntegration(env)
				ctx := context.Background()
				userA := setup.CreateTestUser(t, env.DB, "project_org_a", "project_org_a@test.com", "Password123!")
				userB := setup.CreateTestUser(t, env.DB, "project_org_b", "project_org_b@test.com", "Password123!")
				orgA := setup.CreateTestOrganization(t, env.DB, userA.ID, "Project Org A", "project-org-a-"+uuid.NewString()[:8]).ID
				orgB := setup.CreateTestOrganization(t, env.DB, userB.ID, "Project Org B", "project-org-b-"+uuid.NewString()[:8]).ID

				uc.CreateProject(ctx, "user-1", orgA, model.CreateProjectRequest{Name: "Org A Proj 1", Domain: "a1.com"})
				uc.CreateProject(ctx, "user-1", orgA, model.CreateProjectRequest{Name: "Org A Proj 2", Domain: "a2.com"})
				uc.CreateProject(ctx, "user-2", orgB, model.CreateProjectRequest{Name: "Org B Proj 1", Domain: "b1.com"})

				projectsA, err := uc.GetProjects(ctx, orgA)
				require.NoError(t, err)
				assert.Len(t, projectsA, 2)

				projectsB, err := uc.GetProjects(ctx, orgB)
				require.NoError(t, err)
				assert.Len(t, projectsB, 1)

				projectsNone, err := uc.GetProjects(ctx, "org-nonexistent")
				require.NoError(t, err)
				assert.Len(t, projectsNone, 0)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProjectIntegration_PartialUpdate(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success_PartialUpdate",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				uc := setupProjectIntegration(env)
				ctx := context.Background()
				user := setup.CreateTestUser(t, env.DB, "project_partial", "project_partial@test.com", "Password123!")
				org := setup.CreateTestOrganization(t, env.DB, user.ID, "Project Partial Org", "project-partial-org-"+uuid.NewString()[:8])

				created, err := uc.CreateProject(ctx, user.ID, org.ID, model.CreateProjectRequest{
					Name:   "Partial Update Project",
					Domain: "partial.example.com",
				})
				require.NoError(t, err)

				nameStr := "Updated Name Only"
				updated, err := uc.UpdateProject(ctx, created.ID, model.UpdateProjectRequest{
					Name: &nameStr,
				})
				require.NoError(t, err)
				assert.Equal(t, "Updated Name Only", updated.Name)
				assert.Equal(t, "partial.example.com", updated.Domain, "Domain should not change")
				assert.Equal(t, "active", updated.Status, "Status should not change")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProjectIntegration_Security_SQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_SQLInjection",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				uc := setupProjectIntegration(env)
				ctx := context.Background()
				user := setup.CreateTestUser(t, env.DB, "project_sqli", "project_sqli@test.com", "Password123!")
				org := setup.CreateTestOrganization(t, env.DB, user.ID, "Project SQLi Org", "project-sqli-org-"+uuid.NewString()[:8])

				result, err := uc.CreateProject(ctx, user.ID, org.ID, model.CreateProjectRequest{
					Name:   "'; DROP TABLE projects; --",
					Domain: "sqli.example.com",
				})

				require.NoError(t, err)
				assert.NotNil(t, result)

				var count int64
				env.DB.Model(&projectEntity.Project{}).Count(&count)
				assert.GreaterOrEqual(t, count, int64(1), "Projects table should still exist")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProjectIntegration_OrganizationScopeIsolation(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_ScopeIsolation",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				repo := repository.NewProjectRepository(env.DB)
				ctx := context.Background()

				userA := setup.CreateTestUser(t, env.DB, "project_repo_iso_a", "project_repo_iso_a@test.com", "Password123!")
				userB := setup.CreateTestUser(t, env.DB, "project_repo_iso_b", "project_repo_iso_b@test.com", "Password123!")
				orgA := setup.CreateTestOrganization(t, env.DB, userA.ID, "Project Repo Iso A", "project-repo-iso-a-"+uuid.NewString()[:8]).ID
				orgB := setup.CreateTestOrganization(t, env.DB, userB.ID, "Project Repo Iso B", "project-repo-iso-b-"+uuid.NewString()[:8]).ID

				projA := &projectEntity.Project{
					OrganizationID: orgA,
					UserID:         "user-a",
					Name:           "Org A Secret Project",
					Domain:         "secret-a.com",
					Status:         "active",
				}
				err := repo.Create(ctx, projA)
				require.NoError(t, err)

				ctxOrgB := database.SetOrganizationContext(ctx, orgB)
				_, err = repo.GetByID(ctxOrgB, projA.ID)
				assert.Error(t, err, "Org B should not be able to access Org A's project")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestProjectIsolation_UseCase(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Negative_UseCaseIsolation",
			category: "vulnerability",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				repo := repository.NewProjectRepository(env.DB)
				uc := usecase.NewProjectUseCase(repo)
				ctx := context.Background()

				userA := setup.CreateTestUser(t, env.DB, "project_uc_iso_a", "project_uc_iso_a@test.com", "Password123!")
				userB := setup.CreateTestUser(t, env.DB, "project_uc_iso_b", "project_uc_iso_b@test.com", "Password123!")
				orgA := setup.CreateTestOrganization(t, env.DB, userA.ID, "Project UseCase Iso A", "project-uc-iso-a-"+uuid.NewString()[:8]).ID
				orgB := setup.CreateTestOrganization(t, env.DB, userB.ID, "Project UseCase Iso B", "project-uc-iso-b-"+uuid.NewString()[:8]).ID

				ctxOrgA := database.SetOrganizationContext(ctx, orgA)
				created, err := uc.CreateProject(ctxOrgA, userA.ID, orgA, model.CreateProjectRequest{
					Name:   "Org A Private Project",
					Domain: "a.example.com",
				})
				require.NoError(t, err)
				require.NotNil(t, created)

				t.Run("Cross-tenant GET should fail", func(t *testing.T) {
					ctxOrgB := database.SetOrganizationContext(ctx, orgB)
					_, err := uc.GetProjectByID(ctxOrgB, created.ID)
					assert.Error(t, err, "Should not be able to fetch project from another organization")
				})

				t.Run("Cross-tenant UPDATE should fail", func(t *testing.T) {
					ctxOrgB := database.SetOrganizationContext(ctx, orgB)
					newName := "Hacked Name"
					_, err := uc.UpdateProject(ctxOrgB, created.ID, model.UpdateProjectRequest{
						Name: &newName,
					})
					assert.Error(t, err, "Should not be able to update project from another organization")

					refetched, _ := uc.GetProjectByID(ctxOrgA, created.ID)
					assert.Equal(t, "Org A Private Project", refetched.Name)
				})

				t.Run("Cross-tenant DELETE should fail", func(t *testing.T) {
					ctxOrgB := database.SetOrganizationContext(ctx, orgB)
					err := uc.DeleteProject(ctxOrgB, created.ID)
					assert.NoError(t, err)

					refetched, err := uc.GetProjectByID(ctxOrgA, created.ID)
					assert.NoError(t, err)
					assert.NotNil(t, refetched)
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
