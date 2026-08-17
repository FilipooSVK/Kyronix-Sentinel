package rules

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestMemoryRuleHealthy(t *testing.T) {

	rule := NewMemoryRule()

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 40,
			SwapPercent: 0,
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 0 {
		t.Fatalf(
			"expected no findings, got %d",
			len(findings),
		)
	}
}

func TestMemoryRuleHighUsage(t *testing.T) {

	rule := NewMemoryRule()

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 90,
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 1 {
		t.Fatalf(
			"expected one finding, got %d",
			len(findings),
		)
	}

	if findings[0].Penalty != 10 {
		t.Errorf(
			"penalty mismatch: got %d, want 10",
			findings[0].Penalty,
		)
	}

	if findings[0].Severity != domain.SeverityWarning {
		t.Errorf(
			"severity mismatch: got %s",
			findings[0].Severity,
		)
	}
}

func TestMemoryRuleCriticalSwap(t *testing.T) {

	rule := NewMemoryRule()

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 70,
			SwapPercent: 80,
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 1 {
		t.Fatalf(
			"expected one finding, got %d",
			len(findings),
		)
	}

	if findings[0].Penalty != 20 {
		t.Errorf(
			"penalty mismatch: got %d, want 20",
			findings[0].Penalty,
		)
	}

	if findings[0].Severity != domain.SeverityCritical {
		t.Errorf(
			"severity mismatch: got %s",
			findings[0].Severity,
		)
	}
}

func TestMemoryRuleMultipleFindings(t *testing.T) {

	rule := NewMemoryRule()

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 96,
			SwapPercent: 60,
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 2 {
		t.Fatalf(
			"expected two findings, got %d",
			len(findings),
		)
	}
}
