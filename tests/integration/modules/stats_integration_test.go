//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"
	"time"

	auditEntity "github.com/Roisfaozi/queue-base/internal/modules/audit/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/stats/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupStatsIntegration(env *setup.TestEnvironment) usecase.StatsUseCase {
	log := logrus.New()
	log.SetLevel(logrus.PanicLevel)
	return usecase.NewStatsUseCase(env.DB, log)
}

func TestStatsIntegration_DashboardSummary_RealCounts(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				statsUC := setupStatsIntegration(env)

				setup.CreateTestUser(t, env.DB, "stats_alice", "alice_stats@test.com", "Pass123!")
				setup.CreateTestUser(t, env.DB, "stats_bob", "bob_stats@test.com", "Pass123!")

				summary, err := statsUC.GetDashboardSummary(context.Background())

				require.NoError(t, err)
				assert.NotNil(t, summary)
				assert.GreaterOrEqual(t, summary.TotalUsers, int64(2), "Should have at least 2 users")
				assert.GreaterOrEqual(t, summary.TotalRoles, int64(1), "Should have at least seed roles")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestStatsIntegration_DashboardActivity_WithAuditLogs(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				statsUC := setupStatsIntegration(env)

				now := time.Now()
				todayNoon := time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, now.Location())

				for i := 0; i < 3; i++ {
					env.DB.Create(&auditEntity.AuditLog{
						ID:        uuid.NewString(),
						UserID:    "test-user",
						Action:    "LOGIN",
						Entity:    "Auth",
						EntityID:  uuid.NewString(),
						CreatedAt: todayNoon.Add(time.Duration(i) * time.Minute).UnixMilli(),
					})
				}

				env.DB.Create(&auditEntity.AuditLog{
					ID:        uuid.NewString(),
					UserID:    "test-user",
					Action:    "VIEWED_DASHBOARD",
					Entity:    "Stats",
					EntityID:  uuid.NewString(),
					CreatedAt: todayNoon.Add(10 * time.Minute).UnixMilli(),
				})

				activity, err := statsUC.GetDashboardActivity(context.Background(), 7)

				require.NoError(t, err)
				assert.NotNil(t, activity)
				assert.Len(t, activity.Points, 7)

				for _, point := range activity.Points {
					assert.NotEmpty(t, point.Date)
					assert.GreaterOrEqual(t, point.Audits, int64(0))
					assert.GreaterOrEqual(t, point.Logins, int64(0))
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

func TestStatsIntegration_DashboardActivity_EmptyData(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				statsUC := setupStatsIntegration(env)

				activity, err := statsUC.GetDashboardActivity(context.Background(), 3)

				require.NoError(t, err)
				assert.Len(t, activity.Points, 3)
				for _, point := range activity.Points {
					assert.NotEmpty(t, point.Date)
					assert.GreaterOrEqual(t, point.Audits, int64(0))
					assert.GreaterOrEqual(t, point.Logins, int64(0))
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

func TestStatsIntegration_SystemInsights(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				statsUC := setupStatsIntegration(env)

				insights, err := statsUC.GetSystemInsights(context.Background())

				require.NoError(t, err)
				assert.NotNil(t, insights)
				assert.Greater(t, insights.AvgLatencyMs, float64(0))
				assert.NotEmpty(t, insights.Uptime)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}

func TestStatsIntegration_OrganizationScoping(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T)
	}{
		{
			name:     "Success",
			category: "positive",
			run: func(t *testing.T) {
				env := setup.SetupIntegrationEnvironment(t)
				defer env.Cleanup()

				statsUC := setupStatsIntegration(env)

				globalSummary, err := statsUC.GetDashboardSummary(context.Background())
				require.NoError(t, err)

				ctx := database.SetOrganizationContext(context.Background(), "non-existent-org")
				scopedSummary, err := statsUC.GetDashboardSummary(ctx)
				require.NoError(t, err)

				assert.LessOrEqual(t, scopedSummary.TotalUsers, globalSummary.TotalUsers)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.run(t)
		})
	}
}
