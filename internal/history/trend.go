package history

import "kyronix/sentinel/internal/domain"

// TrendDirection describes health movement.
type TrendDirection string

const (
	TrendStable    TrendDirection = "STABLE"
	TrendImproving TrendDirection = "IMPROVING"
	TrendDegrading TrendDirection = "DEGRADING"
)

// CalculateHealthTrend compares historical health results.
func CalculateHealthTrend(
	results []domain.HealthResult,
) TrendDirection {

	if len(results) < 2 {
		return TrendStable
	}

	first := results[0].HealthScore
	last := results[len(results)-1].HealthScore

	switch {

	case last < first-10:
		return TrendDegrading

	case last > first+10:
		return TrendImproving

	default:
		return TrendStable
	}
}
