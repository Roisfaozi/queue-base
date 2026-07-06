package usecase

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type branchLogoRepo struct {
	branch     *entity.Branch
	tenantLogo string
}

func (r *branchLogoRepo) Create(_ context.Context, branch *entity.Branch) error {
	r.branch = branch
	return nil
}
func (r *branchLogoRepo) FindByID(_ context.Context, tenantID, branchID string) (*entity.Branch, error) {
	if r.branch == nil || r.branch.TenantID != tenantID || r.branch.ID != branchID {
		return nil, exception.ErrNotFound
	}
	return r.branch, nil
}
func (r *branchLogoRepo) FindAll(_ context.Context, tenantID string) ([]*entity.Branch, error) {
	return []*entity.Branch{r.branch}, nil
}
func (r *branchLogoRepo) Update(_ context.Context, branch *entity.Branch) error {
	r.branch = branch
	return nil
}
func (r *branchLogoRepo) Delete(_ context.Context, tenantID, branchID string) error { return nil }
func (r *branchLogoRepo) TenantLogoAssetID(_ context.Context, tenantID string) (string, error) {
	return r.tenantLogo, nil
}

func TestBranchActivationLogoFallback(t *testing.T) {
	ctx := database.SetOrganizationContext(context.Background(), "tenant-1")
	active := entity.BranchStatusActive
	address := "Jl. Test"
	city := "Jakarta"
	province := "DKI"
	phone := "021"
	timezone := "Asia/Jakarta"

	tests := []struct {
		name       string
		tenantLogo string
		wantErr    error
	}{
		{name: "Positive_ActivateWithoutBranchLogoWhenTenantLogoExists", tenantLogo: "tenant-logo"},
		{name: "Negative_RejectActivateWithoutAnyLogo", tenantLogo: "", wantErr: exception.ErrBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &branchLogoRepo{tenantLogo: tt.tenantLogo, branch: &entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusInactive}}
			uc := NewBranchUseCase(repo)
			res, err := uc.UpdateBranch(ctx, "branch-1", &model.UpdateBranchRequest{Address: &address, City: &city, Province: &province, Phone: &phone, Timezone: &timezone, Status: &active})

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, entity.BranchStatusActive, res.Status)
		})
	}
}
