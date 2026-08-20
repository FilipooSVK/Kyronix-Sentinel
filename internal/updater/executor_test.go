package updater

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

type executorTestService struct{}

func (s *executorTestService) Restart(
	ctx context.Context,
) error {
	return nil
}

type executorTestReloader struct{}

func (r *executorTestReloader) DaemonReload(
	ctx context.Context,
) error {
	return nil
}

type executorTestHealth struct{}

func (h *executorTestHealth) WaitForVersion(
	ctx context.Context,
	version string,
) error {
	return nil
}

func newExecutorTestConfig(
	t *testing.T,
) InstallExecutorConfig {

	t.Helper()

	return InstallExecutorConfig{
		CurrentVersion: "0.1.2",

		CheckInterval: 24 * time.Hour,

		StatePath: filepath.Join(
			t.TempDir(),
			"update-state.json",
		),

		InstallDir: filepath.Join(
			t.TempDir(),
			"bin",
		),

		WorkerUnitTarget: filepath.Join(
			t.TempDir(),
			"systemd",
			updateWorkerUnitName,
		),
	}
}

func TestInstallExecutorUpToDate(
	t *testing.T,
) {

	cfg := newExecutorTestConfig(
		t,
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			return CheckResult{
				CurrentVersion: "v0.1.2",

				LatestVersion: "v0.1.2",

				UpdateAvailable: false,
			}, nil
		},
	}

	executor := NewInstallExecutor(
		checker,
		cfg,
		&executorTestService{},
		&executorTestReloader{},
		&executorTestHealth{},
	)

	result, err := executor.Execute(
		context.Background(),
		nil,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !result.UpToDate {

		t.Fatal(
			"expected up-to-date result",
		)
	}

	if result.Stage !=
		ExecutorStageComplete {

		t.Fatalf(
			"unexpected stage: %s",
			result.Stage,
		)
	}

	state, err :=
		NewStateStore(
			cfg.StatePath,
		).Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.CurrentVersion !=
		"v0.1.2" {

		t.Fatalf(
			"unexpected current version: %s",
			state.CurrentVersion,
		)
	}

	if state.UpdateAvailable {

		t.Fatal(
			"unexpected update available state",
		)
	}
}

func TestInstallExecutorHonorsOperationLock(
	t *testing.T,
) {

	cfg := newExecutorTestConfig(
		t,
	)

	lock := NewOperationLock(
		OperationLockPath(
			cfg.StatePath,
		),
	)

	if err := lock.Acquire(); err != nil {
		t.Fatal(err)
	}

	defer lock.Release()

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			t.Fatal(
				"checker must not run while operation lock is held",
			)

			return CheckResult{}, nil
		},
	}

	executor := NewInstallExecutor(
		checker,
		cfg,
		&executorTestService{},
		&executorTestReloader{},
		&executorTestHealth{},
	)

	result, err := executor.Execute(
		context.Background(),
		nil,
	)

	if !errors.Is(
		err,
		ErrUpdateOperationLocked,
	) {

		t.Fatalf(
			"expected operation lock error, got %v",
			err,
		)
	}

	if result.Stage !=
		ExecutorStageLock {

		t.Fatalf(
			"unexpected failure stage: %s",
			result.Stage,
		)
	}
}

