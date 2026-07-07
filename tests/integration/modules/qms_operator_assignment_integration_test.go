//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	operatorModule "github.com/Roisfaozi/queue-base/internal/modules/operator_assignment"
	operatorEntity "github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/entity"
	operatorModel "github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/model"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func ensureOperatorAssignmentTable(t *testing.T, db *gorm.DB) {
	t.Helper()
	require.NoError(t, db.AutoMigrate(&operatorEntity.OperatorCounterAssignment{}))
}

func setupOperatorAssignmentIntegration(t *testing.T) (*gorm.DB, *operatorModule.Module, string, string, string, string, string) {
	env := setup.SetupIntegrationEnvironment(t)
	if env == nil {
		t.Skip("Skipping integration test; DB not available")
	}
	db := env.DB
	ensureOperatorAssignmentTable(t, db)

	mod := operatorModule.NewModule(db, validator.New(), env.Logger)

	tenantID := "tenant-" + uuid.NewString()[:8]
	branchID := "branch-" + uuid.NewString()[:8]
	counterID := "counter-" + uuid.NewString()[:8]
	user1 := "user1-" + uuid.NewString()[:8]
	user2 := "user2-" + uuid.NewString()[:8]

	require.NoError(t, db.Create(&branchEntity.Branch{
		ID:       branchID,
		TenantID: tenantID,
		Code:     "BR1",
		Name:     "Test Branch",
		Status:   branchEntity.BranchStatusActive,
	}).Error)

	require.NoError(t, db.Create(&counterEntity.Counter{
		ID:       counterID,
		TenantID: tenantID,
		BranchID: branchID,
		Code:     "CT1",
		Name:     "Test Counter",
		Status:   counterEntity.CounterStatusActive,
	}).Error)

	require.NoError(t, db.Create(&userEntity.User{
		ID:             user1,
		OrganizationID: &tenantID,
		Email:          user1 + "@example.com",
		Username:       user1,
		Status:         userEntity.UserStatusActive,
	}).Error)

	require.NoError(t, db.Create(&userEntity.User{
		ID:             user2,
		OrganizationID: &tenantID,
		Email:          user2 + "@example.com",
		Username:       user2,
		Status:         userEntity.UserStatusActive,
	}).Error)

	return db, mod, tenantID, branchID, counterID, user1, user2
}

func TestIntegration_OperatorAssignment(t *testing.T) {
	db, mod, tenantID, branchID, counterID, user1, user2 := setupOperatorAssignmentIntegration(t)
	ctx := database.SetOrganizationContext(context.Background(), tenantID)

	// Create assignment for user1
	t.Run("Create_Success", func(t *testing.T) {
		req := &operatorModel.OperatorAssignmentRequest{
			BranchID:  branchID,
			UserID:    user1,
			CounterID: counterID,
		}
		res, err := mod.UseCase.Create(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, branchID, res.BranchID)
		assert.Equal(t, user1, res.UserID)
		assert.Equal(t, counterID, res.CounterID)
	})

	// Try assigning user2 to same counter (multi-operator counter check)
	// Even if UI blocks it, the backend should correctly log and potentially allow or reject depending on logic.
	// Currently operator_assignment uc.Create does not block multiple active assignments to same counter.
	t.Run("Create_MultiOperatorOnSameCounter", func(t *testing.T) {
		req := &operatorModel.OperatorAssignmentRequest{
			BranchID:  branchID,
			UserID:    user2,
			CounterID: counterID,
		}
		res, err := mod.UseCase.Create(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, user2, res.UserID)
	})

	// Get all
	t.Run("GetAll_ReturnsActive", func(t *testing.T) {
		rows, err := mod.UseCase.GetAll(ctx)
		require.NoError(t, err)
		assert.Len(t, rows, 2)
	})

	// Cleanup unassign user1
	t.Run("Unassign_User1", func(t *testing.T) {
		// Find user1 assignment
		var assignment operatorEntity.OperatorCounterAssignment
		require.NoError(t, db.Where("tenant_id = ? AND user_id = ? AND unassigned_at IS NULL", tenantID, user1).First(&assignment).Error)

		err := mod.UseCase.Delete(ctx, assignment.ID)
		require.NoError(t, err)
	})
}
