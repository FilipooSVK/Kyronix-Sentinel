package analyzer

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestRiskCalculatorHealthySystem(t *testing.T) {

	calculator := NewRiskCalculator()

	score := calculator.Calculate(
		domain.Snapshot{},
		nil,
	)

	if score != 0 {
		t.Errorf(
			"risk score mismatch: got %d, want 0",
			score,
		)
	}
}

func TestRiskCalculatorOOM(t *testing.T) {

	value := uint64(1)

	calculator := NewRiskCalculator()

	snapshot := domain.Snapshot{
		Kernel: domain.KernelStats{
			OOM: domain.OOMStats{
				SystemKillCount: &value,
			},
		},
	}

	score := calculator.Calculate(
		snapshot,
		nil,
	)

	if score < 50 {
		t.Errorf(
			"expected high risk score, got %d",
			score,
		)
	}
}

func TestRiskCalculatorIOPressure(t *testing.T) {

	calculator := NewRiskCalculator()

	snapshot := domain.Snapshot{
		Pressure: domain.PressureStats{
			IO: domain.ResourcePressureStats{
				Full: domain.PressureSample{
					Avg10: 25,
				},
			},
		},
	}

	score := calculator.Calculate(
		snapshot,
		nil,
	)

	if score != 30 {
		t.Errorf(
			"risk score mismatch: got %d, want 30",
			score,
		)
	}
}

func TestRiskCalculatorMultipleSignals(t *testing.T) {

	oom := uint64(1)

	calculator := NewRiskCalculator()

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			SwapPercent: 70,
		},

		Pressure: domain.PressureStats{
			IO: domain.ResourcePressureStats{
				Full: domain.PressureSample{
					Avg10: 25,
				},
			},
		},

		Kernel: domain.KernelStats{
			OOM: domain.OOMStats{
				SystemKillCount: &oom,
			},
		},
	}

	score := calculator.Calculate(
		snapshot,
		[]domain.Finding{
			{
				Severity: domain.SeverityCritical,
			},
		},
	)

	if score != 100 {
		t.Errorf(
			"risk score mismatch: got %d, want 100",
			score,
		)
	}
}

func TestRiskFromScore(t *testing.T) {

	tests := []struct {
		score int
		want  domain.FreezeRisk
	}{
		{0, domain.RiskLow},
		{30, domain.RiskMedium},
		{60, domain.RiskHigh},
		{80, domain.RiskCritical},
	}

	for _, test := range tests {

		got := RiskFromScore(test.score)

		if got != test.want {
			t.Errorf(
				"score %d got %s want %s",
				test.score,
				got,
				test.want,
			)
		}
	}
}
