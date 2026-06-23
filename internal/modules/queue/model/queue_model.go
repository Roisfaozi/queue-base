package model

import (
	"github.com/Roisfaozi/go-clean-boilerplate/pkg"
)

type QueueResponse struct {
	ID               string `json:"id"`
	TenantID         string `json:"tenant_id"`
	BranchID         string `json:"branch_id"`
	QueueDate        string `json:"queue_date"`
	TicketNo         string `json:"ticket_no"`
	QueueNo          int    `json:"queue_no"`
	PatientID        string `json:"patient_id,omitempty"`
	PatientName      string `json:"patient_name,omitempty"`
	Status           string `json:"status"`
	CurrentJourneyID string `json:"current_journey_id,omitempty"`
	CreatedAt        int64  `json:"created_at"`
	UpdatedAt        int64  `json:"updated_at"`
}

type RegisterQueueRequest struct {
	ServiceID   string `json:"service_id" validate:"required,uuid4"`
	PatientID   string `json:"patient_id" validate:"omitempty,uuid4"`
	PatientName string `json:"patient_name" validate:"required,min=2,max=255,xss"`
}

func (r *RegisterQueueRequest) Sanitize() {
	r.PatientName = pkg.SanitizeString(r.PatientName)
}

type ForwardQueueRequest struct {
	DestinationServiceID string `json:"destination_service_id" validate:"required,uuid4"`
	DestinationCounterID string `json:"destination_counter_id" validate:"omitempty,uuid4"`
}

const (
	QueueActionCall     = "call"
	QueueActionServe    = "serve"
	QueueActionComplete = "complete"
	QueueActionSkip     = "skip"
	QueueActionCancel   = "cancel"
)

type QueueTransitionRequest struct {
	Action string `json:"action" validate:"required,oneof=call serve complete skip cancel"`
}
