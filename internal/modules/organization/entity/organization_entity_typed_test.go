package entity

import (
	"reflect"
	"testing"
)

func TestTypedOrganizationAndBranchFieldsExist(t *testing.T) {
	tests := []struct {
		name      string
		typeValue any
		fieldName string
	}{
		{name: "organization code", typeValue: Organization{}, fieldName: "Code"},
		{name: "organization timezone", typeValue: Organization{}, fieldName: "Timezone"},
		{name: "branch running text", typeValue: Branch{}, fieldName: "RunningText"},
		{name: "branch logo asset", typeValue: Branch{}, fieldName: "LogoAssetID"},
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
