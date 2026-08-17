package analyzer

import (
	"time"

	"kyronix/sentinel/internal/domain"
)

// BuildResult creates a HealthResult from findings.
func BuildResult(
	findings []domain.Finding,
) domain.HealthResult {

	score := 100

	for _, finding := range findings {
		score -= finding.Penalty
	}

	if score < 0 {
		score = 0
	}

	return domain.HealthResult{
		Timestamp:       time.Now(),
		HealthScore:     score,
		FreezeRiskScore: 100 - score,
		FreezeRisk:      calculateRisk(score),
		Quality:         domain.AssessmentComplete,
		Findings:        findings,
	}
}

func calculateRisk(score int) domain.FreezeRisk {

	switch {
	case score >= 80:
		return domain.RiskLow

	case score >= 60:
		return domain.RiskMedium

	case score >= 30:
		return domain.RiskHigh

	default:
		return domain.RiskCritical
	}
}
