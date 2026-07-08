package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/queue_config/test/mocks"
	"github.com/Roisfaozi/queue-base/pkg/authcontext"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueueConfigUseCase_UpdateAndReset(t *testing.T) {
	repo := &mocks.MockQueueConfigRepository{}
	audit := &mocks.MockAuditUseCase{}
	uc := NewQueueConfigUseCase(repo, audit)

	ctx := authcontext.WithUserID(context.Background(), "user-1")

	t.Run("tenant update audits and upserts", func(t *testing.T) {
		repo.On("UpsertTenantQueueSetting", ctx, "tenant-1", map[string]any{"ticket_prefix": "A"}).Return(nil).Once()
		audit.On("LogActivity", ctx, auditModel.CreateAuditLogRequest{UserID: "user-1", Action: "SETTING_UPDATE", Entity: "queue_setting", EntityID: "tenant_queue_settings:tenant-1", NewValues: map[string]string{"ticket_prefix": "A"}}).Return(nil).Once()

		err := uc.UpdateTenantQueueSetting(ctx, "tenant-1", map[string]any{"ticket_prefix": "A"})
		require.NoError(t, err)
	})

	t.Run("bad request on empty tenant", func(t *testing.T) {
		err := uc.UpdateTenantQueueSetting(ctx, "", map[string]any{})
		assert.ErrorIs(t, err, exception.ErrBadRequest)
	})

	t.Run("reset no audit no error", func(t *testing.T) {
		uc2 := NewQueueConfigUseCase(repo, nil)
		err := uc2.ResetQueueSetting(context.Background(), "branch_queue_settings", "b1", "ticket_prefix")
		require.NoError(t, err)
	})
}
