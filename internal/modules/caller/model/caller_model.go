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
