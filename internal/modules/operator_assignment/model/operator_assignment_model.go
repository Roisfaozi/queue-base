package model

type OperatorAssignmentRequest struct {
	BranchID  string `json:"branch_id" validate:"required,uuid4"`
	UserID    string `json:"user_id" validate:"required,uuid4"`
	CounterID string `json:"counter_id" validate:"required,uuid4"`
}

type OperatorAssignmentResponse struct {
	ID           string `json:"id"`
	TenantID     string `json:"tenant_id,omitempty"`
	BranchID     string `json:"branch_id"`
	UserID       string `json:"user_id"`
	CounterID    string `json:"counter_id"`
	AssignedAt   int64  `json:"assigned_at"`
	UnassignedAt *int64 `json:"unassigned_at,omitempty"`
}
