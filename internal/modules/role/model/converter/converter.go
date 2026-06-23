package converter

import (
	"github.com/Roisfaozi/queue-base/internal/modules/role/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/role/model"
)

func RoleToResponse(role *entity.Role) *model.RoleResponse {
	return &model.RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
}

func RolesToResponse(roles []*entity.Role) []model.RoleResponse {
	var roleResponses []model.RoleResponse
	for _, r := range roles {
		roleResponses = append(roleResponses, *RoleToResponse(r))
	}
	return roleResponses
}
