package model

type EffectiveQueueConfigRequest struct {
	BranchID  string `form:"branch_id" json:"branch_id" validate:"omitempty,uuid4"`
	ServiceID string `form:"service_id" json:"service_id" validate:"omitempty,uuid4"`
	CounterID string `form:"counter_id" json:"counter_id" validate:"omitempty,uuid4"`
}

type ResolvedQueueSetting struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Source      string `json:"source,omitempty"`
	Inherited   bool   `json:"inherited,omitempty"`
	CanOverride bool   `json:"can_override"`
	CanReset    bool   `json:"can_reset"`
}

type EffectiveConfigTenant struct {
	TenantID string `json:"tenant_id"`
}

type EffectiveConfigBranch struct {
	BranchID             string `json:"branch_id,omitempty"`
	EffectiveLogoAssetID string `json:"effective_logo_asset_id,omitempty"`
}

type EffectiveConfigQueue struct {
	QueueResetTime           ResolvedQueueSetting `json:"queue_reset_time"`
	TicketPrefix             ResolvedQueueSetting `json:"ticket_prefix"`
	NumberingStrategy        ResolvedQueueSetting `json:"numbering_strategy"`
	DefaultEstimatedDuration ResolvedQueueSetting `json:"default_estimated_duration"`
	AllowForward             *bool                `json:"allow_forward,omitempty"`
	AllowSkip                *bool                `json:"allow_skip,omitempty"`
	AllowRecall              *bool                `json:"allow_recall,omitempty"`
	AllowCancel              *bool                `json:"allow_cancel,omitempty"`
	AutoCallNext             *bool                `json:"auto_call_next,omitempty"`
	MaxServiceDuration       *int                 `json:"max_service_duration,omitempty"`
	MinServiceDuration       *int                 `json:"min_service_duration,omitempty"`
	RequireCounter           *bool                `json:"require_counter,omitempty"`
	AllowForwardFrom         *bool                `json:"allow_forward_from,omitempty"`
	AllowForwardTo           *bool                `json:"allow_forward_to,omitempty"`
	AudioID                  *string              `json:"audio_id,omitempty"`
	AudioEN                  *string              `json:"audio_en,omitempty"`
	NarrativeInstructionID   *string              `json:"narrative_instruction_id,omitempty"`
	NarrativeInstructionEN   *string              `json:"narrative_instruction_en,omitempty"`
}

type EffectiveQueueConfigResponse struct {
	TenantID                          string                `json:"tenant_id"`
	BranchID                          string                `json:"branch_id,omitempty"`
	ServiceID                         string                `json:"service_id,omitempty"`
	CounterID                         string                `json:"counter_id,omitempty"`
	Tenant                            EffectiveConfigTenant `json:"tenant"`
	Branch                            EffectiveConfigBranch `json:"branch"`
	Queue                             EffectiveConfigQueue  `json:"queue"`
	QueueResetTime                    string                `json:"queue_reset_time"`
	QueueResetTimeSource              string                `json:"queue_reset_time_source,omitempty"`
	QueueResetTimeInherited           bool                  `json:"queue_reset_time_inherited,omitempty"`
	TicketPrefix                      string                `json:"ticket_prefix"`
	TicketPrefixSource                string                `json:"ticket_prefix_source,omitempty"`
	TicketPrefixInherited             bool                  `json:"ticket_prefix_inherited,omitempty"`
	NumberingStrategy                 string                `json:"numbering_strategy"`
	NumberingStrategySource           string                `json:"numbering_strategy_source,omitempty"`
	NumberingStrategyInherited        bool                  `json:"numbering_strategy_inherited,omitempty"`
	DefaultEstimatedDuration          string                `json:"default_estimated_duration,omitempty"`
	DefaultEstimatedDurationSource    string                `json:"default_estimated_duration_source,omitempty"`
	DefaultEstimatedDurationInherited bool                  `json:"default_estimated_duration_inherited,omitempty"`
	AllowForward                      *bool                 `json:"allow_forward,omitempty"`
	AllowSkip                         *bool                 `json:"allow_skip,omitempty"`
	AllowRecall                       *bool                 `json:"allow_recall,omitempty"`
	AllowCancel                       *bool                 `json:"allow_cancel,omitempty"`
	AutoCallNext                      *bool                 `json:"auto_call_next,omitempty"`
	MaxServiceDuration                *int                  `json:"max_service_duration,omitempty"`
	MinServiceDuration                *int                  `json:"min_service_duration,omitempty"`
	RequireCounter                    *bool                 `json:"require_counter,omitempty"`
	AllowForwardFrom                  *bool                 `json:"allow_forward_from,omitempty"`
	AllowForwardTo                    *bool                 `json:"allow_forward_to,omitempty"`
	AudioID                           *string               `json:"audio_id,omitempty"`
	AudioEN                           *string               `json:"audio_en,omitempty"`
	NarrativeInstructionID            *string               `json:"narrative_instruction_id,omitempty"`
	NarrativeInstructionEN            *string               `json:"narrative_instruction_en,omitempty"`
	EffectiveUntil                    *string               `json:"effective_until,omitempty"`
}
