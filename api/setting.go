package api

// SettingName is the Bytebase setting name without settings/ prefix.
type SettingName string

const (
	// SettingWorkspaceApproval is the setting name for workspace approval config.
	SettingWorkspaceApproval SettingName = "bb.workspace.approval"
	// SettingWorkspaceExternalApproval is the setting name for workspace external approval config.
	SettingWorkspaceExternalApproval SettingName = "bb.workspace.approval.external"
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
