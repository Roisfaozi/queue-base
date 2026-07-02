package settings

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	settingsModel "github.com/Roisfaozi/queue-base/internal/modules/settings/model"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type stubSettingsUseCase struct {
	val string
	err error
}

func (s *stubSettingsUseCase) ResolveSetting(ctx context.Context, req *settingsModel.ResolveSettingRequest) (*settingsModel.SettingResponse, error) {
	if s.err != nil {
		return nil, s.err
	}
	return &settingsModel.SettingResponse{Value: s.val}, nil
}
func (s *stubSettingsUseCase) CreateSetting(ctx context.Context, req *settingsModel.CreateSettingRequest) (*settingsModel.SettingResponse, error) {
	return nil, nil
}
func (s *stubSettingsUseCase) GetSetting(ctx context.Context, settingID string) (*settingsModel.SettingResponse, error) {
	return nil, nil
}
func (s *stubSettingsUseCase) UpdateSetting(ctx context.Context, settingID string, req *settingsModel.UpdateSettingRequest) (*settingsModel.SettingResponse, error) {
	return nil, nil
}
func (s *stubSettingsUseCase) DeleteSetting(ctx context.Context, settingID string) error { return nil }

func newResolverTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&entity.TenantQueueSetting{},
		&entity.BranchQueueSetting{},
		&entity.ServiceQueueSetting{},
		&entity.CounterQueueSetting{},
	))
	return db
}

func TestQueueSettingsResolver_Resolve(t *testing.T) {
	db := newResolverTestDB(t)

	// Seed tenant default
	require.NoError(t, db.Create(&entity.TenantQueueSetting{
		ID:                  "ts-1",
		TenantID:            "t-1",
		QueueResetTime:      "04:00",
		DefaultTicketPrefix: "A",
	}).Error)

	val0500 := "05:00"
	valPrefixC := "C"

	// Seed branch override
	require.NoError(t, db.Create(&entity.BranchQueueSetting{
		ID:             "bs-1",
		TenantID:       "t-1",
		BranchID:       "b-1",
		QueueResetTime: &val0500,
		// TicketPrefix is null, should inherit
	}).Error)

	// Seed counter override
	require.NoError(t, db.Create(&entity.CounterQueueSetting{
		ID:           "cs-1",
		TenantID:     "t-1",
		CounterID:    "c-1",
		TicketPrefix: &valPrefixC,
	}).Error)

	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	uc := &stubSettingsUseCase{val: "generic_val"}
	resolver := NewQueueSettingsResolver(db, uc)

	tests := []struct {
		name      string
		key       string
		branchID  string
		serviceID string
		counterID string
		want      string
	}{
		{
			name:      "Positive_InheritsTenantDefault",
			key:       "ticket_prefix",
			branchID:  "b-1",
			serviceID: "",
			counterID: "",
			want:      "A", // from tenant, branch is null
		},
		{
			name:      "Positive_ResolvesBranchOverride",
			key:       "queue_reset_time",
			branchID:  "b-1",
			serviceID: "",
			counterID: "",
			want:      "05:00", // from branch
		},
		{
			name:      "Positive_ResolvesCounterOverride",
			key:       "ticket_prefix",
			branchID:  "b-1",
			serviceID: "",
			counterID: "c-1",
			want:      "C", // from counter
		},
		{
			name:      "Positive_FallbackToGenericForNonCoreKey",
			key:       "custom_theme_color",
			branchID:  "b-1",
			serviceID: "",
			counterID: "",
			want:      "generic_val", // from stubSettingsUseCase
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.Resolve(ctx, tt.key, tt.branchID, tt.serviceID, tt.counterID)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
