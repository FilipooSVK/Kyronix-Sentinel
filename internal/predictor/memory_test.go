package predictor

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

func TestAnalyzeMemoryTrendIncreasing(t *testing.T) {

	now := time.Now()

	entries := []history.SnapshotEntry{

		{
			Timestamp: now.Add(-2 * time.Hour),

			Snapshot: domain.Snapshot{
				Memory: domain.MemoryStats{
					UsedPercent: 40,
				},
			},
		},

		{
			Timestamp: now,

			Snapshot: domain.Snapshot{
				Memory: domain.MemoryStats{
					UsedPercent: 70,
				},
			},
		},
	}

	trend := AnalyzeMemoryTrend(
		entries,
	)

	if trend.Direction != TrendIncreasing {

		t.Fatalf(
			"expected increasing trend, got %s",
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

func TestAnalyzeMemoryTrendStable(t *testing.T) {

	now := time.Now()

	entries := []history.SnapshotEntry{

		{
			Timestamp: now.Add(-time.Hour),

			Snapshot: domain.Snapshot{
				Memory: domain.MemoryStats{
					UsedPercent: 50,
				},
			},
		},

		{
			Timestamp: now,

			Snapshot: domain.Snapshot{
				Memory: domain.MemoryStats{
					UsedPercent: 50,
				},
			},
		},
	}

	trend := AnalyzeMemoryTrend(
		entries,
	)

	if trend.Direction != TrendStable {

		t.Fatalf(
			"expected stable trend, got %s",
			trend.Direction,
		)
	}
}
