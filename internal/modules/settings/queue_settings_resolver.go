package settings

import (
	"context"

	settingsModel "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/model"
	settingsUsecase "github.com/Roisfaozi/go-clean-boilerplate/internal/modules/settings/usecase"
)

type QueueSettingsResolver struct {
	useCase settingsUsecase.SettingsUseCase
}

func NewQueueSettingsResolver(useCase settingsUsecase.SettingsUseCase) *QueueSettingsResolver {
	return &QueueSettingsResolver{useCase: useCase}
}

func (r *QueueSettingsResolver) Resolve(ctx context.Context, key string, branchID string, serviceID string, counterID string) (string, error) {
	res, err := r.useCase.ResolveSetting(ctx, &settingsModel.ResolveSettingRequest{
		Key:       key,
		BranchID:  branchID,
		ServiceID: serviceID,
		CounterID: counterID,
	})
	if err != nil {
		return "", err
	}
	return res.Value, nil
}
