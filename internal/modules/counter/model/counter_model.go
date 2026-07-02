package model

import (
	"strings"

	"github.com/Roisfaozi/queue-base/pkg"
)

type CounterResponse struct {
	ID              string `json:"id"`
	TenantID        string `json:"tenant_id"`
	BranchID        string `json:"branch_id"`
	BranchServiceID string `json:"branch_service_id,omitempty"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	DisplayName     string `json:"display_name,omitempty"`
	Status          string `json:"status"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

type CreateCounterRequest struct {
	BranchID        string `json:"branch_id" validate:"required,uuid4"`
	BranchServiceID string `json:"branch_service_id,omitempty" validate:"omitempty,uuid4"`
	Code            string `json:"code" validate:"required,min=2,max=50,xss"`
	Name            string `json:"name" validate:"required,min=3,max=255,xss"`
	DisplayName     string `json:"display_name,omitempty"`
}

func (r *CreateCounterRequest) Sanitize() {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.Name = pkg.SanitizeString(r.Name)
}

type UpdateCounterRequest struct {
	BranchServiceID *string `json:"branch_service_id,omitempty" validate:"omitempty,uuid4"`
	Code            *string `json:"code" validate:"omitempty,min=2,max=50,xss"`
	Name            *string `json:"name" validate:"omitempty,min=3,max=255,xss"`
	DisplayName     *string `json:"display_name,omitempty"`
	Status          *string `json:"status" validate:"omitempty,oneof=active inactive"`
}

func (r *UpdateCounterRequest) Sanitize() {
	if r.Code != nil {
		v := strings.ToUpper(strings.TrimSpace(*r.Code))
		r.Code = &v
	}
	if r.Name != nil {
		v := pkg.SanitizeString(*r.Name)
		r.Name = &v
	}
}
