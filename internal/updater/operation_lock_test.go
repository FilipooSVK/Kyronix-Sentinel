package updater

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestOperationLockPath(
	t *testing.T,
) {

	got := OperationLockPath(
		"/var/lib/sentinel/update-state.json",
	)

	want :=
		"/var/lib/sentinel/update-state.json.operation.lock"

	if got != want {

		t.Fatalf(
			"unexpected operation lock path: got %q want %q",
			got,
			want,
		)
	}
}

func TestOperationLockPathRejectsEmptyStatePath(
	t *testing.T,
) {

	if got := OperationLockPath(
		"   ",
	); got != "" {

		t.Fatalf(
			"expected empty lock path, got %q",
			got,
		)
	}
}

func TestOperationLockPreventsConcurrentAcquire(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update.operation.lock",
	)

	first := NewOperationLock(
		path,
	)

	if err := first.Acquire(); err != nil {

		t.Fatalf(
			"first acquire failed: %v",
			err,
		)
	}

	defer first.Release()

	second := NewOperationLock(
		path,
	)

	err := second.Acquire()

	if !errors.Is(
		err,
		ErrUpdateOperationLocked,
	) {

		t.Fatalf(
			"expected ErrUpdateOperationLocked, got %v",
			err,
		)
	}
}

func TestOperationLockCanBeAcquiredAfterRelease(
	t *testing.T,
) {

	path := filepath.Join(
		t.TempDir(),
		"update.operation.lock",
	)

	first := NewOperationLock(
		path,
	)

	if err := first.Acquire(); err != nil {
		t.Fatal(err)
	}

	if err := first.Release(); err != nil {
		t.Fatal(err)
	}

	second := NewOperationLock(
		path,
	)

	if err := second.Acquire(); err != nil {

		t.Fatalf(
			"second acquire failed after release: %v",
			err,
		)
	}

	if err := second.Release(); err != nil {
		t.Fatal(err)
	}
}

func TestOperationLockReleaseIsIdempotent(
	t *testing.T,
) {

	lock := NewOperationLock(
		filepath.Join(
			t.TempDir(),
			"update.operation.lock",
		),
	)

	if err := lock.Acquire(); err != nil {
		t.Fatal(err)
	}

	if err := lock.Release(); err != nil {
		t.Fatal(err)
	}

	if err := lock.Release(); err != nil {

		t.Fatalf(
			"second release failed: %v",
			err,
		)
	}
}

func TestOperationLockRejectsEmptyPath(
	t *testing.T,
) {

	lock := NewOperationLock(
		"",
	)

	if err := lock.Acquire(); err == nil {

		t.Fatal(
			"expected empty path error",
		)
	}
}
