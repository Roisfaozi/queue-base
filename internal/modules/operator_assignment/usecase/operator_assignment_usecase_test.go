package usecase

import (
	"context"
	"testing"

	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	counterEntity "github.com/Roisfaozi/queue-base/internal/modules/counter/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/model"
	"github.com/Roisfaozi/queue-base/internal/modules/operator_assignment/test/mocks"
	branchEntity "github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	userEntity "github.com/Roisfaozi/queue-base/internal/modules/user/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&branchEntity.Branch{}, &counterEntity.Counter{}, &userEntity.User{}, &entity.OperatorCounterAssignment{}))
	return db
}

func TestOperatorAssignmentUseCase(t *testing.T) {
	db := newTestDB(t)
	var capturedAuditReqs []auditModel.CreateAuditLogRequest
	mockAudit := mocks.NewMockAuditLogger(t)
	mockAudit.On("LogActivity", mock.Anything, mock.MatchedBy(func(req auditModel.CreateAuditLogRequest) bool {
		capturedAuditReqs = append(capturedAuditReqs, req)
		return true
	})).Return(nil).Maybe()

	uc := NewOperatorAssignmentUseCase(db, nil, mockAudit)
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	require.NoError(t, db.Create(&branchEntity.Branch{ID: "b-1", TenantID: "tenant-1", Code: "B1", Name: "Branch 1", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, db.Create(&branchEntity.Branch{ID: "b-2", TenantID: "tenant-2", Code: "B2", Name: "Branch 2", Status: branchEntity.BranchStatusActive}).Error)
	require.NoError(t, db.Create(&counterEntity.Counter{ID: "c-1", TenantID: "tenant-1", BranchID: "b-1", Code: "C1", Name: "Counter 1", Status: counterEntity.CounterStatusActive}).Error)
	require.NoError(t, db.Create(&counterEntity.Counter{ID: "c-2", TenantID: "tenant-2", BranchID: "b-2", Code: "C2", Name: "Counter 2", Status: counterEntity.CounterStatusActive}).Error)
	tenant1 := "tenant-1"
	tenant2 := "tenant-2"
	require.NoError(t, db.Create(&userEntity.User{ID: "u-1", OrganizationID: &tenant1, Email: "u1@example.com", Username: "u1", Status: userEntity.UserStatusActive}).Error)
	require.NoError(t, db.Create(&userEntity.User{ID: "u-2", OrganizationID: &tenant1, Email: "u2@example.com", Username: "u2", Status: userEntity.UserStatusActive}).Error)
	require.NoError(t, db.Create(&userEntity.User{ID: "u-3", OrganizationID: &tenant2, Email: "u3@example.com", Username: "u3", Status: userEntity.UserStatusActive}).Error)

	tests := []struct {
		name     string
		category string
		fn       func(t *testing.T)
	}{
		{
			name:     "Positive_Create_List_Delete",
			category: "positive",
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
			name:     "Negative_MissingTenant",
			category: "negative",
			fn: func(t *testing.T) {
				_, err := uc.Create(context.Background(), &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-1", CounterID: "c-1"})
				require.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Negative_InvalidOrMissingRequestFields",
			category: "negative",
			fn: func(t *testing.T) {
				_, err := uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "", UserID: "u-1", CounterID: "c-1"})
				require.ErrorIs(t, err, exception.ErrBadRequest)
				_, err = uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "", CounterID: "c-1"})
				require.ErrorIs(t, err, exception.ErrBadRequest)
				_, err = uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-1", CounterID: ""})
				require.ErrorIs(t, err, exception.ErrBadRequest)
			},
		},
		{
			name:     "Edge_UnassignAlreadyUnassigned",
			category: "edge",
			fn: func(t *testing.T) {
				now := int64(123456789)
				require.NoError(t, db.Create(&entity.OperatorCounterAssignment{ID: "assign-unassigned", TenantID: "tenant-1", BranchID: "b-1", UserID: "u-2", CounterID: "c-1", AssignedAt: now, UnassignedAt: &now}).Error)

				err := uc.Delete(ctx, "assign-unassigned")
				require.NoError(t, err)
			},
		},
		{
			name:     "Vulnerability_CrossTenantDeleteRejected",
			category: "vulnerability",
			fn: func(t *testing.T) {
				otherCtx := database.SetOrganizationContext(context.Background(), "tenant-2")
				require.NoError(t, db.Create(&entity.OperatorCounterAssignment{ID: "assign-cross", TenantID: "tenant-1", BranchID: "b-1", UserID: "u-1", CounterID: "c-1", AssignedAt: int64(123456789)}).Error)

				err := uc.Delete(otherCtx, "assign-cross")
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Vulnerability_CrossTenantCreateRejected",
			category: "vulnerability",
			fn: func(t *testing.T) {
				_, err := uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-3", CounterID: "c-1"})
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Negative_InactiveBranchRejected",
			category: "negative",
			fn: func(t *testing.T) {
				require.NoError(t, db.Create(&branchEntity.Branch{ID: "b-inactive", TenantID: "tenant-1", Code: "B3", Name: "Branch Inactive", Status: branchEntity.BranchStatusInactive}).Error)
				_, err := uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-inactive", UserID: "u-1", CounterID: "c-1"})
				require.ErrorIs(t, err, exception.ErrNotFound, "expected rejection when branch is inactive but implementation might be missing")
			},
		},
		{
			name:     "Negative_CounterBranchMismatchRejected",
			category: "negative",
			fn: func(t *testing.T) {
				_, err := uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-1", CounterID: "c-2"})
				require.ErrorIs(t, err, exception.ErrNotFound)
			},
		},
		{
			name:     "Positive_Audit_CreateDeleteEmitEvents",
			category: "positive",
			fn: func(t *testing.T) {
				capturedAuditReqs = nil
				res, err := uc.Create(ctx, &model.OperatorAssignmentRequest{BranchID: "b-1", UserID: "u-2", CounterID: "c-1"})
				require.NoError(t, err)
				require.NoError(t, uc.Delete(ctx, res.ID))
				require.Len(t, capturedAuditReqs, 2)
				assert.Equal(t, "OPERATOR_ASSIGNMENT_CREATE", capturedAuditReqs[0].Action)
				assert.Equal(t, "OPERATOR_ASSIGNMENT_UNASSIGN", capturedAuditReqs[1].Action)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.category)
			tt.fn(t)
		})
	}
}
