package model

type QMSClientRequest struct {
	BranchID   string `json:"branch_id" validate:"required,uuid4"`
	ClientType string `json:"client_type" validate:"required,oneof=caller signage scanner kiosk"`
	Name       string `json:"name" validate:"required,min=1,max=255"`
}

type QMSClientUpdateRequest struct {
	Name            string  `json:"name" validate:"omitempty,min=1,max=255"`
	BranchServiceID *string `json:"branch_service_id" validate:"omitempty,uuid4"`
	CounterID       *string `json:"counter_id" validate:"omitempty,uuid4"`
	IsActive        *bool   `json:"is_active"`
}

type QMSClientResponse struct {
	ID              string  `json:"id"`
	TenantID        string  `json:"tenant_id,omitempty"`
	BranchID        string  `json:"branch_id"`
	ClientType      string  `json:"client_type"`
	Name            string  `json:"name"`
	BranchServiceID *string `json:"branch_service_id,omitempty"`
	CounterID       *string `json:"counter_id,omitempty"`
	IsActive        bool    `json:"is_active"`
	CreatedAt       int64   `json:"created_at"`
}

type QMSClientCredentialRequest struct {
	ClientID string `json:"client_id" validate:"required,uuid4"`
	APIKey   string `json:"api_key" validate:"required,min=8"`
}

type QMSClientCredentialResponse struct {
	ID        string `json:"id"`
	ClientID  string `json:"client_id"`
	ExpiresAt *int64 `json:"expires_at,omitempty"`
	CreatedAt int64  `json:"created_at"`
}
