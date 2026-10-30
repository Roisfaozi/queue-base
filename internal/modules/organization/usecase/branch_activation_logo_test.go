package usecase_test

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/organization/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/model"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/organization/usecase"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

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
			branch := &entity.Branch{ID: "branch-1", TenantID: "tenant-1", Code: "MAIN", Name: "Main", Status: entity.BranchStatusInactive}
			mockRepo := new(mocks.MockBranchLogoRepository)
			mockRepo.On("FindByID", ctx, "tenant-1", "branch-1").Return(branch, nil)
			mockRepo.On("TenantLogoAssetID", ctx, "tenant-1").Return(tt.tenantLogo, nil)
			mockRepo.On("Update", ctx, mock.MatchedBy(func(b *entity.Branch) bool {
				b.Status = entity.BranchStatusActive
				return true
			})).Return(nil).Maybe()

			uc := usecase.NewBranchUseCase(mockRepo)
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
