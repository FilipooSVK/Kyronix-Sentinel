package history

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
)

func TestRestoreSnapshots(
	t *testing.T,
) {

	store := NewStore(
		3,
	)

	entries := []SnapshotEntry{
		{
			Timestamp: time.Now(),

			Health: domain.HealthResult{
				HealthScore: 90,
			},
		},
		{
			Timestamp: time.Now(),

			Health: domain.HealthResult{
				HealthScore: 80,
			},
		},
	}

	store.RestoreSnapshots(
		entries,
	)

	snapshots := store.GetSnapshots()

	if len(snapshots) != 2 {

		t.Fatalf(
			"expected 2 snapshots, got %d",
			len(snapshots),
		)
	}

	results := store.GetAll()

	if len(results) != 2 {

		t.Fatalf(
			"expected 2 health results, got %d",
			len(results),
		)
	}

	if results[1].HealthScore != 80 {

		t.Fatalf(
			"expected latest health score 80, got %d",
			results[1].HealthScore,
		)
	}
}

func TestRestoreSnapshotsRespectsStoreLimit(
	t *testing.T,
) {

	store := NewStore(
		2,
	)

	entries := []SnapshotEntry{
		{
			Health: domain.HealthResult{
				HealthScore: 100,
			},
		},
		{
			Health: domain.HealthResult{
				HealthScore: 90,
			},
		},
		{
			Health: domain.HealthResult{
				HealthScore: 80,
			},
		},
	}

	store.RestoreSnapshots(
		entries,
	)

	snapshots := store.GetSnapshots()

	if len(snapshots) != 2 {

		t.Fatalf(
			"expected 2 retained snapshots, got %d",
			len(snapshots),
		)
	}

	if snapshots[0].Health.HealthScore != 90 {

		t.Fatalf(
			"expected oldest retained score 90, got %d",
			snapshots[0].Health.HealthScore,
		)
	}
}
