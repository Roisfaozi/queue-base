package mocks

import (
	"context"
	mock "github.com/stretchr/testify/mock"
)

type MockQueueConfigRepository struct {
	mock.Mock
}

func (_m *MockQueueConfigRepository) UpsertTenantQueueSetting(ctx context.Context, tenantID string, values map[string]any) error {
	ret := _m.Called(ctx, tenantID, values)
	return ret.Error(0)
}

func (_m *MockQueueConfigRepository) UpsertBranchQueueSetting(ctx context.Context, tenantID string, branchID string, values map[string]any) error {
	ret := _m.Called(ctx, tenantID, branchID, values)
	return ret.Error(0)
}

func (_m *MockQueueConfigRepository) UpsertBranchServiceQueueSetting(ctx context.Context, tenantID string, branchID string, branchServiceID string, values map[string]any) error {
	ret := _m.Called(ctx, tenantID, branchID, branchServiceID, values)
	return ret.Error(0)
}

func (_m *MockQueueConfigRepository) UpsertCounterQueueSetting(ctx context.Context, tenantID string, counterID string, values map[string]any) error {
	ret := _m.Called(ctx, tenantID, counterID, values)
	return ret.Error(0)
}
