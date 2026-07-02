package model

import (
	"strings"

	"github.com/Roisfaozi/queue-base/pkg"
)

type ServiceResponse struct {
	ID                       string `json:"id"`
	TenantID                 string `json:"tenant_id"`
	Code                     string `json:"code"`
	Name                     string `json:"name"`
	Type                     string `json:"type"`
	DefaultEstimatedDuration int    `json:"default_estimated_duration"`
	Status                   string `json:"status"`
	IsPharmacy               bool   `json:"is_pharmacy"`
	IsPharmacyReception      bool   `json:"is_pharmacy_reception"`
	CreatedAt                int64  `json:"created_at"`
	UpdatedAt                int64  `json:"updated_at"`
}

type CreateServiceRequest struct {
	Code                     string `json:"code" validate:"required,min=2,max=50,xss"`
	Name                     string `json:"name" validate:"required,min=3,max=255,xss"`
	Type                     string `json:"type,omitempty"`
	DefaultEstimatedDuration int    `json:"default_estimated_duration,omitempty"`
	IsPharmacy               bool   `json:"is_pharmacy"`
	IsPharmacyReception      bool   `json:"is_pharmacy_reception"`
}

func (r *CreateServiceRequest) Sanitize() {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.Name = pkg.SanitizeString(r.Name)
}

type UpdateServiceRequest struct {
	Code                     *string `json:"code" validate:"omitempty,min=2,max=50,xss"`
	Name                     *string `json:"name" validate:"omitempty,min=3,max=255,xss"`
	Type                     *string `json:"type,omitempty"`
	DefaultEstimatedDuration *int    `json:"default_estimated_duration,omitempty"`
	Status                   *string `json:"status" validate:"omitempty,oneof=active inactive"`
	IsPharmacy               *bool   `json:"is_pharmacy"`
	IsPharmacyReception      *bool   `json:"is_pharmacy_reception"`
}

func (r *UpdateServiceRequest) Sanitize() {
	if r.Code != nil {
		v := strings.ToUpper(strings.TrimSpace(*r.Code))
		r.Code = &v
	}
	if r.Name != nil {
		v := pkg.SanitizeString(*r.Name)
		r.Name = &v
	}
}
