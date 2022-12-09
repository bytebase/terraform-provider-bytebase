package api

// PipelineApprovalType is the type for PipelineApprovalPolicy, used by Terraform provider to simplify the policy configuration.
type PipelineApprovalType string

// PipelineApprovalValue is value for approval policy.
type PipelineApprovalValue string

// IssueType is the type of an issue.
type IssueType string

// AssigneeGroupValue is the value for assignee group policy.
type AssigneeGroupValue string

const (
	// PipelineApprovalTypeNever means the pipeline will automatically be approved without user intervention.
	PipelineApprovalTypeNever PipelineApprovalType = "MANUAL_APPROVAL_NEVER"
	// PipelineApprovalTypeByProjectOwner means the pipeline should be manually approved by project owner to proceed.
	PipelineApprovalTypeByProjectOwner PipelineApprovalType = "MANUAL_APPROVAL_BY_PROJECT_OWNER"
	// PipelineApprovalTypeByWorkspaceOwnerORDBA means the pipeline should be manually approved by workspace owner or DBA to proceed.
	PipelineApprovalTypeByWorkspaceOwnerORDBA PipelineApprovalType = "MANUAL_APPROVAL_BY_WORKSPACE_OWNER_OR_DBA"

	// PipelineApprovalValueManualNever means the pipeline will automatically be approved without user intervention.
	PipelineApprovalValueManualNever PipelineApprovalValue = "MANUAL_APPROVAL_NEVER"
	// PipelineApprovalValueManualAlways means the pipeline should be manually approved by user to proceed.
	PipelineApprovalValueManualAlways PipelineApprovalValue = "MANUAL_APPROVAL_ALWAYS"

	// AssigneeGroupValueWorkspaceOwnerOrDBA means the assignee can be selected from the workspace owners and DBAs.
	AssigneeGroupValueWorkspaceOwnerOrDBA AssigneeGroupValue = "WORKSPACE_OWNER_OR_DBA"
	// AssigneeGroupValueProjectOwner means the assignee can be selected from the project owners.
	AssigneeGroupValueProjectOwner AssigneeGroupValue = "PROJECT_OWNER"

	// IssueDatabaseSchemaUpdate is the issue type for updating database schemas (DDL).
	IssueDatabaseSchemaUpdate IssueType = "bb.issue.database.schema.update"
	// IssueDatabaseSchemaUpdateGhost is the issue type for updating database schemas using gh-ost.
	IssueDatabaseSchemaUpdateGhost IssueType = "bb.issue.database.schema.update.ghost"
	// IssueDatabaseDataUpdate is the issue type for updating database data (DML).
	IssueDatabaseDataUpdate IssueType = "bb.issue.database.data.update"
)

// PipelineApprovalPolicy is the policy configuration for pipeline approval.
type PipelineApprovalPolicy struct {
	Value PipelineApprovalValue `json:"value"`
	// The AssigneeGroup is the final value of the assignee group which overrides the default value.
	// If there is no value provided in the AssigneeGroupList, we use the the workspace owners and DBAs (default) as the available assignee.
	// If the AssigneeGroupValue is PROJECT_OWNER, the available assignee is the project owners.
	AssigneeGroupList []AssigneeGroup `json:"assigneeGroupList"`
}

// AssigneeGroup is the configuration of the assignee group.
type AssigneeGroup struct {
	IssueType IssueType          `json:"issueType"`
	Value     AssigneeGroupValue `json:"value"`
}

// BackupPlanPolicy is the policy configuration for backup plan.
type BackupPlanPolicy struct {
	Schedule string `json:"schedule"`
	// RetentionPeriodTs is the minimum allowed period that backup data is kept for databases in an environment.
	RetentionPeriodTs int `json:"retentionPeriodTs"`
}

// EnvironmentTierPolicy is the tier of an environment.
type EnvironmentTierPolicy struct {
	EnvironmentTier string `json:"environmentTier"`
}
