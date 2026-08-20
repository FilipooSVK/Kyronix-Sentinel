package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallWorkerUnitFirstInstallAndRollback(
	t *testing.T,
) {

	root := t.TempDir()

	sourceDir := filepath.Join(
		root,
		"release",
	)

	targetDir := filepath.Join(
		root,
		"systemd",
	)

	if err := os.MkdirAll(
		sourceDir,
		0755,
	); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(
		sourceDir,
		updateWorkerUnitName,
	)

	target := filepath.Join(
		targetDir,
		updateWorkerUnitName,
	)

	if err := os.WriteFile(
		source,
		[]byte("new worker"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	result, err := InstallWorkerUnit(
		source,
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.PreviousExisted {

		t.Fatal(
			"first install unexpectedly reported previous unit",
		)
	}

	data, err := os.ReadFile(
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "new worker" {

		t.Fatalf(
			"unexpected installed unit: %q",
			string(data),
		)
	}

	if err := RestoreWorkerUnit(
		result,
	); err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(
		target,
	); !os.IsNotExist(err) {

		t.Fatalf(
			"expected worker unit to be removed after rollback, got %v",
			err,
		)
	}
}

func TestInstallWorkerUnitUpgradeAndRollback(
	t *testing.T,
) {

	root := t.TempDir()

	sourceDir := filepath.Join(
		root,
		"release",
	)

	targetDir := filepath.Join(
		root,
		"systemd",
	)

	if err := os.MkdirAll(
		sourceDir,
		0755,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.MkdirAll(
		targetDir,
		0755,
	); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(
		sourceDir,
		updateWorkerUnitName,
	)

	target := filepath.Join(
		targetDir,
		updateWorkerUnitName,
	)

	if err := os.WriteFile(
		source,
		[]byte("new worker"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(
		target,
		[]byte("old worker"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	result, err := InstallWorkerUnit(
		source,
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !result.PreviousExisted {

		t.Fatal(
			"upgrade did not report previous worker unit",
		)
	}

	backup, err := os.ReadFile(
		result.BackupPath,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(backup) != "old worker" {

		t.Fatalf(
			"unexpected backup contents: %q",
			string(backup),
		)
	}

	if err := RestoreWorkerUnit(
		result,
	); err != nil {
		t.Fatal(err)
	}

	restored, err := os.ReadFile(
		target,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(restored) != "old worker" {

		t.Fatalf(
			"unexpected restored unit: %q",
			string(restored),
		)
	}
}

func TestInstallWorkerUnitUses0644(
	t *testing.T,
) {

	root := t.TempDir()

	sourceDir := filepath.Join(
		root,
		"release",
	)

	if err := os.MkdirAll(
		sourceDir,
		0755,
	); err != nil {
		t.Fatal(err)
	}

	source := filepath.Join(
		sourceDir,
		updateWorkerUnitName,
	)

	target := filepath.Join(
		root,
		"systemd",
		updateWorkerUnitName,
	)

	if err := os.WriteFile(
		source,
		[]byte("worker"),
		0755,
	); err != nil {
		t.Fatal(err)
	}

	_, err := InstallWorkerUnit(
		source,
		target,
	)

	if err != nil {
		t.Fatal(err)
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
