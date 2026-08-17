package predictor

import (
	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

// AnalyzeCPUPressureTrend analyzes CPU PSI some.avg10 development.
func AnalyzeCPUPressureTrend(
	entries []history.SnapshotEntry,
) Trend {

	return analyzePressureTrend(
		entries,
		"cpu_pressure_some_avg10",
		func(snapshot domain.Snapshot) float64 {
			return snapshot.Pressure.CPU.Some.Avg10
		},
	)
}

// AnalyzeMemoryPressureTrend analyzes memory PSI some.avg10 development.
func AnalyzeMemoryPressureTrend(
	entries []history.SnapshotEntry,
) Trend {

	return analyzePressureTrend(
		entries,
		"memory_pressure_some_avg10",
		func(snapshot domain.Snapshot) float64 {
			return snapshot.Pressure.Memory.Some.Avg10
		},
	)
}

// AnalyzeIOPressureTrend analyzes I/O PSI some.avg10 development.
func AnalyzeIOPressureTrend(
	entries []history.SnapshotEntry,
) Trend {

	return analyzePressureTrend(
		entries,
		"io_pressure_some_avg10",
		func(snapshot domain.Snapshot) float64 {
			return snapshot.Pressure.IO.Some.Avg10
		},
	)
}

func analyzePressureTrend(
	entries []history.SnapshotEntry,
	metric string,
	value func(domain.Snapshot) float64,
) Trend {

	trend := Trend{
		Metric:    metric,
		Direction: TrendStable,
	}

	if len(entries) < 2 {
		return trend
	}

	first := entries[0]
	last := entries[len(entries)-1]

	previous := value(first.Snapshot)
	current := value(last.Snapshot)

	delta := current - previous

	window := last.Timestamp.Sub(
		first.Timestamp,
	)

	rate := 0.0

	if window > 0 {
		rate = delta / window.Hours()
	}

	direction := TrendStable

	if delta > 0 {
		direction = TrendIncreasing
	}

	if delta < 0 {
		direction = TrendDecreasing
	}

	return Trend{
		Metric:    metric,
		Direction: direction,
		Current:   current,
		Previous:  previous,
		Delta:     delta,
		Rate:      rate,
		Window:    window,
	}
}
