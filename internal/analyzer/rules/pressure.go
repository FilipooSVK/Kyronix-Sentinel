package rules

import (
	"kyronix/sentinel/internal/domain"
)

// PressureRule evaluates Linux PSI pressure metrics.
type PressureRule struct{}

// NewPressureRule creates a pressure health rule.
func NewPressureRule() PressureRule {
	return PressureRule{}
}

// Evaluate analyzes CPU, memory and IO pressure.
func (PressureRule) Evaluate(
	snapshot domain.Snapshot,
) []domain.Finding {

	findings := make([]domain.Finding, 0)

	pressure := snapshot.Pressure

	// IO pressure is highly correlated with system stalls.
	switch {
	case pressure.IO.Full.Avg10 >= 20:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "pressure",
			Message:  "Critical IO pressure detected",
			Penalty:  25,
		})

	case pressure.IO.Full.Avg10 >= 10:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityWarning,
			Source:   "pressure",
			Message:  "High IO pressure detected",
			Penalty:  10,
		})
	}

	// Memory pressure indicates contention and possible instability.
	switch {
	case pressure.Memory.Full.Avg10 >= 20:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "pressure",
			Message:  "Critical memory pressure detected",
			Penalty:  25,
		})

	case pressure.Memory.Some.Avg10 >= 10:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityWarning,
			Source:   "pressure",
			Message:  "High memory pressure detected",
			Penalty:  10,
		})
	}

	// CPU pressure indicates scheduler contention.
	switch {
	case pressure.CPU.Some.Avg10 >= 20:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityWarning,
			Source:   "pressure",
			Message:  "High CPU pressure detected",
			Penalty:  10,
		})
	}

	return findings
}
