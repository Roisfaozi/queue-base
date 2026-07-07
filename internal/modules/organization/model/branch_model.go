package model

import (
	"strings"

	"github.com/Roisfaozi/queue-base/pkg"
)

type BranchResponse struct {
	ID          string `json:"id"`
	TenantID    string `json:"tenant_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Province    string `json:"province,omitempty"`
	PostalCode  string `json:"postal_code,omitempty"`
	Phone       string `json:"phone,omitempty"`
	Email       string `json:"email,omitempty"`
	LogoAssetID string `json:"logo_asset_id,omitempty"`
	RunningText string `json:"running_text,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

type ResolveBranchContextRequest struct {
	BranchID string `json:"branch_id" validate:"required,uuid4"`
}

type CreateBranchRequest struct {
	Code        string `json:"code" validate:"required,min=2,max=50,xss"`
	Name        string `json:"name" validate:"required,min=3,max=255,xss"`
	Address     string `json:"address,omitempty"`
	City        string `json:"city,omitempty"`
	Province    string `json:"province,omitempty"`
	Phone       string `json:"phone,omitempty"`
	RunningText string `json:"running_text,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

func (r *CreateBranchRequest) Sanitize() {
	r.Code = strings.ToUpper(strings.TrimSpace(r.Code))
	r.Name = pkg.SanitizeString(r.Name)
}

type UpdateBranchRequest struct {
	Code        *string `json:"code" validate:"omitempty,min=2,max=50,xss"`
	Name        *string `json:"name" validate:"omitempty,min=3,max=255,xss"`
	Address     *string `json:"address,omitempty"`
	City        *string `json:"city,omitempty"`
	Province    *string `json:"province,omitempty"`
	PostalCode  *string `json:"postal_code,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	Email       *string `json:"email,omitempty"`
	LogoAssetID *string `json:"logo_asset_id,omitempty"`
	RunningText *string `json:"running_text,omitempty"`
	Timezone    *string `json:"timezone,omitempty"`
	Status      *string `json:"status" validate:"omitempty,oneof=active inactive"`
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
