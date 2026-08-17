package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
)

func TestAppendAndLoadSnapshots(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"history.jsonl",
	)

	now := time.Now().UTC()

	entry := SnapshotEntry{
		Timestamp: now,

		Snapshot: domain.Snapshot{
			Timestamp: now,

			Memory: domain.MemoryStats{
				UsedPercent: 42.5,
			},
		},

		Health: domain.HealthResult{
			HealthScore: 95,
		},
	}

	if err := AppendSnapshot(
		path,
		entry,
	); err != nil {

		t.Fatalf(
			"append snapshot failed: %v",
			err,
		)
	}

	entries, err := LoadSnapshots(
		path,
		100,
	)

	if err != nil {

		t.Fatalf(
			"load snapshots failed: %v",
			err,
		)
	}

	if len(entries) != 1 {

		t.Fatalf(
			"expected 1 entry, got %d",
			len(entries),
		)
	}

	if entries[0].Health.HealthScore != 95 {

		t.Fatalf(
			"expected health score 95, got %d",
			entries[0].Health.HealthScore,
		)
	}

	if entries[0].Snapshot.Memory.UsedPercent != 42.5 {

		t.Fatalf(
			"expected memory 42.5, got %f",
			entries[0].Snapshot.Memory.UsedPercent,
		)
	}
}

func TestLoadSnapshotsAppliesLimit(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"history.jsonl",
	)

	for i := 1; i <= 5; i++ {

		entry := SnapshotEntry{
			Timestamp: time.Now().UTC(),

			Health: domain.HealthResult{
				HealthScore: i,
			},
		}

		if err := AppendSnapshot(
			path,
			entry,
		); err != nil {

			t.Fatalf(
				"append failed: %v",
				err,
			)
		}
	}

	entries, err := LoadSnapshots(
		path,
		3,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 3 {

		t.Fatalf(
			"expected 3 entries, got %d",
			len(entries),
		)
	}

	if entries[0].Health.HealthScore != 3 {

		t.Fatalf(
			"expected oldest retained score 3, got %d",
			entries[0].Health.HealthScore,
		)
	}

	if entries[2].Health.HealthScore != 5 {

		t.Fatalf(
			"expected newest score 5, got %d",
			entries[2].Health.HealthScore,
		)
	}
}

func TestLoadSnapshotsIgnoresMalformedRecord(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"history.jsonl",
	)

	entry := SnapshotEntry{
		Timestamp: time.Now().UTC(),

		Health: domain.HealthResult{
			HealthScore: 88,
		},
	}

	if err := AppendSnapshot(
		path,
		entry,
	); err != nil {

		t.Fatal(err)
	}

	file, err := os.OpenFile(
		path,
		os.O_WRONLY|os.O_APPEND,
		0640,
	)

	if err != nil {
		t.Fatal(err)
	}

	if _, err := file.WriteString(
		`{"Timestamp":`,
	); err != nil {

		file.Close()
		t.Fatal(err)
	}

	file.Close()

	entries, err := LoadSnapshots(
		path,
		100,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {

		t.Fatalf(
			"expected 1 valid entry, got %d",
			len(entries),
		)
	}

	if entries[0].Health.HealthScore != 88 {

		t.Fatalf(
			"expected health score 88, got %d",
			entries[0].Health.HealthScore,
		)
	}
}

func TestLoadSnapshotsMissingFile(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"missing.jsonl",
	)

	entries, err := LoadSnapshots(
		path,
		100,
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 0 {

		t.Fatalf(
			"expected empty history, got %d",
			len(entries),
		)
	}
}
