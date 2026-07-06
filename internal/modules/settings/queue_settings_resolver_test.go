package settings

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/settings/entity"
	"github.com/Roisfaozi/queue-base/pkg/database"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newResolverTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(
		&entity.TenantQueueSetting{},
		&entity.BranchQueueSetting{},
		&entity.BranchServiceQueueSetting{},
		&entity.CounterQueueSetting{},
	))
	return db
}

func TestQueueSettingsResolver_Resolve(t *testing.T) {
	db := newResolverTestDB(t)

	require.NoError(t, db.Create(&entity.TenantQueueSetting{ID: "ts-1", TenantID: "t-1", QueueResetTime: "04:00", DefaultTicketPrefix: "A"}).Error)
	val0500 := "05:00"
	valPrefixC := "C"
	valPrefixNil := (*string)(nil)
	autoCallNext := true
	require.NoError(t, db.Create(&entity.BranchQueueSetting{ID: "bs-1", TenantID: "t-1", BranchID: "b-1", QueueResetTime: &val0500}).Error)
	require.NoError(t, db.Create(&entity.BranchQueueSetting{ID: "bs-2", TenantID: "t-1", BranchID: "b-2", TicketPrefix: valPrefixNil}).Error)
	require.NoError(t, db.Create(&entity.BranchServiceQueueSetting{ID: "bss-1", TenantID: "t-1", BranchID: "b-1", BranchServiceID: "bsvc-1", DefaultEstimatedDuration: &[]int{20}[0], AutoCallNext: &autoCallNext}).Error)
	require.NoError(t, db.Create(&entity.CounterQueueSetting{ID: "cs-1", TenantID: "t-1", CounterID: "c-1", TicketPrefix: &valPrefixC}).Error)

	ctx := database.SetOrganizationContext(context.Background(), "t-1")
	resolver := NewQueueSettingsResolver(db)

	tests := []struct {
		name      string
		key       string
		branchID  string
		serviceID string
		counterID string
		want      string
	}{
		{name: "Positive_InheritsTenantDefault", key: "ticket_prefix", branchID: "b-1", want: "A"},
		{name: "Edge_NullBranchOverrideInheritsTenantDefault", key: "ticket_prefix", branchID: "b-2", want: "A"},
		{name: "Positive_ResolvesBranchOverride", key: "queue_reset_time", branchID: "b-1", want: "05:00"},
		{name: "Positive_ResolvesCounterOverride", key: "ticket_prefix", branchID: "b-1", counterID: "c-1", want: "C"},
		{name: "Positive_ResolvesBranchServiceOverride", key: "default_estimated_duration", branchID: "b-1", serviceID: "bsvc-1", want: "20"},
		{name: "Positive_ResolvesAutoCallNextFromTypedTable", key: "auto_call_next", branchID: "b-1", serviceID: "bsvc-1", want: "true"},
		{name: "Negative_NoGenericFallback", key: "custom_theme_color", branchID: "b-1", want: ""},
		{name: "Negative_LogoFallbackMiss", key: "logo_url", branchID: "b-1", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolver.Resolve(ctx, tt.key, tt.branchID, tt.serviceID, tt.counterID)
			if tt.name == "Negative_NoGenericFallback" || tt.name == "Negative_LogoFallbackMiss" {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
