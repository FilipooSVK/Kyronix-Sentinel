package predictor

type RiskLevel string

const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

type Recommendation string

const (
	ActionMonitor       Recommendation = "MONITOR"
	ActionInvestigate   Recommendation = "INVESTIGATE"
	ActionRebootAdvised Recommendation = "REBOOT_ADVISED"
	ActionAutoRecovery  Recommendation = "AUTO_RECOVERY"
)

// Prediction represents Sentinel's predictive assessment.
type Prediction struct {
	Risk RiskLevel

	Score int

	Confidence float64

	Reason string

	Reasons []string

	ETASeconds int64

	Recommendation Recommendation

	Trends []Trend

	ActiveSignals int

	PersistentSignals int

	KernelEvidence bool

	Signals []string
}
