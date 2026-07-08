package mocks

import (
	"context"
	auditModel "github.com/Roisfaozi/queue-base/internal/modules/audit/model"
	"github.com/Roisfaozi/queue-base/pkg/querybuilder"
	mock "github.com/stretchr/testify/mock"
)

type MockAuditUseCase struct {
	mock.Mock
}

func (_m *MockAuditUseCase) LogActivity(ctx context.Context, req auditModel.CreateAuditLogRequest) error {
	ret := _m.Called(ctx, req)
	return ret.Error(0)
}

func (_m *MockAuditUseCase) GetLogsDynamic(ctx context.Context, filter *querybuilder.DynamicFilter) ([]auditModel.AuditLogResponse, int64, error) {
	return nil, 0, nil
}

func (_m *MockAuditUseCase) ExportLogs(ctx context.Context, fromDate, toDate string, process func([]auditModel.AuditLogResponse) error) error {
	return nil
}

func (_m *MockAuditUseCase) ExportLogsAsync(ctx context.Context, userID, orgID, fromDate, toDate, format string) error {
	return nil
}

