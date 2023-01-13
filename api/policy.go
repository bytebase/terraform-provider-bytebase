package api

// PolicyType is the type for the policy.
type PolicyType string

// ApprovalStrategy is strategy for deployment approval policy.
type ApprovalStrategy string

// ApprovalGroup is the group for deployment approval policy.
type ApprovalGroup string

// BackupPlanSchedule is schedule for backup plan policy.
type BackupPlanSchedule string

// SQLReviewRuleLevel is the error level for SQL review rule.
type SQLReviewRuleLevel string

// SensitiveDataMaskType is the mask type for sensitive data.
type SensitiveDataMaskType string

const (
	// PolicyTypeDeploymentApproval is the policy type for deployment approval policy.
	PolicyTypeDeploymentApproval PolicyType = "DEPLOYMENT_APPROVAL"
	// PolicyTypeBackupPlan is the policy type for backup plan policy.
	PolicyTypeBackupPlan PolicyType = "BACKUP_PLAN"
	// PolicyTypeSQLReview is the policy type for SQL review policy.
	PolicyTypeSQLReview PolicyType = "SQL_REVIEW"
	// PolicyTypeSensitiveData is the policy type for sensitive data policy.
	PolicyTypeSensitiveData PolicyType = "SENSITIVE_DATA"
	// PolicyTypeAccessControl is the policy type for access control policy.
	PolicyTypeAccessControl PolicyType = "ACCESS_CONTROL"

	// ApprovalStrategyAutomatic means the pipeline will automatically be approved without user intervention.
	ApprovalStrategyAutomatic ApprovalStrategy = "AUTOMATIC"
	// ApprovalStrategyManual means the pipeline should be manually approved by user to proceed.
	ApprovalStrategyManual ApprovalStrategy = "MANUAL"

	// ApprovalGroupDBA means the assignee can be selected from the workspace owners and DBAs.
	ApprovalGroupDBA ApprovalGroup = "APPROVAL_GROUP_DBA"
	// ApprovalGroupOwner means the assignee can be selected from the project owners.
	ApprovalGroupOwner ApprovalGroup = "APPROVAL_GROUP_PROJECT_OWNER"

	// BackupPlanScheduleUnset is NEVER backup plan policy value.
	BackupPlanScheduleUnset BackupPlanSchedule = "UNSET"
	// BackupPlanScheduleDaily is DAILY backup plan policy value.
	BackupPlanScheduleDaily BackupPlanSchedule = "DAILY"
	// BackupPlanScheduleWeekly is WEEKLY backup plan policy value.
	BackupPlanScheduleWeekly BackupPlanSchedule = "WEEKLY"

	// SensitiveDataMaskTypeDefault is the sensitive data type to hide data with a default method.
	// The default method is subject to change.
	SensitiveDataMaskTypeDefault SensitiveDataMaskType = "DEFAULT"

	// SQLReviewRuleLevelError is the error level for SQL review rule.
	SQLReviewRuleLevelError SQLReviewRuleLevel = "ERROR"
	// SQLReviewRuleLevelWarning is the warning level for SQL review rule.
	SQLReviewRuleLevelWarning SQLReviewRuleLevel = "WARNING"
	// SQLReviewRuleLevelDisabled is the disabled level for SQL review rule.
	SQLReviewRuleLevelDisabled SQLReviewRuleLevel = "DISABLED"
)

// DeploymentApprovalPolicy is the policy configuration for deployment approval.
type DeploymentApprovalPolicy struct {
	DefaultStrategy              ApprovalStrategy              `json:"defaultStrategy"`
	DeploymentApprovalStrategies []*DeploymentApprovalStrategy `json:"deploymentApprovalStrategies"`
}

// DeploymentApprovalStrategy is the API message for deployment approval strategy.
type DeploymentApprovalStrategy struct {
	ApprovalGroup    ApprovalGroup    `json:"approvalGroup"`
	ApprovalStrategy ApprovalStrategy `json:"approvalStrategy"`
	DeploymentType   DeploymentType   `json:"deploymentType"`
}

// BackupPlanPolicy is the policy configuration for backup plan.
type BackupPlanPolicy struct {
	Schedule BackupPlanSchedule `json:"schedule"`
	// RetentionDuration is the minimum allowed period that backup data is kept for databases in an environment.
	RetentionDuration string `json:"retentionDuration"`
}

// SensitiveDataPolicy is the API message for sensitive data policy.
type SensitiveDataPolicy struct {
	SensitiveData []*SensitiveData `json:"sensitiveData"`
}

// SensitiveData is the API message for sensitive data.
type SensitiveData struct {
	Schema   string                `json:"schema"`
	Table    string                `json:"table"`
	Column   string                `json:"column"`
	MaskType SensitiveDataMaskType `json:"maskType"`
}

// AccessControlPolicy is the API message for access control policy.
type AccessControlPolicy struct {
	DisallowRules []*AccessControlRule `json:"disallowRules"`
}

// AccessControlRule is the API message for access control rule.
type AccessControlRule struct {
	FullDatabase bool `json:"fullDatabase"`
}

// SQLReviewPolicy is the API message for SQL review policy.
type SQLReviewPolicy struct {
	Title string           `type:"title"`
	Rules []*SQLReviewRule `type:"rules"`
}

// SQLReviewRule is the API message for SQL review rule.
type SQLReviewRule struct {
	Type    string             `json:"type"`
	Level   SQLReviewRuleLevel `json:"level"`
	Payload string             `json:"payload"`
}

// PolicyFindMessage is the API message for finding policies.
type PolicyFindMessage struct {
	ProjectID     *string
	EnvironmentID *string
	InstanceID    *string
	DatabaseName  *string
	Type          *PolicyType
	ShowDeleted   bool
}

// PolicyMessage is the API message for policy.
type PolicyMessage struct {
	UID               string     `json:"uid"`
	Name              string     `json:"name"`
	State             State      `json:"state,omitempty"`
	InheritFromParent bool       `json:"inheritFromParent"`
	Type              PolicyType `json:"type"`

	// The policy payload
	DeploymentApprovalPolicy *DeploymentApprovalPolicy `json:"deploymentApprovalPolicy"`
	BackupPlanPolicy         *BackupPlanPolicy         `json:"backupPlanPolicy"`
	SensitiveDataPolicy      *SensitiveDataPolicy      `json:"sensitiveDataPolicy"`
	AccessControlPolicy      *AccessControlPolicy      `json:"accessControlPolicy"`
	SQLReviewPolicy          *SQLReviewPolicy          `json:"sqlReviewPolicy"`
}

// PolicyPatchMessage is the API message to patch the policy.
type PolicyPatchMessage struct {
	InheritFromParent *bool      `json:"inheritFromParent"`
	Type              PolicyType `json:"type"`

	// The policy payload
	DeploymentApprovalPolicy *DeploymentApprovalPolicy `json:"deploymentApprovalPolicy"`
	BackupPlanPolicy         *BackupPlanPolicy         `json:"backupPlanPolicy"`
	SensitiveDataPolicy      *SensitiveDataPolicy      `json:"sensitiveDataPolicy"`
	AccessControlPolicy      *AccessControlPolicy      `json:"accessControlPolicy"`
	SQLReviewPolicy          *SQLReviewPolicy          `json:"sqlReviewPolicy"`
}

// ListPolicyMessage is the API message for list policy response.
type ListPolicyMessage struct {
	Policies      []*PolicyMessage `json:"policies"`
	NextPageToken string           `json:"nextPageToken"`
}
