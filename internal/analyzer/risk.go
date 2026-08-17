package analyzer

import (
	"kyronix/sentinel/internal/domain"
)

// RiskCalculator calculates freeze probability.
type RiskCalculator struct{}

// NewRiskCalculator creates freeze risk calculator.
func NewRiskCalculator() RiskCalculator {
	return RiskCalculator{}
}

// Calculate evaluates freeze risk based on snapshot and findings.
//
// Result:
// 0   = no risk
// 100 = imminent freeze risk
func (RiskCalculator) Calculate(
	snapshot domain.Snapshot,
	findings []domain.Finding,
) int {

	score := 0

	// Kernel OOM events are the strongest indicator.
	if snapshot.Kernel.OOM.SystemKillCount != nil &&
		*snapshot.Kernel.OOM.SystemKillCount > 0 {
		score += 50
	}

	if snapshot.Kernel.OOM.CgroupKillCount != nil &&
		*snapshot.Kernel.OOM.CgroupKillCount > 0 {
		score += 40
	}

	// IO pressure is a strong freeze indicator.
	if snapshot.Pressure.IO.Full.Avg10 >= 20 {
		score += 30
	} else if snapshot.Pressure.IO.Full.Avg10 >= 10 {
		score += 15
	}

	// Memory pressure indicates reclaim stalls.
	if snapshot.Pressure.Memory.Full.Avg10 >= 20 {
		score += 30
	} else if snapshot.Pressure.Memory.Some.Avg10 >= 10 {
		score += 15
	}

	// Swap activity increases freeze probability.
	if snapshot.Memory.SwapPercent >= 50 {
		score += 20
	} else if snapshot.Memory.SwapPercent >= 10 {
		score += 10
	}

	// Severe findings increase risk.
	for _, finding := range findings {

		if finding.Severity == domain.SeverityCritical {
			score += 10
		}
	}

	if score > 100 {
		score = 100
	}

	return score
}

// RiskFromScore converts numeric risk into category.
func RiskFromScore(
	score int,
) domain.FreezeRisk {

	switch {
	case score >= 80:
		return domain.RiskCritical

	case score >= 60:
		return domain.RiskHigh

	case score >= 30:
		return domain.RiskMedium

	default:
		return domain.RiskLow
	}
}
