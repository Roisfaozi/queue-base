package model

type BranchServiceResponse struct {
	ID         string `json:"id"`
	TenantID   string `json:"tenant_id"`
	BranchID   string `json:"branch_id"`
	ServiceID  string `json:"service_id"`
	CustomName string `json:"custom_name,omitempty"`
	IsActive   bool   `json:"is_active"`
	SortOrder  int    `json:"sort_order"`
	CreatedAt  int64  `json:"created_at"`
	UpdatedAt  int64  `json:"updated_at"`
}

type CreateBranchServiceRequest struct {
	ServiceID  string `json:"service_id" validate:"required,uuid4"`
	CustomName string `json:"custom_name,omitempty" validate:"omitempty,max=255,xss"`
	IsActive   *bool  `json:"is_active,omitempty"`
	SortOrder  int    `json:"sort_order,omitempty"`
}

type UpdateBranchServiceRequest struct {
	CustomName *string `json:"custom_name,omitempty" validate:"omitempty,max=255,xss"`
	IsActive   *bool   `json:"is_active,omitempty"`
	SortOrder  *int    `json:"sort_order,omitempty"`
}
