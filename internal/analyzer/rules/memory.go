package rules

import (
	"kyronix/sentinel/internal/domain"
)

// MemoryRule evaluates memory utilization risks.
type MemoryRule struct{}

// NewMemoryRule creates a memory health rule.
func NewMemoryRule() MemoryRule {
	return MemoryRule{}
}

// Evaluate checks RAM and swap utilization.
func (MemoryRule) Evaluate(
	snapshot domain.Snapshot,
) []domain.Finding {

	findings := make([]domain.Finding, 0)

	memory := snapshot.Memory

	// RAM usage evaluation.
	switch {
	case memory.UsedPercent >= 95:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "memory",
			Message:  "Critical memory utilization detected",
			Penalty:  25,
		})

	case memory.UsedPercent >= 85:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityWarning,
			Source:   "memory",
			Message:  "High memory utilization detected",
			Penalty:  10,
		})
	}

	// Swap evaluation.
	switch {
	case memory.SwapPercent >= 50:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "memory",
			Message:  "Critical swap utilization detected",
			Penalty:  20,
		})

	case memory.SwapPercent >= 10:
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityWarning,
			Source:   "memory",
			Message:  "Swap utilization detected",
			Penalty:  10,
		})
	}

	return findings
}
