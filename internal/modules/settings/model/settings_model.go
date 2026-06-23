package model

import (
	"github.com/Roisfaozi/queue-base/pkg"
)

type SettingResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	ScopeType string `json:"scope_type"`
	ScopeID   string `json:"scope_id"`
	Key       string `json:"key"`
	Value     string `json:"value"`
	ValueType string `json:"value_type"`
	IsActive  bool   `json:"is_active"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type CreateSettingRequest struct {
	ScopeType string `json:"scope_type" validate:"required,oneof=tenant branch service counter"`
	ScopeID   string `json:"scope_id" validate:"required,uuid4"`
	Key       string `json:"key" validate:"required,min=1,max=100"`
	Value     string `json:"value" validate:"required"`
	ValueType string `json:"value_type" validate:"omitempty,oneof=string number boolean json"`
}

func (r *CreateSettingRequest) Sanitize() {
	r.Key = pkg.SanitizeString(r.Key)
}

type UpdateSettingRequest struct {
	Value    *string `json:"value" validate:"omitempty"`
	IsActive *bool   `json:"is_active" validate:"omitempty"`
}

type ResolveSettingRequest struct {
	Key       string `json:"key" validate:"required"`
	BranchID  string `json:"branch_id" validate:"omitempty,uuid4"`
	ServiceID string `json:"service_id" validate:"omitempty,uuid4"`
	CounterID string `json:"counter_id" validate:"omitempty,uuid4"`
}
