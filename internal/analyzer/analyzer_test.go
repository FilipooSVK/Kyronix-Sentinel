package analyzer

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
)

type fakeRule struct {
	findings []domain.Finding
}

func (f fakeRule) Evaluate(
	domain.Snapshot,
) []domain.Finding {
	return f.findings
}

func TestAnalyzerAnalyze(t *testing.T) {

	analyzer := NewAnalyzer(
		fakeRule{
			findings: []domain.Finding{
				{
					Severity: domain.SeverityWarning,
					Source:   "test",
					Message:  "test warning",
					Penalty:  10,
				},
			},
		},
	)

	snapshot := domain.Snapshot{
		Timestamp: time.Now(),
	}

	result := analyzer.Analyze(snapshot)

	if result.HealthScore != 90 {
		t.Errorf(
			"HealthScore mismatch: got %d, want 90",
			result.HealthScore,
		)
	}

	if result.FreezeRiskScore != 0 {
		t.Errorf(
			"FreezeRiskScore mismatch: got %d, want 0",
			result.FreezeRiskScore,
		)
	}

	if result.FreezeRisk != domain.RiskLow {
		t.Errorf(
			"FreezeRisk mismatch: got %s",
			result.FreezeRisk,
		)
	}

	if len(result.Findings) != 1 {
		t.Errorf(
			"Findings count mismatch: got %d",
			len(result.Findings),
		)
	}

	if result.Quality != domain.AssessmentComplete {
		t.Errorf(
			"Quality mismatch: got %s",
			result.Quality,
		)
	}
}

func TestAnalyzerDetectsFreezeRisk(t *testing.T) {

	oom := uint64(1)

	analyzer := NewAnalyzer()

	snapshot := domain.Snapshot{

		Kernel: domain.KernelStats{
			OOM: domain.OOMStats{
				SystemKillCount: &oom,
			},
		},

		Pressure: domain.PressureStats{
			IO: domain.ResourcePressureStats{
				Full: domain.PressureSample{
					Avg10: 25,
				},
			},
		},
	}

	result := analyzer.Analyze(snapshot)

	if result.FreezeRiskScore < 80 {
		t.Errorf(
			"expected high freeze risk, got %d",
			result.FreezeRiskScore,
		)
	}

	if result.FreezeRisk != domain.RiskCritical {
		t.Errorf(
			"expected critical risk, got %s",
			result.FreezeRisk,
		)
	}
}
