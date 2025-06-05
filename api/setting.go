package api

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
