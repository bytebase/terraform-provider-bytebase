package api

// SettingName is the Bytebase setting name without settings/ prefix.
type SettingName string

const (
	// SettingWorkspaceApproval is the setting name for workspace approval config.
	SettingWorkspaceApproval SettingName = "bb.workspace.approval"
	// SettingWorkspaceProfile is the setting name for workspace profile settings.
	SettingWorkspaceProfile SettingName = "bb.workspace.profile"
	// SettingWorkspaceExternalApproval is the setting name for workspace external approval config.
	SettingWorkspaceExternalApproval SettingName = "bb.workspace.approval.external"
	// SettingDataClassification is the setting name for data classification.
	SettingDataClassification SettingName = "bb.workspace.data-classification"
)

// RiskLevel is the approval risk level.
type RiskLevel string

const (
	// RiskLevelDefault is the default risk level, the level number should be 0.
	RiskLevelDefault RiskLevel = "DEFAULT"
	// RiskLevelLow is the low risk level, the level number should be 100.
	RiskLevelLow RiskLevel = "LOW"
	// RiskLevelModerate is the moderate risk level, the level number should be 200.
	RiskLevelModerate RiskLevel = "MODERATE"
	// RiskLevelHigh is the high risk level, the level number should be 300.
	RiskLevelHigh RiskLevel = "HIGH"
)

// Int returns the int value for risk.
func (r RiskLevel) Int() int {
	switch r {
	case RiskLevelLow:
		return 100
	case RiskLevelModerate:
		return 200
	case RiskLevelHigh:
		return 300
	default:
		return 0
	}
}

// ApprovalNodeType is the type for approval node.
type ApprovalNodeType string

const (
	// ApprovalNodeTypeGroup means the approval node is a group.
	ApprovalNodeTypeGroup ApprovalNodeType = "GROUP"
	// ApprovalNodeTypeRole means the approval node is a role, the value should be role fullname.
	ApprovalNodeTypeRole ApprovalNodeType = "ROLE"
	// ApprovalNodeTypeExternalNodeID means the approval node is a external node, the value should be the node id.
	ApprovalNodeTypeExternalNodeID ApprovalNodeType = "EXTERNAL_NODE"
)
