package model

type CallerActionRequest struct {
	Action string `json:"action" validate:"required,oneof=call serve complete skip cancel"`
}

type CallerActionResponse struct {
	Success   bool   `json:"success"`
	TrackNo   string `json:"track_no,omitempty"`
	QueueNo   int    `json:"queue_no,omitempty"`
	Status    string `json:"status"`
	JourneyID string `json:"journey_id"`
}

type CallerLoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50,xss"`
	Password string `json:"password" validate:"required,min=8,max=72"`
}

type CallerLoginResponse struct {
	AccessToken string `json:"access_token"`
	Context     struct {
		TenantID           string `json:"tenant_id"`
		TenantName         string `json:"tenant_name,omitempty"`
		BranchID           string `json:"branch_id"`
		BranchName         string `json:"branch_name,omitempty"`
		BranchServiceID    string `json:"branch_service_id,omitempty"`
		ServiceName        string `json:"service_name,omitempty"`
		CounterID          string `json:"counter_id,omitempty"`
		CounterName        string `json:"counter_name,omitempty"`
		CounterDisplayName string `json:"display_name,omitempty"`
	} `json:"context"`
	Permissions []string `json:"permissions"`
}

type CallerMeResponse struct {
	CallerLoginResponse
}
