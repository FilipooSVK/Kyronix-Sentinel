package updater

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type fakeSystemdReloader struct {
	reloads int

	failFirst bool
}

func (f *fakeSystemdReloader) DaemonReload(
	ctx context.Context,
) error {

	f.reloads++

	if f.failFirst &&
		f.reloads == 1 {

		return errors.New(
			"simulated daemon-reload failure",
		)
	}

	return nil
}

func prepareWorkerActivationTest(
	t *testing.T,
) (
	string,
	ExtractedRelease,
	string,
) {

	t.Helper()

	targetDir := t.TempDir()

	releaseDir := t.TempDir()

	writeTestBinary(
		t,
		filepath.Join(
			targetDir,
			sentineldBinaryName,
		),
		[]byte("old sentineld"),
	)

	writeTestBinary(
		t,
		filepath.Join(
			targetDir,
			sentinelctlBinaryName,
		),
		[]byte("old sentinelctl"),
	)

	newSentineld := filepath.Join(
		releaseDir,
		sentineldBinaryName,
	)

	newSentinelctl := filepath.Join(
		releaseDir,
		sentinelctlBinaryName,
	)

	workerSource := filepath.Join(
		releaseDir,
		updateWorkerUnitName,
	)

	writeTestBinary(
		t,
		newSentineld,
		[]byte("new sentineld"),
	)

	writeTestBinary(
		t,
		newSentinelctl,
		[]byte("new sentinelctl"),
	)

	if err := os.WriteFile(
		workerSource,
		[]byte("new worker unit"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	workerTarget := filepath.Join(
		t.TempDir(),
		updateWorkerUnitName,
	)

	return targetDir,
		ExtractedRelease{
			SentineldPath: newSentineld,

			SentinelctlPath: newSentinelctl,

			WorkerUnitPath: workerSource,
		},
		workerTarget
}

func TestActivateReleaseInstallsWorkerUnit(
	t *testing.T,
) {

	targetDir, release, workerTarget :=
		prepareWorkerActivationTest(
			t,
		)

	service := &fakeServiceController{}

	health := &fakeHealthChecker{}

	reloader := &fakeSystemdReloader{}

	result, err := ActivateRelease(
		context.Background(),
		release,
		targetDir,
		"v0.1.4",
		"v0.1.3",
		service,
		health,
		ActivationOptions{
			WorkerUnitTarget: workerTarget,

			Reloader: reloader,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.WorkerUnit == nil {

		t.Fatal(
			"expected worker unit transaction result",
		)
	}

	if reloader.reloads != 1 {

		t.Fatalf(
			"expected one daemon-reload, got %d",
			reloader.reloads,
		)
	}

	assertFileContent(
		t,
		workerTarget,
		[]byte("new worker unit"),
	)
}

func TestActivateReleaseHealthFailureRemovesFirstWorkerUnit(
	t *testing.T,
) {

	targetDir, release, workerTarget :=
		prepareWorkerActivationTest(
			t,
		)

	service := &fakeServiceController{}

	health := &fakeHealthChecker{
		failVersion: "v0.1.4",
	}

	reloader := &fakeSystemdReloader{}

	result, err := ActivateRelease(
		context.Background(),
		release,
		targetDir,
		"v0.1.4",
		"v0.1.3",
		service,
		health,
		ActivationOptions{
			WorkerUnitTarget: workerTarget,

			Reloader: reloader,
		},
	)

	if err == nil {

		t.Fatal(
			"expected activation failure",
		)
	}

	if !result.RolledBack ||
		!result.RollbackVerified {

		t.Fatal(
			"expected verified rollback",
		)
	}

	if reloader.reloads != 2 {

		t.Fatalf(
			"expected install and rollback daemon-reload, got %d",
			reloader.reloads,
		)
	}

	if _, err := os.Stat(
		workerTarget,
	); !os.IsNotExist(err) {

		t.Fatalf(
			"expected first-installed worker unit to be removed, got %v",
			err,
		)
	}

	assertFileContent(
		t,
		filepath.Join(
			targetDir,
			sentineldBinaryName,
		),
		[]byte("old sentineld"),
	)

	assertFileContent(
		t,
		filepath.Join(
			targetDir,
			sentinelctlBinaryName,
		),
		[]byte("old sentinelctl"),
	)
}

func TestActivateReleaseHealthFailureRestoresPreviousWorkerUnit(
	t *testing.T,
) {

	targetDir, release, workerTarget :=
		prepareWorkerActivationTest(
			t,
		)

	if err := os.WriteFile(
		workerTarget,
		[]byte("old worker unit"),
		0644,
	); err != nil {
		t.Fatal(err)
	}

	service := &fakeServiceController{}

	health := &fakeHealthChecker{
		failVersion: "v0.1.4",
	}

	reloader := &fakeSystemdReloader{}

	result, err := ActivateRelease(
		context.Background(),
		release,
		targetDir,
		"v0.1.4",
		"v0.1.3",
		service,
		health,
		ActivationOptions{
			WorkerUnitTarget: workerTarget,

			Reloader: reloader,
		},
	)

	if err == nil {
		t.Fatal("expected activation failure")
	}

	if !result.RollbackVerified {

		t.Fatal(
			"expected verified rollback",
		)
	}

	assertFileContent(
		t,
		workerTarget,
		[]byte("old worker unit"),
	)
}

func TestActivateReleaseDaemonReloadFailureRollsBack(
	t *testing.T,
) {

	targetDir, release, workerTarget :=
		prepareWorkerActivationTest(
			t,
		)

	service := &fakeServiceController{}

	health := &fakeHealthChecker{}

	reloader := &fakeSystemdReloader{
		failFirst: true,
	}

	result, err := ActivateRelease(
		context.Background(),
		release,
		targetDir,
		"v0.1.4",
		"v0.1.3",
		service,
		health,
		ActivationOptions{
			WorkerUnitTarget: workerTarget,

			Reloader: reloader,
		},
	)

	if err == nil {

		t.Fatal(
			"expected daemon-reload failure",
		)
	}

	if !result.RolledBack ||
		!result.RollbackVerified {

		t.Fatal(
			"expected verified rollback",
		)
	}

	if reloader.reloads != 2 {

		t.Fatalf(
			"expected failed reload plus rollback reload, got %d",
			reloader.reloads,
		)
	}

	if _, err := os.Stat(
		workerTarget,
	); !os.IsNotExist(err) {

		t.Fatalf(
			"expected worker unit removed after rollback, got %v",
			err,
		)
	}
}
