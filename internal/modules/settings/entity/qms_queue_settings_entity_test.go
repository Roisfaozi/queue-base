package entity

import (
	"reflect"
	"testing"
)

func TestTypedQueueSettingTableNames(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "tenant", got: TenantQueueSetting{}.TableName(), want: "tenant_queue_settings"},
		{name: "branch", got: BranchQueueSetting{}.TableName(), want: "branch_queue_settings"},
		{name: "service", got: ServiceQueueSetting{}.TableName(), want: "service_queue_settings"},
		{name: "counter", got: CounterQueueSetting{}.TableName(), want: "counter_queue_settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("TableName() = %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestTypedQueueSettingFields(t *testing.T) {
	tests := []struct {
		name      string
		typeValue any
		fieldName string
	}{
		{name: "tenant queue reset time", typeValue: TenantQueueSetting{}, fieldName: "QueueResetTime"},
		{name: "branch ticket prefix", typeValue: BranchQueueSetting{}, fieldName: "TicketPrefix"},
		{name: "service require counter", typeValue: ServiceQueueSetting{}, fieldName: "RequireCounter"},
		{name: "counter numbering strategy", typeValue: CounterQueueSetting{}, fieldName: "NumberingStrategy"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typ := reflect.TypeOf(tt.typeValue)
			if _, ok := typ.FieldByName(tt.fieldName); !ok {
				t.Fatalf("missing field %s on %s", tt.fieldName, typ.Name())
			}
		})
	}
}
