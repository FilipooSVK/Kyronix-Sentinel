package updater

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

type fakeServiceController struct {
	restarts int

	failFirst bool
}

func (f *fakeServiceController) Restart(
	ctx context.Context,
) error {

	f.restarts++

	if f.failFirst &&
		f.restarts == 1 {

		return errors.New(
			"simulated restart failure",
		)
	}

	return nil
}

type fakeHealthChecker struct {
	calls []string

	failVersion string
}

func (f *fakeHealthChecker) WaitForVersion(
	ctx context.Context,
	version string,
) error {

	f.calls = append(
		f.calls,
		NormalizeVersion(version),
	)

	if NormalizeVersion(version) ==
		NormalizeVersion(f.failVersion) {

		return errors.New(
			"simulated health check failure",
		)
	}

	return nil
}

func TestActivateReleaseSuccess(
	t *testing.T,
) {

	targetDir := t.TempDir()

	releaseDir := t.TempDir()

	oldSentineld := []byte(
		"old sentineld",
	)

	oldSentinelctl := []byte(
		"old sentinelctl",
	)

	newSentineld := []byte(
		"new sentineld",
	)

	newSentinelctl := []byte(
		"new sentinelctl",
	)

	writeTestBinary(
		t,
		filepath.Join(
			targetDir,
			sentineldBinaryName,
		),
		oldSentineld,
	)

	writeTestBinary(
		t,
		filepath.Join(
			targetDir,
			sentinelctlBinaryName,
		),
		oldSentinelctl,
	)

	newSentineldPath := filepath.Join(
		releaseDir,
		sentineldBinaryName,
	)

	newSentinelctlPath := filepath.Join(
		releaseDir,
		sentinelctlBinaryName,
	)

	writeTestBinary(
		t,
		newSentineldPath,
		newSentineld,
	)

	writeTestBinary(
		t,
		newSentinelctlPath,
		newSentinelctl,
	)

	service := &fakeServiceController{}

	health := &fakeHealthChecker{}

	result, err := ActivateRelease(
		context.Background(),
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: newSentinelctlPath,
		},
		targetDir,
		"v0.1.1",
		"v0.1.0",
		service,
		health,
	)

	if err != nil {
		t.Fatal(err)
	}

	if result.RolledBack {

		t.Fatal(
			"successful activation must not rollback",
		)
	}

	if service.restarts != 1 {

		t.Fatalf(
			"expected 1 restart, got %d",
			service.restarts,
		)
	}

	assertFileContent(
		t,
		filepath.Join(
			targetDir,
			sentineldBinaryName,
		),
		newSentineld,
	)

	assertFileContent(
		t,
		filepath.Join(
			targetDir,
			sentinelctlBinaryName,
		),
		newSentinelctl,
	)
}

func TestActivateReleaseHealthFailureRollsBack(
	t *testing.T,
) {

	targetDir := t.TempDir()

	releaseDir := t.TempDir()

	oldSentineld := []byte(
		"old sentineld",
	)

	oldSentinelctl := []byte(
		"old sentinelctl",
	)

	sentineldTarget := filepath.Join(
		targetDir,
		sentineldBinaryName,
	)

	sentinelctlTarget := filepath.Join(
		targetDir,
		sentinelctlBinaryName,
	)

	writeTestBinary(
		t,
		sentineldTarget,
		oldSentineld,
	)

	writeTestBinary(
		t,
		sentinelctlTarget,
		oldSentinelctl,
	)

	newSentineldPath := filepath.Join(
		releaseDir,
		sentineldBinaryName,
	)

	newSentinelctlPath := filepath.Join(
		releaseDir,
		sentinelctlBinaryName,
	)

	writeTestBinary(
		t,
		newSentineldPath,
		[]byte("broken sentineld"),
	)

	writeTestBinary(
		t,
		newSentinelctlPath,
		[]byte("new sentinelctl"),
	)

	service := &fakeServiceController{}

	health := &fakeHealthChecker{
		failVersion: "v0.1.1",
	}

	result, err := ActivateRelease(
		context.Background(),
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: newSentinelctlPath,
		},
		targetDir,
		"v0.1.1",
		"v0.1.0",
		service,
		health,
	)

	if err == nil {

		t.Fatal(
			"expected activation health failure",
		)
	}

	if !result.RolledBack {

		t.Fatal(
			"expected rollback",
		)
	}

	if !result.RollbackVerified {

		t.Fatal(
			"expected rollback verification",
		)
	}

	if service.restarts != 2 {

		t.Fatalf(
			"expected 2 restarts, got %d",
			service.restarts,
		)
	}

	assertFileContent(
		t,
		sentineldTarget,
		oldSentineld,
	)

	assertFileContent(
		t,
		sentinelctlTarget,
		oldSentinelctl,
	)
}

func TestActivateReleaseRestartFailureRollsBack(
	t *testing.T,
) {

	targetDir := t.TempDir()

	releaseDir := t.TempDir()

	oldSentineld := []byte(
		"old sentineld",
	)

	oldSentinelctl := []byte(
		"old sentinelctl",
	)

	sentineldTarget := filepath.Join(
		targetDir,
		sentineldBinaryName,
	)

	sentinelctlTarget := filepath.Join(
		targetDir,
		sentinelctlBinaryName,
	)

	writeTestBinary(
		t,
		sentineldTarget,
		oldSentineld,
	)

	writeTestBinary(
		t,
		sentinelctlTarget,
		oldSentinelctl,
	)

	newSentineldPath := filepath.Join(
		releaseDir,
		sentineldBinaryName,
	)

	newSentinelctlPath := filepath.Join(
		releaseDir,
		sentinelctlBinaryName,
	)

	writeTestBinary(
		t,
		newSentineldPath,
		[]byte("new sentineld"),
	)

	writeTestBinary(
		t,
		newSentinelctlPath,
		[]byte("new sentinelctl"),
	)

	service := &fakeServiceController{
		failFirst: true,
	}

	health := &fakeHealthChecker{}

	result, err := ActivateRelease(
		context.Background(),
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: newSentinelctlPath,
		},
		targetDir,
		"v0.1.1",
		"v0.1.0",
		service,
		health,
	)

	if err == nil {

		t.Fatal(
			"expected restart failure",
		)
	}

	if !result.RolledBack {

		t.Fatal(
			"expected rollback",
		)
	}

	if !result.RollbackVerified {

		t.Fatal(
			"expected verified rollback",
		)
	}

	if service.restarts != 2 {

		t.Fatalf(
			"expected first failed restart plus rollback restart, got %d",
			service.restarts,
		)
	}

	assertFileContent(
		t,
		sentineldTarget,
		oldSentineld,
	)

	assertFileContent(
		t,
		sentinelctlTarget,
		oldSentinelctl,
	)
}
