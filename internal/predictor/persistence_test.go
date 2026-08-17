package predictor

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

func TestAnalyzePersistenceDetectsPersistentCPUPressure(
	t *testing.T,
) {

	start := time.Now()

	entries := make(
		[]history.SnapshotEntry,
		8,
	)

	for i := range entries {

		cpuPressure := 10.0

		if i >= 2 {
			cpuPressure = 45
		}

		entries[i] = history.SnapshotEntry{
			Timestamp: start.Add(
				time.Duration(i) * time.Minute,
			),

			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					CPU: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: cpuPressure,
						},
					},
				},
			},

			Health: domain.HealthResult{
				HealthScore: 100,
			},
		}
	}

	report := AnalyzePersistence(
		entries,
		8,
	)

	if report.CPUPressure.Hits != 6 {

		t.Fatalf(
			"expected 6 CPU pressure hits, got %d",
			report.CPUPressure.Hits,
		)
	}

	if report.CPUPressure.Samples != 8 {

		t.Fatalf(
			"expected 8 samples, got %d",
			report.CPUPressure.Samples,
		)
	}

	if report.CPUPressure.Ratio != 0.75 {

		t.Fatalf(
			"expected ratio 0.75, got %f",
			report.CPUPressure.Ratio,
		)
	}

	if !report.CPUPressure.Persistent(
		0.75,
		5,
	) {

		t.Fatal(
			"expected CPU pressure to be persistent",
		)
	}
}

func TestAnalyzePersistenceIgnoresSingleSpike(
	t *testing.T,
) {

	entries := make(
		[]history.SnapshotEntry,
		8,
	)

	for i := range entries {

		ioPressure := 5.0

		if i == 4 {
			ioPressure = 80
		}

		entries[i] = history.SnapshotEntry{
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					IO: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: ioPressure,
						},
					},
				},
			},

			Health: domain.HealthResult{
				HealthScore: 100,
			},
		}
	}

	report := AnalyzePersistence(
		entries,
		8,
	)

	if report.IOPressure.Hits != 1 {

		t.Fatalf(
			"expected 1 I/O pressure hit, got %d",
			report.IOPressure.Hits,
		)
	}

	if report.IOPressure.Persistent(
		0.75,
		5,
	) {

		t.Fatal(
			"single I/O spike must not be persistent",
		)
	}
}

func TestAnalyzePersistenceUsesRecentWindow(
	t *testing.T,
) {

	entries := make(
		[]history.SnapshotEntry,
		8,
	)

	for i := range entries {

		memory := 90.0

		if i >= 4 {
			memory = 40
		}

		entries[i] = history.SnapshotEntry{
			Snapshot: domain.Snapshot{
				Memory: domain.MemoryStats{
					UsedPercent: memory,
				},
			},

			Health: domain.HealthResult{
				HealthScore: 100,
			},
		}
	}

	report := AnalyzePersistence(
		entries,
		4,
	)

	if report.MemoryUtilization.Hits != 0 {

		t.Fatalf(
			"expected old high-memory samples to be ignored, got %d hits",
			report.MemoryUtilization.Hits,
		)
	}

	if report.MemoryUtilization.Samples != 4 {

		t.Fatalf(
			"expected 4 recent samples, got %d",
			report.MemoryUtilization.Samples,
		)
	}
}

func TestPersistenceRequiresMinimumSamples(
	t *testing.T,
) {

	signal := PersistenceSignal{
		Metric:  "cpu_pressure",
		Hits:    2,
		Samples: 2,
		Ratio:   1,
	}

	if signal.Persistent(
		0.75,
		5,
	) {

		t.Fatal(
			"signal with insufficient history must not be persistent",
		)
	}
}

func TestAnalyzePersistenceMultipleSignals(
	t *testing.T,
) {

	entries := make(
		[]history.SnapshotEntry,
		8,
	)

	for i := range entries {

		entries[i] = history.SnapshotEntry{
			Snapshot: domain.Snapshot{
				Memory: domain.MemoryStats{
					UsedPercent: 85,
				},

				Pressure: domain.PressureStats{
					Memory: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 20,
						},
					},
				},
			},

			Health: domain.HealthResult{
				HealthScore: 60,
			},
		}
	}

	report := AnalyzePersistence(
		entries,
		8,
	)

	if !report.MemoryUtilization.Persistent(
		0.75,
		5,
	) {

		t.Fatal(
			"expected persistent memory utilization",
		)
	}

	if !report.MemoryPressure.Persistent(
		0.75,
		5,
	) {

		t.Fatal(
			"expected persistent memory pressure",
		)
	}

	if !report.HealthDegradation.Persistent(
		0.75,
		5,
	) {

		t.Fatal(
			"expected persistent health degradation",
		)
	}
}
