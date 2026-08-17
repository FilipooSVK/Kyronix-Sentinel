package rules

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestPressureRuleHealthy(t *testing.T) {

	rule := NewPressureRule()

	snapshot := domain.Snapshot{
		Pressure: domain.PressureStats{},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 0 {
		t.Fatalf(
			"expected no findings, got %d",
			len(findings),
		)
	}
}

func TestPressureRuleHighIO(t *testing.T) {

	rule := NewPressureRule()

	snapshot := domain.Snapshot{
		Pressure: domain.PressureStats{
			IO: domain.ResourcePressureStats{
				Full: domain.PressureSample{
					Avg10: 15,
				},
			},
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
}

func TestPressureRuleCriticalMemory(t *testing.T) {

	rule := NewPressureRule()

	snapshot := domain.Snapshot{
		Pressure: domain.PressureStats{
			Memory: domain.ResourcePressureStats{
				Full: domain.PressureSample{
					Avg10: 25,
				},
			},
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 1 {
		t.Fatalf(
			"expected one finding, got %d",
			len(findings),
		)
	}

	if findings[0].Severity != domain.SeverityCritical {
		t.Errorf(
			"severity mismatch: got %s",
			findings[0].Severity,
		)
	}

	if findings[0].Penalty != 25 {
		t.Errorf(
			"penalty mismatch: got %d, want 25",
			findings[0].Penalty,
		)
	}
}

func TestPressureRuleMultipleSignals(t *testing.T) {

	rule := NewPressureRule()

	snapshot := domain.Snapshot{
		Pressure: domain.PressureStats{
			IO: domain.ResourcePressureStats{
				Full: domain.PressureSample{
					Avg10: 30,
				},
			},
			Memory: domain.ResourcePressureStats{
				Some: domain.PressureSample{
					Avg10: 15,
				},
			},
			CPU: domain.ResourcePressureStats{
				Some: domain.PressureSample{
					Avg10: 30,
				},
			},
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 3 {
		t.Fatalf(
			"expected three findings, got %d",
			len(findings),
		)
	}
}
