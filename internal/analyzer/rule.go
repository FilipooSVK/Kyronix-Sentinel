package analyzer

import "kyronix/sentinel/internal/domain"

// Rule evaluates one aspect of system health.
type Rule interface {

	// Evaluate analyzes one snapshot and returns findings.
	Evaluate(snapshot domain.Snapshot) []domain.Finding
}
