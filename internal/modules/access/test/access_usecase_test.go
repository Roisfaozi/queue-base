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
		name     string
		category string
		req      model.CreateAccessRightRequest
		repoErr  error
		matcher  interface{}
		wantErr  error
		wantName string
		wantDesc string
	}{
		{name: "Success - Create Valid Access Right", category: "positive", req: model.CreateAccessRightRequest{Name: "view_dashboard", Description: "Allows viewing the main dashboard"}, matcher: mock.AnythingOfType("*entity.AccessRight"), wantName: "view_dashboard", wantDesc: "Allows viewing the main dashboard"},
		{name: "Error - Repository Create Fails", category: "negative", req: model.CreateAccessRightRequest{Name: "error_right"}, matcher: mock.AnythingOfType("*entity.AccessRight"), repoErr: errors.New("db error"), wantErr: errors.New("db error")},
		{name: "Success - Sanitize Inputs", category: "vulnerability", req: model.CreateAccessRightRequest{Name: "<b>Bold Name</b>", Description: "<script>alert('xss')</script>"}, matcher: mock.MatchedBy(func(ar *entity.AccessRight) bool {
			return ar.Name == "&lt;b&gt;Bold Name&lt;/b&gt;" && ar.Description == "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"
		}), wantName: "&lt;b&gt;Bold Name&lt;/b&gt;", wantDesc: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"},
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
	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockAccessRepository, context.Context, string)
		wantErr   error
	}{
		{
			name: "Success - Delete Access Right",
			id:   "1",
			setupMock: func(repo *mocks.MockAccessRepository, ctx context.Context, id string) {
				repo.On("GetAccessRightByID", ctx, id).Return(&entity.AccessRight{ID: id}, nil).Once()
				repo.On("DeleteAccessRight", ctx, id).Return(nil).Once()
			},
		},
		{
			name: "Error - Not Found",
			id:   "1",
			setupMock: func(repo *mocks.MockAccessRepository, ctx context.Context, id string) {
				repo.On("GetAccessRightByID", ctx, id).Return(nil, gorm.ErrRecordNotFound).Once()
			},
			wantErr: exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			tt.setupMock(deps.Repo, ctx, tt.id)

			err := uc.DeleteAccessRight(ctx, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestDeleteEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		id        string
		setupMock func(*mocks.MockAccessRepository, context.Context, string)
		wantErr   error
	}{
		{
			name: "Success - Delete Endpoint",
			id:   "1",
			setupMock: func(repo *mocks.MockAccessRepository, ctx context.Context, id string) {
				repo.On("DeleteEndpoint", ctx, id).Return(nil).Once()
			},
		},
		{
			name: "Error - Not Found (GORM delete behavior)",
			id:   "1",
			setupMock: func(repo *mocks.MockAccessRepository, ctx context.Context, id string) {
				repo.On("DeleteEndpoint", ctx, id).Return(gorm.ErrRecordNotFound).Once()
			},
			wantErr: exception.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			tt.setupMock(deps.Repo, ctx, tt.id)

			err := uc.DeleteEndpoint(ctx, tt.id)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestAccessUseCase_GetEndpointsDynamic(t *testing.T) {
	tests := []struct {
		name       string
		filter     *querybuilder.DynamicFilter
		results    []*entity.Endpoint
		total      int64
		repoErr    error
		wantErr    error
		wantLen    int
		wantMethod string
	}{
		{
			name: "Success - Get Endpoints Dynamically",
			filter: &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{
				"Method": {Type: "equals", From: "GET"},
			}},
			results:    []*entity.Endpoint{{ID: "1", Path: "/api/test", Method: "GET"}},
			total:      1,
			wantLen:    1,
			wantMethod: "GET",
		},
		{
			name:    "Error - Repository Error",
			filter:  &querybuilder.DynamicFilter{},
			repoErr: errors.New("repo error"),
			wantErr: exception.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("FindEndpointsDynamic", ctx, tt.filter).Return(tt.results, tt.total, tt.repoErr).Once()

			results, total, err := uc.GetEndpointsDynamic(ctx, tt.filter)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, results)
				assert.Equal(t, int64(0), total)
				assert.ErrorIs(t, err, tt.wantErr)
				deps.Repo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, results, tt.wantLen)
			assert.Equal(t, tt.total, total)
			assert.Equal(t, tt.wantMethod, results[0].Method)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestAccessUseCase_GetAccessRightsDynamic(t *testing.T) {
	tests := []struct {
		name     string
		filter   *querybuilder.DynamicFilter
		results  []*entity.AccessRight
		total    int64
		repoErr  error
		wantErr  error
		wantLen  int
		wantName string
	}{
		{
			name: "Success - Get Access Rights Dynamically",
			filter: &querybuilder.DynamicFilter{Filter: map[string]querybuilder.Filter{
				"Name": {Type: "contains", From: "Manage"},
			}},
			results:  []*entity.AccessRight{{ID: "1", Name: "Manage Users"}},
			total:    1,
			wantLen:  1,
			wantName: "Manage Users",
		},
		{
			name:    "Error - Repository Error",
			filter:  &querybuilder.DynamicFilter{},
			repoErr: errors.New("repo error"),
			wantErr: exception.ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("FindAccessRightsDynamic", ctx, tt.filter).Return(tt.results, tt.total, tt.repoErr).Once()

			results, total, err := uc.GetAccessRightsDynamic(ctx, tt.filter)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Nil(t, results)
				assert.Equal(t, int64(0), total)
				assert.ErrorIs(t, err, tt.wantErr)
				deps.Repo.AssertExpectations(t)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, results.Data, tt.wantLen)
			assert.Equal(t, tt.total, total)
			assert.Equal(t, tt.wantName, results.Data[0].Name)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestCreateAccessRight_Sanitization(t *testing.T) {
	tests := []struct {
		name            string
		req             model.CreateAccessRightRequest
		wantName        string
		wantDescription string
	}{
		{
			name: "sanitizes html",
			req: model.CreateAccessRightRequest{
				Name:        "<b>Bold</b> Right",
				Description: "<script>alert('xss')</script> Description",
			},
			wantName:        "&lt;b&gt;Bold&lt;/b&gt; Right",
			wantDescription: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt; Description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			var capturedEntity *entity.AccessRight
			deps.Repo.On("CreateAccessRight", ctx, mock.AnythingOfType("*entity.AccessRight")).
				Run(func(args mock.Arguments) {
					capturedEntity = args.Get(1).(*entity.AccessRight)
				}).
				Return(nil).Once()

			createdAccessRight, err := uc.CreateAccessRight(ctx, tt.req)
			assert.NoError(t, err)
			assert.NotNil(t, createdAccessRight)
			assert.Equal(t, tt.wantName, capturedEntity.Name)
			assert.Equal(t, tt.wantDescription, capturedEntity.Description)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestCreateEndpoint_Sanitization(t *testing.T) {
	tests := []struct {
		name     string
		req      model.CreateEndpointRequest
		wantPath string
	}{
		{
			name: "sanitizes html path",
			req: model.CreateEndpointRequest{
				Path:   "/api/v1/test/<script>alert(1)</script>",
				Method: "GET",
			},
			wantPath: "/api/v1/test/&lt;script&gt;alert(1)&lt;/script&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()

			var capturedEntity *entity.Endpoint
			deps.Repo.On("CreateEndpoint", ctx, mock.AnythingOfType("*entity.Endpoint")).
				Run(func(args mock.Arguments) {
					capturedEntity = args.Get(1).(*entity.Endpoint)
				}).
				Return(nil).Once()

			createdEndpoint, err := uc.CreateEndpoint(ctx, tt.req)
			assert.NoError(t, err)
			assert.NotNil(t, createdEndpoint)
			assert.Equal(t, tt.wantPath, capturedEntity.Path)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestCreateEndpoint_DuplicateDetection(t *testing.T) {
	tests := []struct {
		name    string
		req     model.CreateEndpointRequest
		repoErr error
		wantErr error
	}{
		{
			name:    "duplicate detection",
			req:     model.CreateEndpointRequest{Path: "/api/users", Method: "GET"},
			repoErr: exception.ErrConflict,
			wantErr: exception.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			deps.Repo.On("CreateEndpoint", mock.Anything, mock.MatchedBy(func(e interface{}) bool {
				return true
			})).Return(tt.repoErr).Once()

			resp, err := uc.CreateEndpoint(context.Background(), tt.req)

			assert.Error(t, err)
			assert.Nil(t, resp)
			assert.Equal(t, tt.wantErr, err)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestLinkEndpointToAccessRight_Duplicate(t *testing.T) {
	tests := []struct {
		name    string
		req     model.LinkEndpointRequest
		repoErr error
		wantMsg string
	}{
		{
			name:    "duplicate link",
			req:     model.LinkEndpointRequest{AccessRightID: "access-right-uuid", EndpointID: "endpoint-uuid"},
			repoErr: errors.New("duplicate entry"),
			wantMsg: "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			deps.Repo.On("LinkEndpointToAccessRight", mock.Anything, tt.req.AccessRightID, tt.req.EndpointID).Return(tt.repoErr).Once()

			err := uc.LinkEndpointToAccessRight(context.Background(), tt.req)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMsg)
			deps.Repo.AssertExpectations(t)
		})
	}
}

func TestUnlinkEndpointFromAccessRight(t *testing.T) {
	tests := []struct {
		name    string
		req     model.LinkEndpointRequest
		repoErr error
		wantErr error
	}{
		{
			name: "Success - Unlink Valid IDs",
			req:  model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"},
		},
		{
			name:    "Error - Repository Fails",
			req:     model.LinkEndpointRequest{AccessRightID: "1", EndpointID: "2"},
			repoErr: errors.New("db error"),
			wantErr: errors.New("db error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps, uc := setupAccessTest()
			ctx := context.Background()
			deps.Repo.On("UnlinkEndpointFromAccessRight", ctx, tt.req.AccessRightID, tt.req.EndpointID).Return(tt.repoErr).Once()

			err := uc.UnlinkEndpointFromAccessRight(ctx, tt.req)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			deps.Repo.AssertExpectations(t)
		})
	}
}
