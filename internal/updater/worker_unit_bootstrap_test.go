package updater

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type bootstrapTestReloader struct {
	reloads int

	failFirst bool
}

func (r *bootstrapTestReloader) DaemonReload(
	ctx context.Context,
) error {

	r.reloads++

	if r.failFirst &&
		r.reloads == 1 {

		return errors.New(
			"simulated daemon-reload failure",
		)
	}

	return nil
}

func TestEnsureUpdateWorkerUnitInstallsMissingUnit(
	t *testing.T,
) {

	target := filepath.Join(
		t.TempDir(),
		updateWorkerUnitName,
	)

	reloader := &bootstrapTestReloader{}

	installed, err := EnsureUpdateWorkerUnit(
		context.Background(),
		target,
		reloader,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !installed {

		t.Fatal(
			"expected worker unit to be installed",
		)
	}

	if reloader.reloads != 1 {

		t.Fatalf(
			"expected one daemon-reload, got %d",
			reloader.reloads,
		)
	}

	data, err := os.ReadFile(
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(data) !=
		updateWorkerUnitContent {

		t.Fatal(
			"installed worker unit content mismatch",
		)
	}

	info, err := os.Stat(
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if info.Mode().Perm() != 0644 {

		t.Fatalf(
			"expected mode 0644, got %o",
			info.Mode().Perm(),
		)
	}
}

func TestEnsureUpdateWorkerUnitLeavesExistingUnitUntouched(
	t *testing.T,
) {

	target := filepath.Join(
		t.TempDir(),
		updateWorkerUnitName,
	)

	if err := os.WriteFile(
		target,
		[]byte("existing unit"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	reloader := &bootstrapTestReloader{}

	installed, err := EnsureUpdateWorkerUnit(
		context.Background(),
		target,
		reloader,
	)

	if err != nil {
		t.Fatal(err)
	}

	if installed {

		t.Fatal(
			"existing worker unit must not be replaced",
		)
	}

	if reloader.reloads != 0 {

		t.Fatalf(
			"unexpected daemon-reload count: %d",
			reloader.reloads,
		)
	}

	data, err := os.ReadFile(
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "existing unit" {

		t.Fatal(
			"existing worker unit was modified",
		)
	}
}

func TestEnsureUpdateWorkerUnitReloadFailureRollsBack(
	t *testing.T,
) {

	target := filepath.Join(
		t.TempDir(),
		updateWorkerUnitName,
	)

	reloader := &bootstrapTestReloader{
		failFirst: true,
	}

	installed, err := EnsureUpdateWorkerUnit(
		context.Background(),
		target,
		reloader,
	)

	if err == nil {

		t.Fatal(
			"expected daemon-reload failure",
		)
	}

	if installed {

		t.Fatal(
			"failed bootstrap must not report installation",
		)
	}

	if reloader.reloads != 2 {

		t.Fatalf(
			"expected failed reload plus rollback reload, got %d",
			reloader.reloads,
		)
	}

	if _, err := os.Stat(
		target,
	); !os.IsNotExist(err) {

		t.Fatalf(
			"expected bootstrap worker unit removed after rollback, got %v",
			err,
		)
	}
}

func TestUpdateWorkerUnitContentMatchesSystemdSource(
	t *testing.T,
) {

	path := filepath.Join(
		"..",
		"..",
		"systemd",
		updateWorkerUnitName,
	)

	data, err := os.ReadFile(
		path,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(data) !=
		updateWorkerUnitContent {

		t.Fatal(
			"embedded worker unit differs from systemd source file",
		)
	}
}
