//go:build integration
// +build integration

package modules

import (
	"context"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/access/model"
	"github.com/Roisfaozi/queue-base/internal/modules/access/repository"
	"github.com/Roisfaozi/queue-base/internal/modules/access/usecase"
	"github.com/Roisfaozi/queue-base/pkg"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/Roisfaozi/queue-base/tests/integration/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupAccessIntegration(env *setup.TestEnvironment) usecase.IAccessUseCase {
	repo := repository.NewAccessRepository(env.DB, env.Logger)
	return usecase.NewAccessUseCase(repo, env.Logger)
}

func TestAccessIntegration_CreateAccessRight(t *testing.T) {
	tests := []struct {
		name       string
		req        model.CreateAccessRightRequest
		prepare    func(t *testing.T, uc usecase.IAccessUseCase, req model.CreateAccessRightRequest)
		assertions func(t *testing.T, ar *model.AccessRightResponse, err error, req model.CreateAccessRightRequest)
	}{
		{
			name: "Success",
			req: model.CreateAccessRightRequest{
				Name:        "User Management",
				Description: "Manage users",
			},
			assertions: func(t *testing.T, ar *model.AccessRightResponse, err error, req model.CreateAccessRightRequest) {
				require.NoError(t, err)
				assert.NotEmpty(t, ar.ID)
				assert.Equal(t, req.Name, ar.Name)
			},
		},
		{
			name: "Fail_Duplicate",
			req:  model.CreateAccessRightRequest{Name: "Dup", Description: "d"},
			prepare: func(t *testing.T, uc usecase.IAccessUseCase, req model.CreateAccessRightRequest) {
				_, err := uc.CreateAccessRight(context.Background(), req)
				require.NoError(t, err)
			},
			assertions: func(t *testing.T, _ *model.AccessRightResponse, err error, _ model.CreateAccessRightRequest) {
				assert.Error(t, err)
			},
		},
		{
			name: "Security_SQLInjectionPrevention",
			req:  model.CreateAccessRightRequest{Name: "name' OR '1'='1"},
			assertions: func(t *testing.T, ar *model.AccessRightResponse, err error, req model.CreateAccessRightRequest) {
				if err == nil {
					assert.Equal(t, pkg.SanitizeString(req.Name), ar.Name)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			uc := setupAccessIntegration(env)
			if tt.prepare != nil {
				tt.prepare(t, uc, tt.req)
			}

			ar, err := uc.CreateAccessRight(context.Background(), tt.req)
			tt.assertions(t, ar, err, tt.req)
		})
	}
}

func TestAccessIntegration_DeleteAccessRight(t *testing.T) {
	tests := []struct {
		name       string
		prepare    func(t *testing.T, uc usecase.IAccessUseCase) string
		assertions func(t *testing.T, err error)
	}{
		{
			name: "Success",
			prepare: func(t *testing.T, uc usecase.IAccessUseCase) string {
				ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Del", Description: "d"})
				require.NoError(t, err)
				return ar.ID
			},
			assertions: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
		{
			name: "Fail_NotFound",
			prepare: func(t *testing.T, uc usecase.IAccessUseCase) string {
				return "ghost-id"
			},
			assertions: func(t *testing.T, err error) {
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			uc := setupAccessIntegration(env)
			id := tt.prepare(t, uc)
			err := uc.DeleteAccessRight(context.Background(), id)
			tt.assertions(t, err)
		})
	}
}

func TestAccessIntegration_EndpointLinkFlow(t *testing.T) {
	tests := []struct {
		name     string
		category string
		run      func(t *testing.T, uc usecase.IAccessUseCase)
	}{
		{
			name:     "CreateEndpoint_LinkToAccessRight",
			category: "positive",
			run: func(t *testing.T, uc usecase.IAccessUseCase) {
				ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Roles", Description: "d"})
				require.NoError(t, err)

				ep, err := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/api/v1/roles", Method: "GET"})
				require.NoError(t, err)

				err = uc.LinkEndpointToAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
				require.NoError(t, err)

				list, err := uc.GetAllAccessRights(context.Background())
				require.NoError(t, err)

				foundLinked := false
				for _, item := range list.Data {
					if item.ID != ar.ID {
						continue
					}
					for _, endpoint := range item.Endpoints {
						if endpoint.ID == ep.ID {
							foundLinked = true
						}
					}
				}
				assert.True(t, foundLinked)
			},
		},
		{
			name:     "LinkEndpoint_Negative_DuplicateLink",
			category: "negative",
			run: func(t *testing.T, uc usecase.IAccessUseCase) {
				ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "DupLink", Description: "d"})
				require.NoError(t, err)

				ep, err := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/api/v1/duplink", Method: "GET"})
				require.NoError(t, err)

				err = uc.LinkEndpointToAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
				require.NoError(t, err)

				err = uc.LinkEndpointToAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
				t.Logf("Duplicate link error (expected idempotent or conflict): %v", err)
			},
		},
		{
			name:     "UnlinkEndpointFromAccessRight_Success",
			category: "positive",
			run: func(t *testing.T, uc usecase.IAccessUseCase) {
				ar, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Unlink", Description: "d"})
				require.NoError(t, err)

				ep, err := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/api/v1/unlink", Method: "GET"})
				require.NoError(t, err)

				err = uc.LinkEndpointToAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
				require.NoError(t, err)

				err = uc.UnlinkEndpointFromAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: ar.ID, EndpointID: ep.ID})
				require.NoError(t, err)

				list, err := uc.GetAllAccessRights(context.Background())
				require.NoError(t, err)

				foundLinked := false
				for _, item := range list.Data {
					if item.ID != ar.ID {
						continue
					}
					for _, endpoint := range item.Endpoints {
						if endpoint.ID == ep.ID {
							foundLinked = true
						}
					}
				}
				assert.False(t, foundLinked)
			},
		},
		{
			name:     "UnlinkEndpointFromAccessRight_Negative_NonExistent",
			category: "negative",
			run: func(t *testing.T, uc usecase.IAccessUseCase) {
				err := uc.UnlinkEndpointFromAccessRight(context.Background(), model.LinkEndpointRequest{AccessRightID: "non-existent", EndpointID: "non-existent"})
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			uc := setupAccessIntegration(env)
			tt.run(t, uc)
		})
	}
}

