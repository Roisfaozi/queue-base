package model

type CheckInRequest struct {
	Action               string `json:"action" validate:"required,oneof=register forward"`
	ClientID             string `json:"client_id" validate:"required"`
	APIKey               string `json:"api_key" validate:"required"`
	ServiceID            string `json:"service_id"`
	PatientID            string `json:"patient_id"`
	PatientName          string `json:"patient_name"`
	QueueID              string `json:"queue_id"`
	DestinationServiceID string `json:"destination_service_id"`
	DestinationCounterID string `json:"destination_counter_id"`
}