func TestInstallExecutorSuccessfulPipeline(
	t *testing.T,
) {

	cfg := newExecutorTestConfig(
		t,
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			return CheckResult{
				CurrentVersion: "v0.1.2",

				LatestVersion: "v0.1.3",

				UpdateAvailable: true,

				Release: Release{
					TagName: "v0.1.3",
				},
			}, nil
		},
	}

	executor := NewInstallExecutor(
		checker,
		cfg,
		&executorTestService{},
		&executorTestReloader{},
		&executorTestHealth{},
	)

	executor.selectAssets =
		func(
			release Release,
		) (ReleaseAssets, error) {

			return ReleaseAssets{
				Package: Asset{
					Name: "sentinel-v0.1.3-linux-arm64.tar.gz",
				},

				Checksum: Asset{
					Name: "sentinel-v0.1.3-linux-arm64.tar.gz.sha256",
				},

				OS: "linux",

				Arch: "arm64",
			}, nil
		}

	executor.download =
		func(
			ctx context.Context,
			assets ReleaseAssets,
			destination string,
		) (VerifiedArtifact, error) {

			return VerifiedArtifact{
				PackagePath: filepath.Join(
					destination,
					"release.tar.gz",
				),

				ChecksumPath: filepath.Join(
					destination,
					"release.sha256",
				),

				SHA256: "test-sha256",
			}, nil
		}

	executor.extract =
		func(
			artifact VerifiedArtifact,
			destination string,
			expectedVersion string,
			expectedOS string,
			expectedArch string,
		) (ExtractedRelease, error) {

			return ExtractedRelease{},
				nil
		}

	executor.activate =
		func(
			ctx context.Context,
			release ExtractedRelease,
			targetDir string,
			expectedVersion string,
			previousVersion string,
			service ServiceController,
			health VersionHealthChecker,
			options ...ActivationOptions,
		) (ActivationResult, error) {

			return ActivationResult{
				InstalledVersion: "v0.1.3",
			}, nil
		}

	var events []ExecutorEventType

	result, err := executor.Execute(
		context.Background(),
		func(
			event ExecutorEvent,
		) {

			events = append(
				events,
				event.Type,
			)
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.Stage !=
		ExecutorStageComplete {

		t.Fatalf(
			"unexpected final stage: %s",
			result.Stage,
		)
	}

	if result.Activation.InstalledVersion !=
		"v0.1.3" {

		t.Fatalf(
			"unexpected installed version: %s",
			result.Activation.InstalledVersion,
		)
	}

	if len(events) == 0 {

		t.Fatal(
			"expected executor events",
		)
	}

	state, err :=
		NewStateStore(
			cfg.StatePath,
		).Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.LastInstallResult !=
		InstallResultSuccess {

		t.Fatalf(
			"unexpected lifecycle result: %s",
			state.LastInstallResult,
		)
	}

	if state.LastInstalledVersion !=
		"v0.1.3" {

		t.Fatalf(
			"unexpected installed version in state: %s",
			state.LastInstalledVersion,
		)
	}
}

func TestInstallExecutorRecordsPipelineFailure(
	t *testing.T,
) {

	cfg := newExecutorTestConfig(
		t,
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			return CheckResult{
				CurrentVersion: "v0.1.2",

				LatestVersion: "v0.1.3",

				UpdateAvailable: true,

				Release: Release{
					TagName: "v0.1.3",
				},
			}, nil
		},
	}

	executor := NewInstallExecutor(
		checker,
		cfg,
		&executorTestService{},
		&executorTestReloader{},
		&executorTestHealth{},
	)

	expectedErr :=
		errors.New(
			"asset selection failed",
		)

	executor.selectAssets =
		func(
			release Release,
		) (ReleaseAssets, error) {

			return ReleaseAssets{},
				expectedErr
		}

	result, err := executor.Execute(
		context.Background(),
		nil,
	)

	if !errors.Is(
		err,
		expectedErr,
	) {

		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if result.Stage !=
		ExecutorStageAssets {

		t.Fatalf(
			"unexpected failure stage: %s",
			result.Stage,
		)
	}

	state, loadErr :=
		NewStateStore(
			cfg.StatePath,
		).Load()

	if loadErr != nil {
		t.Fatal(loadErr)
	}

	if state.LastInstallResult !=
		InstallResultFailed {

		t.Fatalf(
			"unexpected lifecycle result: %s",
			state.LastInstallResult,
		)
	}

	if state.LastInstallError !=
		expectedErr.Error() {

		t.Fatalf(
			"unexpected lifecycle error: %s",
			state.LastInstallError,
		)
	}
}
