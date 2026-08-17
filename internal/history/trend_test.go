package history

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestTrendDegrading(t *testing.T) {

	results := []domain.HealthResult{

		{
			HealthScore: 100,
		},

		{
			HealthScore: 70,
		},
	}

	trend := CalculateHealthTrend(results)

	if trend != TrendDegrading {
		t.Errorf(
			"expected degrading, got %s",
			trend,
		)
	}
}

func TestTrendStable(t *testing.T) {

	results := []domain.HealthResult{

		{
			HealthScore: 90,
		},

		{
			HealthScore: 85,
		},
	}

	trend := CalculateHealthTrend(results)

	if trend != TrendStable {
		t.Errorf(
			"expected stable, got %s",
			trend,
		)
	}
}
