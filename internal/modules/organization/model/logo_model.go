package model

// LogoScope identifies which tenant-owned entity a logo belongs to.
type LogoScope string

const (
	LogoScopeOrganization LogoScope = "organization"
	LogoScopeBranch       LogoScope = "branch"
)
