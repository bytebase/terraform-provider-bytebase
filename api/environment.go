package api

// Environment is the API message for an environment.
type Environment struct {
	ID int `json:"id"`

	// Related fields
	EnvironmentTierPolicy  *EnvironmentTierPolicy  `json:"environmentTierPolicy,omitempty"`
	PipelineApprovalPolicy *PipelineApprovalPolicy `json:"pipelineApprovalPolicy,omitempty"`
	BackupPlanPolicy       *BackupPlanPolicy       `json:"backupPlanPolicy,omitempty"`

	// Domain specific fields
	Name  string `json:"name"`
	Order int    `json:"order"`
}

// EnvironmentFind is the API message for finding environment.
type EnvironmentFind struct {
	// Domain specific fields
	Name string `url:"name,omitempty"`
}

// EnvironmentUpsert is the API message for upserting an environment.
type EnvironmentUpsert struct {
	// Related fields
	EnvironmentTierPolicy  *EnvironmentTierPolicy  `json:"environmentTierPolicy,omitempty"`
	PipelineApprovalPolicy *PipelineApprovalPolicy `json:"pipelineApprovalPolicy,omitempty"`
	BackupPlanPolicy       *BackupPlanPolicy       `json:"backupPlanPolicy,omitempty"`

	// Domain specific fields
	Name  *string `json:"name,omitempty"`
	Order *int    `json:"order,omitempty"`
}

func (e *EnvironmentUpsert) HasChange() bool {
	return e.Name != nil || e.Order != nil || e.EnvironmentTierPolicy != nil || e.PipelineApprovalPolicy != nil || e.BackupPlanPolicy != nil
}
