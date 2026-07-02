package model

type SignageMeResponse struct {
	ClientID    string `json:"client_id"`
	TenantID    string `json:"tenant_id"`
	BranchID    string `json:"branch_id"`
	ClientType  string `json:"client_type"`
	Name        string `json:"name"`
	RunningText string `json:"running_text,omitempty"`
	LogoAssetID string `json:"logo_asset_id,omitempty"`
}

type SignageCurrentCallResponse struct {
	QueueID            string `json:"queue_id"`
	TicketNo           string `json:"ticket_no"`
	CounterID          string `json:"counter_id"`
	CounterDisplayName string `json:"counter_display_name,omitempty"`
	ServiceID          string `json:"service_id"`
	ServiceType        string `json:"service_type,omitempty"`
	AudioID            string `json:"audio_id,omitempty"`
	AudioEN            string `json:"audio_en,omitempty"`
}
