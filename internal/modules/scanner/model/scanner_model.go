package model

type CheckInRequest struct {
	Action               string `json:"action" validate:"required,oneof=register forward"`
	BranchID             string `json:"branch_id" validate:"required,uuid4"`
	ServiceID            string `json:"service_id"`
	PatientID            string `json:"patient_id"`
	PatientName          string `json:"patient_name"`
	QueueID              string `json:"queue_id"`
	DestinationServiceID string `json:"destination_service_id"`
	DestinationCounterID string `json:"destination_counter_id"`
}
