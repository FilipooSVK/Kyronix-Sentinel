package updater

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStateStoreSaveAndLoad(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update-state.json",
	)

	store := NewStateStore(
		path,
	)

	lastCheck := time.Date(
		2026,
		time.August,
		17,
		18,
		25,
		0,
		0,
		time.UTC,
	)

	state := UpdateState{
		LastCheck: lastCheck,

		NextCheck: lastCheck.Add(
			24 * time.Hour,
		),

		CurrentVersion: "v0.1.1",

		LatestVersion: "v0.1.2",

		UpdateAvailable: true,

		LastResult: UpdateResultSuccess,
	}

	if err := store.Save(
		state,
	); err != nil {

		t.Fatal(err)
	}

	loaded, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if loaded.CurrentVersion != "v0.1.1" {
		t.Fatalf(
			"unexpected current version: %s",
			loaded.CurrentVersion,
		)
	}

	if loaded.LatestVersion != "v0.1.2" {
		t.Fatalf(
			"unexpected latest version: %s",
			loaded.LatestVersion,
		)
	}

	if !loaded.UpdateAvailable {
		t.Fatal(
			"expected update available",
		)
	}

	if loaded.LastResult != UpdateResultSuccess {
		t.Fatalf(
			"unexpected result: %s",
			loaded.LastResult,
		)
	}
}

func TestStateStoreReplacesExistingState(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update-state.json",
	)

	store := NewStateStore(
		path,
	)

	now := time.Now().UTC()

	first := UpdateState{
		LastCheck: now,

		NextCheck: now.Add(
			time.Hour,
		),

		CurrentVersion: "v0.1.1",

		LatestVersion: "v0.1.1",

		LastResult: UpdateResultSuccess,
	}

	if err := store.Save(
		first,
	); err != nil {

		t.Fatal(err)
	}

	second := first

	second.LastCheck = now.Add(
		time.Minute,
	)

	second.NextCheck = second.LastCheck.Add(
		time.Hour,
	)

	second.LatestVersion = "v0.1.2"

	second.UpdateAvailable = true

	if err := store.Save(
		second,
	); err != nil {

		t.Fatal(err)
	}

	loaded, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if loaded.LatestVersion != "v0.1.2" {
		t.Fatalf(
			"expected replacement state, got %s",
			loaded.LatestVersion,
		)
	}
}

func TestStateStoreRejectsMalformedState(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update-state.json",
	)

	if err := os.WriteFile(
		path,
		[]byte(`{"broken":`),
		0644,
	); err != nil {

		t.Fatal(err)
	}

	store := NewStateStore(
		path,
	)

	if _, err := store.Load(); err == nil {

		t.Fatal(
			"expected malformed state error",
		)
	}
}

func TestStateStoreRejectsInvalidState(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update-state.json",
	)

	store := NewStateStore(
		path,
	)

	err := store.Save(
		UpdateState{},
	)

	if err == nil {

		t.Fatal(
			"expected invalid state error",
		)
	}
}
