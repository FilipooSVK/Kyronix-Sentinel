package predictor

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

func TestAnalyzeHealthTrendDecreasing(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Health: domain.HealthResult{
				HealthScore: 100,
			},
		},
		{
			Timestamp: start.Add(
				2 * time.Hour,
			),
			Health: domain.HealthResult{
				HealthScore: 70,
			},
		},
	}

	trend := AnalyzeHealthTrend(
		entries,
	)

	if trend.Direction != TrendDecreasing {
		t.Fatalf(
			"expected DECREASING, got %s",
			trend.Direction,
		)
	}

	if trend.Previous != 100 {
		t.Fatalf(
			"expected previous 100, got %f",
			trend.Previous,
		)
	}

	if trend.Current != 70 {
		t.Fatalf(
			"expected current 70, got %f",
			trend.Current,
		)
	}

	if trend.Delta != -30 {
		t.Fatalf(
			"expected delta -30, got %f",
			trend.Delta,
		)
	}

	if trend.Rate != -15 {
		t.Fatalf(
			"expected rate -15, got %f",
			trend.Rate,
		)
	}
}

func TestAnalyzeHealthTrendImproving(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Health: domain.HealthResult{
				HealthScore: 60,
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Health: domain.HealthResult{
				HealthScore: 90,
			},
		},
	}

	trend := AnalyzeHealthTrend(
		entries,
	)

	if trend.Direction != TrendIncreasing {
		t.Fatalf(
			"expected INCREASING, got %s",
			trend.Direction,
		)
	}

	if trend.Delta != 30 {
		t.Fatalf(
			"expected delta 30, got %f",
			trend.Delta,
		)
	}
}

func TestAnalyzeHealthTrendStable(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Health: domain.HealthResult{
				HealthScore: 90,
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Health: domain.HealthResult{
				HealthScore: 90,
			},
		},
	}

	trend := AnalyzeHealthTrend(
		entries,
	)

	if trend.Direction != TrendStable {
		t.Fatalf(
			"expected STABLE, got %s",
			trend.Direction,
		)
	}

	if trend.Delta != 0 {
		t.Fatalf(
			"expected delta 0, got %f",
			trend.Delta,
		)
	}
}

func TestAnalyzeHealthTrendInsufficientHistory(
	t *testing.T,
) {

	trend := AnalyzeHealthTrend(
		nil,
	)

	if trend.Metric != "health_score" {
		t.Fatalf(
			"unexpected metric: %s",
			trend.Metric,
		)
	}

	if trend.Direction != TrendStable {
		t.Fatalf(
			"expected STABLE, got %s",
			trend.Direction,
		)
	}
}
