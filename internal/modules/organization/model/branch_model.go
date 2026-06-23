package model

import (
	"strings"

	"github.com/Roisfaozi/queue-base/pkg"
)

type BranchResponse struct {
	ID        string `json:"id"`
	TenantID  string `json:"tenant_id"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

type ResolveBranchContextRequest struct {
	BranchID string `json:"branch_id" validate:"required,uuid4"`
}

type CreateBranchRequest struct {
	Code string `json:"code" validate:"required,min=2,max=50,xss"`
	Name string `json:"name" validate:"required,min=3,max=255,xss"`
}

func (r *CreateBranchRequest) Sanitize() {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.Name = pkg.SanitizeString(r.Name)
}

type UpdateBranchRequest struct {
	Code   *string `json:"code" validate:"omitempty,min=2,max=50,xss"`
	Name   *string `json:"name" validate:"omitempty,min=3,max=255,xss"`
	Status *string `json:"status" validate:"omitempty,oneof=active inactive"`
}

func (r *UpdateBranchRequest) Sanitize() {
	if r.Code != nil {
		v := strings.ToUpper(strings.TrimSpace(*r.Code))
		r.Code = &v
	}
	if r.Name != nil {
		v := pkg.SanitizeString(*r.Name)
		r.Name = &v
	}
}
