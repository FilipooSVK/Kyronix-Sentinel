package updater

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newQuarantineTestStore(
	t *testing.T,
) *StateStore {

	t.Helper()

	store := NewStateStore(
		filepath.Join(
			t.TempDir(),
			"update-state.json",
		),
	)

	now := time.Now().UTC()

	if err := store.Save(
		testUpdateState(
			now,
		),
	); err != nil {

		t.Fatal(err)
	}

	return store
}

func TestRollbackQuarantinesFailedRelease(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	const installError = "post-install health check failed"

	if err := store.RecordInstallFailure(
		installError,
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.QuarantinedVersion != "v0.1.2" {

		t.Fatalf(
			"unexpected quarantined version: %s",
			state.QuarantinedVersion,
		)
	}

	if state.QuarantinedAt == nil ||
		state.QuarantinedAt.IsZero() {

		t.Fatal(
			"expected quarantine timestamp",
		)
	}

	if state.QuarantineReason !=
		QuarantineReasonActivationFailure {

		t.Fatalf(
			"unexpected quarantine reason: %s",
			state.QuarantineReason,
		)
	}

	if state.QuarantineFailureCount != 1 {

		t.Fatalf(
			"unexpected failure count: %d",
			state.QuarantineFailureCount,
		)
	}

	if state.QuarantineLastError !=
		installError {

		t.Fatalf(
			"unexpected quarantine error: %s",
			state.QuarantineLastError,
		)
	}

	quarantined, err :=
		store.IsVersionQuarantined(
			"v0.1.2",
		)

	if err != nil {
		t.Fatal(err)
	}

	if !quarantined {

		t.Fatal(
			"expected v0.1.2 to be quarantined",
		)
	}
}

func TestFailureWithoutRollbackDoesNotQuarantine(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"download verification failed",
		false,
		false,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.QuarantinedVersion != "" {

		t.Fatalf(
			"unexpected quarantine: %s",
			state.QuarantinedVersion,
		)
	}
}

func TestRepeatedRollbackIncrementsFailureCount(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	for i := 0; i < 2; i++ {

		if err := store.RecordInstallStarted(
			time.Now().UTC(),
			"v0.1.1",
			"v0.1.2",
		); err != nil {

			t.Fatal(err)
		}

		if err := store.RecordInstallFailure(
			"activation failed",
			true,
			true,
		); err != nil {

			t.Fatal(err)
		}
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.QuarantineFailureCount != 2 {

		t.Fatalf(
			"expected failure count 2, got %d",
			state.QuarantineFailureCount,
		)
	}
}

func TestDifferentFailedReleaseReplacesQuarantine(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"activation failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.3",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"second activation failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.QuarantinedVersion != "v0.1.3" {

		t.Fatalf(
			"unexpected quarantined version: %s",
			state.QuarantinedVersion,
		)
	}

	if state.QuarantineFailureCount != 1 {

		t.Fatalf(
			"expected reset failure count, got %d",
			state.QuarantineFailureCount,
		)
	}
}

func TestSaveCheckPreservesQuarantine(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"activation failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	check := testUpdateState(
		time.Now().UTC().Add(
			time.Hour,
		),
	)

	check.LatestVersion = "v0.1.2"

	check.UpdateAvailable = true

	if err := store.SaveCheck(
		check,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.QuarantinedVersion != "v0.1.2" {

		t.Fatalf(
			"quarantine lost after SaveCheck: %s",
			state.QuarantinedVersion,
		)
	}

	if state.QuarantineFailureCount != 1 {

		t.Fatalf(
			"failure count lost after SaveCheck: %d",
			state.QuarantineFailureCount,
		)
	}
}

func TestClearQuarantinePreservesInstallAudit(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"activation failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	if err := store.ClearQuarantine(); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.QuarantinedVersion != "" {

		t.Fatalf(
			"quarantine was not cleared: %s",
			state.QuarantinedVersion,
		)
	}

	if state.LastInstallResult !=
		InstallResultFailed {

		t.Fatalf(
			"install audit was modified: %s",
			state.LastInstallResult,
		)
	}

	if !strings.Contains(
		state.LastInstallError,
		"activation failed",
	) {

		t.Fatalf(
			"install error was modified: %s",
			state.LastInstallError,
		)
	}
}

func TestQuarantineRegistryPreservesOlderFailedRelease(
	t *testing.T,
) {

	store := newQuarantineTestStore(
		t,
	)

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.2",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"v0.1.2 activation failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallStarted(
		time.Now().UTC(),
		"v0.1.1",
		"v0.1.3",
	); err != nil {

		t.Fatal(err)
	}

	if err := store.RecordInstallFailure(
		"v0.1.3 activation failed",
		true,
		true,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if len(state.QuarantinedVersions) != 2 {

		t.Fatalf(
			"expected 2 quarantined versions, got %d",
			len(state.QuarantinedVersions),
		)
	}

	oldReleaseBlocked, err :=
		store.IsVersionQuarantined(
			"v0.1.2",
		)

	if err != nil {
		t.Fatal(err)
	}

	if !oldReleaseBlocked {

		t.Fatal(
			"older failed release was forgotten",
		)
	}

	newReleaseBlocked, err :=
		store.IsVersionQuarantined(
			"v0.1.3",
		)

	if err != nil {
		t.Fatal(err)
	}

	if !newReleaseBlocked {

		t.Fatal(
			"latest failed release is not quarantined",
		)
	}
}