func TestAccessIntegration_DeleteEndpoint(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, uc usecase.IAccessUseCase)
	}{
		{
			name: "Success",
			run: func(t *testing.T, uc usecase.IAccessUseCase) {
				ep, err := uc.CreateEndpoint(context.Background(), model.CreateEndpointRequest{Path: "/temp", Method: "POST"})
				require.NoError(t, err)

				err = uc.DeleteEndpoint(context.Background(), ep.ID)
				assert.NoError(t, err)
			},
		},
		{
			name: "Idempotent_NotFound",
			run: func(t *testing.T, uc usecase.IAccessUseCase) {
				err := uc.DeleteEndpoint(context.Background(), "nonexistent-id")
				assert.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := setup.SetupIntegrationEnvironment(t)
			defer env.Cleanup()

			uc := setupAccessIntegration(env)
			tt.run(t, uc)
		})
	}
}

func TestAccessIntegration_DynamicSearch_AccessRights(t *testing.T) {
	env := setup.SetupIntegrationEnvironment(t)
	defer env.Cleanup()

	uc := setupAccessIntegration(env)

	_, err := uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "User Management"})
	require.NoError(t, err)
	_, err = uc.CreateAccessRight(context.Background(), model.CreateAccessRightRequest{Name: "Audit Logs"})
	require.NoError(t, err)

	filter := &querybuilder.DynamicFilter{
		Filter: map[string]querybuilder.Filter{
			"name": {Type: "contains", From: "User"},
		},
	}
	list, _, err := uc.GetAccessRightsDynamic(context.Background(), filter)
	require.NoError(t, err)
	assert.Len(t, list.Data, 1)
	assert.Equal(t, "User Management", list.Data[0].Name)
}
