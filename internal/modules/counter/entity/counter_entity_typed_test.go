package entity

import (
	"reflect"
	"testing"
)

func TestTypedCounterFieldsExist(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
	}{
		{name: "branch service id", fieldName: "BranchServiceID"},
		{name: "display name", fieldName: "DisplayName"},
	}

	typ := reflect.TypeOf(Counter{})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, ok := typ.FieldByName(tt.fieldName); !ok {
				t.Fatalf("missing field %s on %s", tt.fieldName, typ.Name())
			}
		})
	}
}
