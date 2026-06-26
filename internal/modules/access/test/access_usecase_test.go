package test

import (
	"context"
	"errors"
	"testing"

	"github.com/Roisfaozi/queue-base/internal/modules/access/entity"
	"github.com/Roisfaozi/queue-base/internal/modules/access/model"
	"github.com/Roisfaozi/queue-base/internal/modules/access/test/mocks"
	"github.com/Roisfaozi/queue-base/internal/modules/access/usecase"
	"github.com/Roisfaozi/queue-base/pkg/exception"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type nullWriter struct{}

func (w *nullWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type accessTestDeps struct {
	Repo *mocks.MockAccessRepository
}

func setupAccessTest() (*accessTestDeps, usecase.IAccessUseCase) {
	deps := &accessTestDeps{
		Repo: new(mocks.MockAccessRepository),
	}
	log := logrus.New()
	log.SetOutput(&nullWriter{})
	uc := usecase.NewAccessUseCase(deps.Repo, log)
	return deps, uc
}

func TestCreateAccessRight(t *testing.T) {
	tests := []struct {
		name        string
		category    string
		req         model.CreateAccessRightRequest
		repoErr     error
		matcher     interface{}
		wantErr     error
		wantName    string
		wantDesc    string
	}{
		{name: "Success - Create Valid Access Right", category: "positive", req: model.CreateAccessRightRequest{Name: "view_dashboard", Description: "Allows viewing the main dashboard"}, matcher: mock.AnythingOfType("*entity.AccessRight"), wantName: "view_dashboard", wantDesc: "Allows viewing the main dashboard"},
		{name: "Error - Repository Create Fails", category: "negative", req: model.CreateAccessRightRequest{Name: "error_right"}, matcher: mock.AnythingOfType("*entity.AccessRight"), repoErr: errors.New("db error"), wantErr: errors.New("db error")},
		{name: "Success - Sanitize Inputs", category: "vulnerability", req: model.CreateAccessRightRequest{Name: "<b>Bold Name</b>", Description: "<script>alert('xss')</script>"}, matcher: mock.MatchedBy(func(ar *entity.AccessRight) bool { return ar.Name == "&lt;b&gt;Bold Name&lt;/b&gt;" && ar.Description == "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;" }), wantName: "&lt;b&gt;Bold Name&lt;/b&gt;", wantDesc: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("CreateAccessRight", ctx, tt.matcher).Return(tt.repoErr).Once()

			createdAccessRight, err := uc.CreateAccessRight(ctx, tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				deps.Repo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, createdAccessRight)
			assert.Equal(t, tt.wantName, createdAccessRight.Name)
			assert.Equal(t, tt.wantDesc, createdAccessRight.Description)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestGetAllAccessRights(t *testing.T) {
	tests := []struct {
		name     string
		category string
		entities []*entity.AccessRight
		repoErr  error
		wantErr  error
		wantLen  int
	}{
		{name: "Success - Has Data", category: "positive", entities: []*entity.AccessRight{{ID: "1", Name: "view_dashboard"}, {ID: "2", Name: "edit_settings"}}, wantLen: 2},
		{name: "Success - No Data", category: "edge", entities: []*entity.AccessRight{}, wantLen: 0},
		{name: "Error - Repository Fails", category: "negative", repoErr: errors.New("db error"), wantErr: errors.New("db error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("GetAccessRights", ctx).Return(tt.entities, tt.repoErr).Once()

			results, err := uc.GetAllAccessRights(ctx)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, results)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				deps.Repo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, results)
			assert.Len(t, results.Data, tt.wantLen)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestCreateEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      model.CreateEndpointRequest
		repoErr  error
		wantErr  error
		wantPath string
	}{
		{name: "Success - Create Valid Endpoint", category: "positive", req: model.CreateEndpointRequest{Path: "/api/v1/test", Method: "GET"}, wantPath: "/api/v1/test"},
		{name: "Error - Repository Create Fails", category: "negative", req: model.CreateEndpointRequest{Path: "/error", Method: "POST"}, repoErr: errors.New("db error"), wantErr: errors.New("db error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).Return(tt.repoErr).Once()

			createdEndpoint, err := uc.CreateEndpoint(ctx, tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				deps.Repo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, createdEndpoint)
			assert.Equal(t, tt.wantPath, createdEndpoint.Path)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestLinkEndpointToAccessRight(t *testing.T) {
	tests := []struct {
		name     string
		category string
		req      model.LinkEndpointRequest
		repoErr  error
		wantErr  error
	}{
		{name: "Success - Link Valid IDs", category: "positive", req: model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}},
		{name: "Error - Repository Fails", category: "negative", req: model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}, repoErr: errors.New("db error"), wantErr: errors.New("db error")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("LinkEndpointToAccessRight", ctx, tt.req.AccessRightID, tt.req.EndpointID).Return(tt.repoErr).Once()

			err := uc.LinkEndpointToAccessRight(ctx, tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				deps.Repo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestDeleteAccessRight(t *testing.T) {
	id := "1"

	t.Run("Success - Delete Access Right", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("GetAccessRightByID", ctx, id).Return(&entity.AccessRight{ID: id}, nil).Once()
		deps.Repo.On("DeleteAccessRight", ctx, id).Return(nil).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Not Found", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("GetAccessRightByID", ctx, id).Return(nil, gorm.ErrRecordNotFound).Once()
		err := uc.DeleteAccessRight(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})
}

func TestDeleteEndpoint(t *testing.T) {
	id := "1"

	t.Run("Success - Delete Endpoint", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("DeleteEndpoint", ctx, id).Return(nil).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Not Found (GORM delete behavior)", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		deps.Repo.On("DeleteEndpoint", ctx, id).Return(gorm.ErrRecordNotFound).Once()
		err := uc.DeleteEndpoint(ctx, id)
		assert.ErrorIs(t, err, exception.ErrNotFound)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetEndpointsDynamic(t *testing.T) {
	t.Run("Success - Get Endpoints Dynamically", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Method": {Type: "equals", From: "GET"},
			},
		}
		expectedEndpoints := []*entity.Endpoint{
			{ID: "1", Path: "/api/test", Method: "GET"},
		}
		deps.Repo.On("FindEndpointsDynamic", ctx, filter).Return(expectedEndpoints, int64(1), nil).Once()

		results, total, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "GET", results[0].Method)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{}
		repoError := errors.New("repo error")
		deps.Repo.On("FindEndpointsDynamic", ctx, filter).Return(nil, int64(0), repoError).Once()

		results, total, err := uc.GetEndpointsDynamic(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, int64(0), total)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestAccessUseCase_GetAccessRightsDynamic(t *testing.T) {
	t.Run("Success - Get Access Rights Dynamically", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{
			Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Manage"},
			},
		}
		expectedAccessRights := []*entity.AccessRight{
			{ID: "1", Name: "Manage Users"},
		}
		deps.Repo.On("FindAccessRightsDynamic", ctx, filter).Return(expectedAccessRights, int64(1), nil).Once()

		results, total, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.NoError(t, err)
		assert.Len(t, results.Data, 1)
		assert.Equal(t, int64(1), total)
		assert.Equal(t, "Manage Users", results.Data[0].Name)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Error", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		filter := &querybuilder.DynamicFilter{}
		repoError := errors.New("repo error")
		deps.Repo.On("FindAccessRightsDynamic", ctx, filter).Return(nil, int64(0), repoError).Once()

		results, total, err := uc.GetAccessRightsDynamic(ctx, filter)
		assert.Error(t, err)
		assert.Nil(t, results)
		assert.Equal(t, int64(0), total)
		assert.ErrorIs(t, err, exception.ErrInternalServer)
		deps.Repo.AssertExpectations(t)
	})
}

func TestCreateAccessRight_Sanitization(t *testing.T) {
	deps, uc := setupAccessTest()
	ctx := context.Background()

	// Capture the entity passed to CreateAccessRight to verify sanitization
	var capturedEntity *entity.AccessRight
	deps.Repo.On("CreateAccessRight", ctx, mock.AnythingOfType("*entity.AccessRight")).
		Run(func(args mock.Arguments) {
			capturedEntity = args.Get(1).(*entity.AccessRight)
		}).
		Return(nil).Once()

	req := model.CreateAccessRightRequest{
		Name:        "<b>Bold</b> Right",
		Description: "<script>alert('xss')</script> Description",
	}

	createdAccessRight, err := uc.CreateAccessRight(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdAccessRight)

	// Verify that the entity passed to repo was HTML escaped
	assert.Equal(t, "&lt;b&gt;Bold&lt;/b&gt; Right", capturedEntity.Name)
	assert.Equal(t, "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; Description", capturedEntity.Description)

	deps.Repo.AssertExpectations(t)
}

func TestCreateEndpoint_Sanitization(t *testing.T) {
	deps, uc := setupAccessTest()
	ctx := context.Background()

	var capturedEntity *entity.Endpoint
	deps.Repo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).
		Run(func(args mock.Arguments) {
			capturedEntity = args.Get(1).(*entity.Endpoint)
		}).
		Return(nil).Once()

	req := model.CreateEndpointRequest{
		Path:   "/api/v1/test/<script>alert(1)</script>",
		Method: "GET",
	}

	createdEndpoint, err := uc.CreateEndpoint(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, createdEndpoint)

	// Verify that the entity passed to repo was HTML escaped
	assert.Equal(t, "/api/v1/test/&lt;script&gt;alert(1)&lt;/script&gt;", capturedEntity.Path)

	deps.Repo.AssertExpectations(t)
}

func TestCreateEndpoint_DuplicateDetection(t *testing.T) {
	deps, uc := setupAccessTest()

	req := model.CreateEndpointRequest{
		Path:   "/api/users",
		Method: "GET",
	}

	expectedErr := exception.ErrConflict
	deps.Repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e interface{}) bool {
		return true
	})).Return(expectedErr).Once()

	resp, err := uc.CreateEndpoint(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, expectedErr, err)

	deps.Repo.On("CreateEndpoint", mock.Anything, mock.Anything).Return(nil).Once()

}

func TestLinkEndpointToAccessRight_Duplicate(t *testing.T) {
	deps, uc := setupAccessTest()

	req := model.LinkEndpointRequest{
		AccessRightID: "access-right-uuid",
		EndpointID:    "endpoint-uuid",
	}

	// Case: Duplicate link
	deps.Repo.On("LinkEndpointToAccessRight", mock.Anything, req.AccessRightID, req.EndpointID).
		Return(errors.New("duplicate entry")) // Simulate DB error

	err := uc.LinkEndpointToAccessRight(context.Background(), req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "duplicate")
}

func TestUnlinkEndpointFromAccessRight(t *testing.T) {
	t.Run("Success - Unlink Valid IDs", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		deps.Repo.On("UnlinkEndpointFromAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(nil).Once()
		err := uc.UnlinkEndpointFromAccessRight(ctx, req)
		assert.NoError(t, err)
		deps.Repo.AssertExpectations(t)
	})

	t.Run("Error - Repository Fails", func(t *testing.T) {
		deps, uc := setupAccessTest()
		ctx := context.Background()

		req := model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"}
		repoErr := errors.New("db error")
		deps.Repo.On("UnlinkEndpointFromAccessRight", ctx, req.AccessRightID, req.EndpointID).Return(repoErr).Once()

		err := uc.UnlinkEndpointFromAccessRight(ctx, req)
		assert.Error(t, err)
		assert.Equal(t, repoErr, err)
		deps.Repo.AssertExpectations(t)
	})
}
