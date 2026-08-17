package updater

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallReleaseCreatesBackups(
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

	result, err := InstallRelease(
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: newSentinelctlPath,
		},
		targetDir,
	)

	if err != nil {
		t.Fatal(err)
	}

	assertFileContent(
		t,
		result.SentineldPath,
		newSentineld,
	)

	assertFileContent(
		t,
		result.SentinelctlPath,
		newSentinelctl,
	)

	assertFileContent(
		t,
		result.SentineldBackupPath,
		oldSentineld,
	)

	assertFileContent(
		t,
		result.SentinelctlBackupPath,
		oldSentinelctl,
	)

	for _, path := range []string{
		result.SentineldPath,
		result.SentinelctlPath,
		result.SentineldBackupPath,
		result.SentinelctlBackupPath,
	} {

		info, err := os.Stat(
			path,
		)

		if err != nil {
			t.Fatal(err)
		}

		if info.Mode().Perm() != 0755 {

			t.Fatalf(
				"expected %s mode 0755, got %o",
				path,
				info.Mode().Perm(),
			)
		}
	}
}

func TestRestorePrevious(
	t *testing.T,
) {

	targetDir := t.TempDir()

	releaseDir := t.TempDir()

	oldSentineld := []byte(
		"previous sentineld",
	)

	oldSentinelctl := []byte(
		"previous sentinelctl",
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
		[]byte("new sentineld"),
	)

	writeTestBinary(
		t,
		newSentinelctlPath,
		[]byte("new sentinelctl"),
	)

	_, err := InstallRelease(
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: newSentinelctlPath,
		},
		targetDir,
	)

	if err != nil {
		t.Fatal(err)
	}

	if err := RestorePrevious(
		targetDir,
	); err != nil {

		t.Fatal(err)
	}

	assertFileContent(
		t,
		filepath.Join(
			targetDir,
			sentineldBinaryName,
		),
		oldSentineld,
	)

	assertFileContent(
		t,
		filepath.Join(
			targetDir,
			sentinelctlBinaryName,
		),
		oldSentinelctl,
	)
}

func TestInstallReleaseMissingNewBinaryDoesNotModifyCurrent(
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

	writeTestBinary(
		t,
		newSentineldPath,
		[]byte("new sentineld"),
	)

	_, err := InstallRelease(
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: filepath.Join(
				releaseDir,
				sentinelctlBinaryName,
			),
		},
		targetDir,
	)

	if err == nil {

		t.Fatal(
			"expected missing source binary error",
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

func TestInstallReleaseRequiresExistingInstallation(
	t *testing.T,
) {

	targetDir := t.TempDir()

	releaseDir := t.TempDir()

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

	_, err := InstallRelease(
		ExtractedRelease{
			SentineldPath: newSentineldPath,

			SentinelctlPath: newSentinelctlPath,
		},
		targetDir,
	)

	if err == nil {

		t.Fatal(
			"expected missing current installation error",
		)
	}
}

func writeTestBinary(
	t *testing.T,
	path string,
	data []byte,
) {

	t.Helper()

	if err := os.WriteFile(
		path,
		data,
		0755,
	); err != nil {

		t.Fatal(err)
	}
}

func assertFileContent(
	t *testing.T,
	path string,
	expected []byte,
) {

	t.Helper()

	actual, err := os.ReadFile(
		path,
	)

	if err != nil {
		t.Fatal(err)
	}

	if string(actual) != string(expected) {

		t.Fatalf(
			"unexpected content in %s: %q",
			path,
			string(actual),
		)
	}
}
