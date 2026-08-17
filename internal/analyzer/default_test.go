package analyzer

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestDefaultAnalyzerHealthySystem(t *testing.T) {

	analyzer := NewDefaultAnalyzer()

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 40,
			SwapPercent: 0,
		},
	}

	result := analyzer.Analyze(snapshot)

	if result.HealthScore != 100 {
		t.Errorf(
			"HealthScore mismatch: got %d, want 100",
			result.HealthScore,
		)
	}

	if result.FreezeRisk != domain.RiskLow {
		t.Errorf(
			"FreezeRisk mismatch: got %s, want %s",
			result.FreezeRisk,
			domain.RiskLow,
		)
	}

	if len(result.Findings) != 0 {
		t.Errorf(
			"expected no findings, got %d",
			len(result.Findings),
		)
	}
}

func TestDefaultAnalyzerDetectsDegradedSystem(t *testing.T) {

	analyzer := NewDefaultAnalyzer()

	oomKills := uint64(1)

	snapshot := domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 96,
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
				SystemKillCount: &oomKills,
			},
		},

		Disks: []domain.DiskStats{
			{
				MountPoint:     "/",
				UsedPercent:    97,
				AvailableBytes: 500 * 1024 * 1024,
			},
		},
	}

	result := analyzer.Analyze(snapshot)

	if result.HealthScore >= 100 {
		t.Errorf(
			"expected degraded health score, got %d",
			result.HealthScore,
		)
	}

	if len(result.Findings) == 0 {
		t.Fatal(
			"expected findings but got none",
		)
	}

	if result.FreezeRisk == domain.RiskLow {
		t.Errorf(
			"expected elevated risk, got %s",
			result.FreezeRisk,
		)
	}
}
