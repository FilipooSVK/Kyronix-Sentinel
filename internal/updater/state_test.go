package updater

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func testUpdateState(
	now time.Time,
) UpdateState {

	return UpdateState{
		LastCheck: now,

		NextCheck: now.Add(
			24 * time.Hour,
		),

		CurrentVersion: "v0.1.1",

		LatestVersion: "v0.1.1",

		UpdateAvailable: false,

		LastResult: UpdateResultSuccess,
	}
}

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

	state := testUpdateState(
		lastCheck,
	)

	state.LatestVersion = "v0.1.2"

	state.UpdateAvailable = true

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

	first := testUpdateState(
		now,
	)

	if err := store.Save(
		first,
	); err != nil {

		t.Fatal(err)
	}

	second := testUpdateState(
		now.Add(
			time.Minute,
		),
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

func TestLegacyStateWithoutInstallLifecycleLoads(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update-state.json",
	)

	data := []byte(`{
  "last_check": "2026-08-17T18:25:00Z",
  "next_check": "2026-08-18T18:25:00Z",
  "current_version": "v0.1.1",
  "latest_version": "v0.1.1",
  "update_available": false,
  "last_result": "success"
}
`)

	if err := os.WriteFile(
		path,
		data,
		0644,
	); err != nil {

		t.Fatal(err)
	}

	store := NewStateStore(
		path,
	)

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.LastInstallAttempt != nil {

		t.Fatal(
			"legacy state unexpectedly contains install lifecycle",
		)
	}
}

func TestRecordInstallSuccess(
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

	if err := store.Save(
		testUpdateState(now),
	); err != nil {

		t.Fatal(err)
	}

	attempt := now.Add(
		time.Minute,
	)

	if err := store.RecordInstallStarted(
		attempt,
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.LastInstallResult !=
		InstallResultInProgress {

		t.Fatalf(
			"expected in-progress install, got %s",
			state.LastInstallResult,
		)
	}

	if err := store.RecordInstallSuccess(); err != nil {

		t.Fatal(err)
	}

	state, err = store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.LastInstallResult !=
		InstallResultSuccess {

		t.Fatalf(
			"expected successful install, got %s",
			state.LastInstallResult,
		)
	}

	if state.LastInstalledVersion != "v0.1.2" {

		t.Fatalf(
			"unexpected installed version: %s",
			state.LastInstalledVersion,
		)
	}

	if state.LastRollback {

		t.Fatal(
			"successful install unexpectedly rolled back",
		)
	}
}

func TestRecordInstallFailureWithVerifiedRollback(
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

	if err := store.Save(
		testUpdateState(now),
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallStarted(
		now.Add(time.Minute),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"post-install health check failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.LastInstallResult !=
		InstallResultFailed {

		t.Fatalf(
			"expected failed install, got %s",
			state.LastInstallResult,
		)
	}

	if !state.LastRollback {

		t.Fatal(
			"expected rollback attempt",
		)
	}

	if !state.LastRollbackVerified {

		t.Fatal(
			"expected verified rollback",
		)
	}

	if state.RecoveredVersion != "v0.1.1" {

		t.Fatalf(
			"unexpected recovered version: %s",
			state.RecoveredVersion,
		)
	}
}

func TestSaveCheckPreservesInstallLifecycle(
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

	if err := store.Save(
		testUpdateState(now),
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallStarted(
		now.Add(time.Minute),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallSuccess(); err != nil {

		t.Fatal(err)
	}

	nextCheck := testUpdateState(
		now.Add(
			2 * time.Hour,
		),
	)

	nextCheck.CurrentVersion = "v0.1.2"

	nextCheck.LatestVersion = "v0.1.2"

	if err := store.SaveCheck(
		nextCheck,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.CurrentVersion != "v0.1.2" {

		t.Fatalf(
			"unexpected current version: %s",
			state.CurrentVersion,
		)
	}

	if state.LastInstallResult !=
		InstallResultSuccess {

		t.Fatalf(
			"install lifecycle was lost: %s",
			state.LastInstallResult,
		)
	}

	if state.LastInstalledVersion != "v0.1.2" {

		t.Fatalf(
			"installed version was lost: %s",
			state.LastInstalledVersion,
		)
	}
}

func TestRecordInstallStartedRejectsNonNewerVersion(
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

	if err := store.Save(
		testUpdateState(now),
	); err != nil {

		t.Fatal(err)
	}

	err := store.RecordInstallStarted(
		now.Add(time.Minute),
		"v0.1.1",
		"v0.1.1",
	)

	if err == nil {

		t.Fatal(
			"expected non-newer target version error",
		)
	}
}
