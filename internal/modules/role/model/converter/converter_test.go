package converter_test

import (
	"testing"
	"time"

	"github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/role/model/converter"
	"github.com/stretchr/testify/assert"
)

func TestRoleToResponse(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name string
		role *entity.Role
	}{
		{name: "single role", role: &entity.Role{ID: "role-1", Name: "Admin", Description: "Administrator", CreatedAt: now, UpdatedAt: now}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := converter.RoleToResponse(tt.role)
			assert.NotNil(t, res)
			assert.Equal(t, tt.role.ID, res.ID)
			assert.Equal(t, tt.role.Name, res.Name)
			assert.Equal(t, tt.role.Description, res.Description)
		})
	}
}

func TestRolesToResponse(t *testing.T) {
	now := time.Now().Unix()
	tests := []struct {
		name    string
		roles   []*entity.Role
		wantNil bool
		wantLen int
	}{
		{name: "has data", roles: []*entity.Role{{ID: "role-1", Name: "Admin", Description: "Administrator", CreatedAt: now, UpdatedAt: now}, {ID: "role-2", Name: "User", Description: "Standard User", CreatedAt: now, UpdatedAt: now}}, wantLen: 2},
		{name: "empty", roles: nil, wantNil: true, wantLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := converter.RolesToResponse(tt.roles)
			if tt.wantNil {
				assert.Nil(t, res)
				assert.Len(t, res, tt.wantLen)
				return
			}
			assert.NotNil(t, res)
			assert.Len(t, res, tt.wantLen)
			assert.Equal(t, tt.roles[0].ID, res[0].ID)
			assert.Equal(t, tt.roles[1].ID, res[1].ID)
		})
	}
}
