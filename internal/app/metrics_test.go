package app

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestBuildCollectorMetricOK(
	t *testing.T,
) {
	metric := buildCollectorMetric(
		"host",
		domain.CollectorStatus{
			State:        domain.CollectorOK,
			CollectionMS: 7,
		},
	)

	if !metric.Up {
		t.Fatal(
			"expected collector metric to be up",
		)
	}

	if metric.CollectionDuration.Milliseconds() != 7 {
		t.Fatalf(
			"unexpected collection duration: %s",
			metric.CollectionDuration,
		)
	}
}

func TestBuildCollectorMetricError(
	t *testing.T,
) {
	metric := buildCollectorMetric(
		"memory",
		domain.CollectorStatus{
			State:        domain.CollectorError,
			CollectionMS: 3,
		},
	)

	if metric.Up {
		t.Fatal(
			"expected collector metric to be down",
		)
	}

	if metric.CollectionDuration.Milliseconds() != 3 {
		t.Fatalf(
			"unexpected collection duration: %s",
			metric.CollectionDuration,
		)
	}
}

func TestBuildCollectorMetricUnavailable(
	t *testing.T,
) {
	metric := buildCollectorMetric(
		"pressure",
		domain.CollectorStatus{
			State: domain.CollectorUnavailable,
		},
	)

	if metric.Up {
		t.Fatal(
			"unavailable collector must not be reported as up",
		)
	}
}
