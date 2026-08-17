package history

import (
	"path/filepath"
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestCompactSnapshots(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"history.jsonl",
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
		{
			Health: domain.HealthResult{
				HealthScore: 70,
			},
		},
	}

	if err := CompactSnapshots(
		path,
		entries,
		2,
	); err != nil {

		t.Fatalf(
			"compaction failed: %v",
			err,
		)
	}

	loaded, err := LoadSnapshots(
		path,
		100,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(loaded) != 2 {

		t.Fatalf(
			"expected 2 entries, got %d",
			len(loaded),
		)
	}

	if loaded[0].Health.HealthScore != 80 {

		t.Fatalf(
			"expected oldest retained health 80, got %d",
			loaded[0].Health.HealthScore,
		)
	}

	if loaded[1].Health.HealthScore != 70 {

		t.Fatalf(
			"expected newest health 70, got %d",
			loaded[1].Health.HealthScore,
		)
	}
}
