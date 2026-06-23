package usecase

import (
	"context"

	"github.com/Roisfaozi/queue-base/internal/modules/role/model"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
)

type RoleUseCase interface {
	Create(ctx context.Context, request *model.CreateRoleRequest) (*model.RoleResponse, error)
	Update(ctx context.Context, id string, request *model.UpdateRoleRequest) (*model.RoleResponse, error)
	GetAll(ctx context.Context) ([]model.RoleResponse, error)
	GetAllRolesDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]model.RoleResponse, error)
	Delete(ctx context.Context, id string) error
}
