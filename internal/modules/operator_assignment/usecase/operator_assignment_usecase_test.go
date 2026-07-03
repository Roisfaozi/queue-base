package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubAudit struct {
	reqs []auditModel.CreateAuditLogRequest
}

func (s *stubAudit) LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error {
	s.reqs = append(s.reqs, req)
	return nil
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&entity.OperatorCounterAssignment{}))
	return db
}

func TestOperatorAssignmentUseCase(t *testing.T) {
	db := newTestDB(t)
	audit := &stubAudit{}
	uc := NewOperatorAssignmentUseCase(db, audit)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")

	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "Positive_Create_List_Delete",
			fn: func(t *testing.T) {
				res, err := uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-1", CounterID: "c-1"})
				require.NoError(t, err)
				require.NotNil(t, res)
				assert.Equal(t, "tenant-1", res.TenantID)

				rows, err := uc.GetAll(ctx)
				require.NoError(t, err)
				require.Len(t, rows, 1)

				require.NoError(t, uc.Delete(ctx, res.ID))
				rows, err = uc.GetAll(ctx)
				require.NoError(t, err)
				require.Len(t, rows, 1)
				assert.NotNil(t, rows[0].UnassignedAt)
			},
		},
		{
			name: "Negative_MissingTenant",
			fn: func(t *testing.T) {
				_, err := uc.Create(context.Background(), &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-1", CounterID: "c-1"})
				require.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
