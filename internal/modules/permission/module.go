package permission

import (
	accessRepository "github.com/Roisfaozi/queue-base/internal/modules/access/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/audit"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/delivery/http"
	"github.com/Roisfaozi/queue-base/internal/modules/permission/usecase"
	roleRepository "github.com/Roisfaozi/queue-base/internal/modules/role/repository"
	userRepository "github.com/Roisfaozi/queue-base/internal/modules/user/repository"
	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
)

type IEnforcer = usecase.IEnforcer

type PermissionModule struct {
	PermissionController *http.PermissionController
	PermissionUseCase    usecase.IPermissionUseCase
}

func NewPermissionModule(
	enforcer usecase.IEnforcer,
	validate *validator.Validate,
	log *logrus.Logger,
	roleRepo roleRepository.RoleRepository,
	userRepo userRepository.UserRepository,
	accessRepo accessRepository.AccessRepository,
	auditModule *audit.AuditModule,
) *PermissionModule {

	permissionUseCase := usecase.NewPermissionUseCase(enforcer, log, roleRepo, userRepo, accessRepo, auditModule.AuditUseCase)

	permissionController := http.NewPermissionController(permissionUseCase, log, validate)

	return &PermissionModule{
		PermissionController: permissionController,
		PermissionUseCase:    permissionUseCase,
	}
}

func (m *PermissionModule) Controller() *http.PermissionController {
	return m.PermissionController
}
