package predictor

import "kyronix/sentinel/internal/history"

const (
	persistentMemoryUtilizationThreshold = 80.0
	persistentCPUPressureThreshold       = 40.0
	persistentMemoryPressureThreshold    = 15.0
	persistentIOPressureThreshold        = 40.0
	persistentHealthThreshold            = 70.0
)

// PersistenceSignal describes how often a condition occurred
// during the analyzed sample window.
type PersistenceSignal struct {
	Metric string

	Hits int

	Samples int

	Ratio float64
}

// Persistent reports whether the signal satisfies the required
// sample count and occurrence ratio.
func (s PersistenceSignal) Persistent(
	minRatio float64,
	minSamples int,
) bool {

	if s.Samples < minSamples {
		return false
	}

	return s.Ratio >= minRatio
}

// PersistenceReport contains persistent degradation signals.
type PersistenceReport struct {
	MemoryUtilization PersistenceSignal

	CPUPressure PersistenceSignal

	MemoryPressure PersistenceSignal

	IOPressure PersistenceSignal

	HealthDegradation PersistenceSignal
}

// AnalyzePersistence analyzes the most recent snapshots and determines
// how frequently degradation conditions were present.
func AnalyzePersistence(
	entries []history.SnapshotEntry,
	sampleWindow int,
) PersistenceReport {

	report := PersistenceReport{
		MemoryUtilization: PersistenceSignal{
			Metric: "memory_utilization",
		},

		CPUPressure: PersistenceSignal{
			Metric: "cpu_pressure",
		},

		MemoryPressure: PersistenceSignal{
			Metric: "memory_pressure",
		},

		IOPressure: PersistenceSignal{
			Metric: "io_pressure",
		},

		HealthDegradation: PersistenceSignal{
			Metric: "health_degradation",
		},
	}

	if len(entries) == 0 || sampleWindow <= 0 {
		return report
	}

	start := len(entries) - sampleWindow

	if start < 0 {
		start = 0
	}

	samples := entries[start:]

	sampleCount := len(samples)

	report.MemoryUtilization.Samples = sampleCount
	report.CPUPressure.Samples = sampleCount
	report.MemoryPressure.Samples = sampleCount
	report.IOPressure.Samples = sampleCount
	report.HealthDegradation.Samples = sampleCount

	for _, entry := range samples {

		if entry.Snapshot.Memory.UsedPercent >=
			persistentMemoryUtilizationThreshold {

			report.MemoryUtilization.Hits++
		}

		if entry.Snapshot.Pressure.CPU.Some.Avg10 >=
			persistentCPUPressureThreshold {

			report.CPUPressure.Hits++
		}

		if entry.Snapshot.Pressure.Memory.Some.Avg10 >=
			persistentMemoryPressureThreshold {

			report.MemoryPressure.Hits++
		}

		if entry.Snapshot.Pressure.IO.Some.Avg10 >=
			persistentIOPressureThreshold {

			report.IOPressure.Hits++
		}

		if float64(entry.Health.HealthScore) <=
			persistentHealthThreshold {

			report.HealthDegradation.Hits++
		}
	}

	report.MemoryUtilization.Ratio = persistenceRatio(
		report.MemoryUtilization.Hits,
		sampleCount,
	)

	report.CPUPressure.Ratio = persistenceRatio(
		report.CPUPressure.Hits,
		sampleCount,
	)

	report.MemoryPressure.Ratio = persistenceRatio(
		report.MemoryPressure.Hits,
		sampleCount,
	)

	report.IOPressure.Ratio = persistenceRatio(
		report.IOPressure.Hits,
		sampleCount,
	)

	report.HealthDegradation.Ratio = persistenceRatio(
		report.HealthDegradation.Hits,
		sampleCount,
	)

	return report
}

func persistenceRatio(
	hits int,
	samples int,
) float64 {

	if samples == 0 {
		return 0
	}

	return float64(hits) /
		float64(samples)
}
