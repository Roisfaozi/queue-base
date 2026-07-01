//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	auditRepo "github.com/Roisfaozi/queue-base/internal/modules/audit/repository"
	auditUseCase "github.com/Roisfaozi/queue-base/internal/modules/audit/usecase"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAuditIntegration(env *setup.TestEnvironment) auditUseCase.AuditUseCase {
	repo := auditRepo.NewAuditRepository(env.DB, env.Logger)
	return auditUseCase.NewAuditUseCase(repo, env.Logger, nil, nil)
}

func TestAuditIntegration_LogActivity(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, uc auditUseCase.AuditUseCase, env *setup.TestEnvironment)
	}{
		{
			name:     "LogActivity_And_Query",
			category: "integration",
			run: func(t *testing.T, uc auditUseCase.AuditUseCase, env *setup.TestEnvironment) {
				req := model.CreateAuditLogRequest{
					UserID: "user-1", Action: "CREATE", Entity: "User", EntityID: "entity-1",
					OldValues: map[string]any{"a": 1}, NewValues: map[string]any{"a": 2},
					IPAddress: "127.0.0.1", UserAgent: "TestAgent",
				}
				err := uc.LogActivity(context.Background(), req)
				require.NoError(t, err)

				filter := &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{"entity": {Type: "equals", From: "User"}}}
				logs, _, err := uc.GetLogsDynamic(context.Background(), filter)
				require.NoError(t, err)
				assert.Len(t, logs, 1)
				assert.Equal(t, "CREATE", logs[0].Action)
				assert.Equal(t, "User", logs[0].Entity)
			},
		},
		{
			name:     "LogActivity_MissingRequiredFields",
			category: "integration",
			run: func(t *testing.T, uc auditUseCase.AuditUseCase, env *setup.TestEnvironment) {
				err := uc.LogActivity(context.Background(), model.CreateAuditLogRequest{})
				assert.Error(t, err)
			},
		},
		{
			name:     "LogActivity_NilValues",
			category: "integration",
			run: func(t *testing.T, uc auditUseCase.AuditUseCase, env *setup.TestEnvironment) {
				err := uc.LogActivity(context.Background(), model.CreateAuditLogRequest{
					UserID: "user-1", Action: "UPDATE", Entity: "User", EntityID: "entity-1",
					OldValues: nil, NewValues: nil,
				})
				require.NoError(t, err)
			},
		},
		{
			name:     "Security_SQLInjectionInEntity",
			category: "security",
			run: func(t *testing.T, uc auditUseCase.AuditUseCase, env *setup.TestEnvironment) {
				payload := "User'; DROP TABLE audit_logs;--"
				err := uc.LogActivity(context.Background(), model.CreateAuditLogRequest{
					UserID: "user-1", Action: "CREATE", Entity: payload, EntityID: "entity-1",
				})

				if err == nil {
					filter := &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{"entity": {Type: "equals", From: payload}}}
					logs, _, err := uc.GetLogsDynamic(context.Background(), filter)
					require.NoError(t, err)
					assert.NotEmpty(t, logs)
					assert.Equal(t, payload, logs[0].Entity)
				} else {
					assert.Error(t, err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()
			uc := setupAuditIntegration(env)
			tt.run(t, uc, env)
		})
	}
}
