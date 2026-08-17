package predictor

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

func TestAnalyzeCPUPressureTrendIncreasing(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					CPU: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 10,
						},
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				2 * time.Hour,
			),
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					CPU: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 40,
						},
					},
				},
			},
		},
	}

	trend := AnalyzeCPUPressureTrend(
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

	if trend.Rate != 15 {
		t.Fatalf(
			"expected rate 15, got %f",
			trend.Rate,
		)
	}
}

func TestAnalyzeMemoryPressureTrendIncreasing(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					Memory: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 2,
						},
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					Memory: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 8,
						},
					},
				},
			},
		},
	}

	trend := AnalyzeMemoryPressureTrend(
		entries,
	)

	if trend.Direction != TrendIncreasing {
		t.Fatalf(
			"expected INCREASING, got %s",
			trend.Direction,
		)
	}

	if trend.Current != 8 {
		t.Fatalf(
			"expected current 8, got %f",
			trend.Current,
		)
	}
}

func TestAnalyzeIOPressureTrendStable(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					IO: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 5,
						},
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Snapshot: domain.Snapshot{
				Pressure: domain.PressureStats{
					IO: domain.ResourcePressureStats{
						Some: domain.PressureSample{
							Avg10: 5,
						},
					},
				},
			},
		},
	}

	trend := AnalyzeIOPressureTrend(
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

func TestAnalyzePressureTrendInsufficientHistory(
	t *testing.T,
) {

	trend := AnalyzeCPUPressureTrend(
		nil,
	)

	if trend.Direction != TrendStable {
		t.Fatalf(
			"expected STABLE, got %s",
			trend.Direction,
		)
	}

	if trend.Metric != "cpu_pressure_some_avg10" {
		t.Fatalf(
			"unexpected metric: %s",
			trend.Metric,
		)
	}
}
