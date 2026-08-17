package analyzer

import (
	"time"

	"kyronix/sentinel/internal/domain"
)

// Analyzer executes health evaluation rules
// and calculates final system health state.
type Analyzer struct {
	rules          []Rule
	riskCalculator RiskCalculator
}

// NewAnalyzer creates analyzer with custom rules.
func NewAnalyzer(
	rules ...Rule,
) *Analyzer {

	return &Analyzer{
		rules:          rules,
		riskCalculator: NewRiskCalculator(),
	}
}

// Analyze evaluates a snapshot and returns health result.
func (a *Analyzer) Analyze(
	snapshot domain.Snapshot,
) domain.HealthResult {

	findings := make([]domain.Finding, 0)

	for _, rule := range a.rules {

		results := rule.Evaluate(snapshot)

		findings = append(
			findings,
			results...,
		)
	}

	result := BuildResult(findings)

	riskScore := a.riskCalculator.Calculate(
		snapshot,
		findings,
	)

	result.FreezeRiskScore = riskScore
	result.FreezeRisk = RiskFromScore(riskScore)

	result.Timestamp = time.Now()

	return result
}
