package predictor

import "kyronix/sentinel/internal/history"

// AnalyzeHealthTrend analyzes HealthScore development over time.
func AnalyzeHealthTrend(
	entries []history.SnapshotEntry,
) Trend {

	trend := Trend{
		Metric:    "health_score",
		Direction: TrendStable,
	}

	if len(entries) < 2 {
		return trend
	}

	first := entries[0]
	last := entries[len(entries)-1]

	previous := float64(
		first.Health.HealthScore,
	)

	current := float64(
		last.Health.HealthScore,
	)

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
		Metric:    "health_score",
		Direction: direction,
		Current:   current,
		Previous:  previous,
		Delta:     delta,
		Rate:      rate,
		Window:    window,
	}
}
