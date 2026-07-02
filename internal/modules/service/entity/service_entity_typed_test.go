package entity

import (
	"reflect"
	"testing"
)

func TestBranchServiceTableName(t *testing.T) {
	if got, want := (BranchService{}).TableName(), "branch_services"; got != want {
		t.Fatalf("TableName() = %q, want %q", got, want)
	}
}

func TestTypedServiceFieldsExist(t *testing.T) {
	tests := []struct {
		name      string
		typeValue any
		fieldName string
	}{
		{name: "service type", typeValue: Service{}, fieldName: "Type"},
		{name: "service duration", typeValue: Service{}, fieldName: "DefaultEstimatedDuration"},
		{name: "branch service sort order", typeValue: BranchService{}, fieldName: "SortOrder"},
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
