package predictor

import (
	"kyronix/sentinel/internal/history"
)

// AnalyzeMemoryTrend calculates memory usage trend.
func AnalyzeMemoryTrend(
	entries []history.SnapshotEntry,
) Trend {

	if len(entries) < 2 {

		return Trend{
			Metric:    "memory",
			Direction: TrendStable,
		}
	}

	first := entries[0]

	last := entries[len(entries)-1]

	previous := first.Snapshot.Memory.UsedPercent

	current := last.Snapshot.Memory.UsedPercent

	delta := current - previous

	direction := TrendStable

	switch {

	case delta > 0:
		direction = TrendIncreasing

	case delta < 0:
		direction = TrendDecreasing
	}

	window := last.Timestamp.Sub(
		first.Timestamp,
	)

	rate := 0.0

	if window.Hours() > 0 {

		rate = delta / window.Hours()
	}

	return Trend{

		Metric: "memory",

		Direction: direction,

		Current: current,

		Previous: previous,

		Delta: delta,

		Rate: rate,

		Window: window,
	}
}
