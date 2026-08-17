package updater

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	sentineldBinaryName   = "sentineld"
	sentinelctlBinaryName = "sentinelctl"

	previousSuffix = ".previous"
)

// InstallResult describes paths affected by an update installation.
type InstallResult struct {
	SentineldPath string

	SentinelctlPath string

	SentineldBackupPath string

	SentinelctlBackupPath string
}

// InstallRelease atomically installs validated Sentinel binaries.
//
// Both new binaries are staged before the current installation is
// modified. Existing binaries are copied to .previous backups.
//
// This function does not restart any service.
func InstallRelease(
	release ExtractedRelease,
	targetDir string,
) (InstallResult, error) {

	if targetDir == "" {

		return InstallResult{}, fmt.Errorf(
			"install target directory is empty",
		)
	}

	if err := validateInstallSource(
		release.SentineldPath,
		sentineldBinaryName,
	); err != nil {

		return InstallResult{}, err
	}

	if err := validateInstallSource(
		release.SentinelctlPath,
		sentinelctlBinaryName,
	); err != nil {

		return InstallResult{}, err
	}

	sentineldTarget := filepath.Join(
		targetDir,
		sentineldBinaryName,
	)

	sentinelctlTarget := filepath.Join(
		targetDir,
		sentinelctlBinaryName,
	)

	sentineldBackup := sentineldTarget +
		previousSuffix

	sentinelctlBackup := sentinelctlTarget +
		previousSuffix

	if err := validateInstalledBinary(
		sentineldTarget,
	); err != nil {

		return InstallResult{}, fmt.Errorf(
			"current sentineld validation failed: %w",
			err,
		)
	}

	if err := validateInstalledBinary(
		sentinelctlTarget,
	); err != nil {

		return InstallResult{}, fmt.Errorf(
			"current sentinelctl validation failed: %w",
			err,
		)
	}

	// Stage both new binaries before changing the current installation.
	stagedSentineld, err := stageBinary(
		release.SentineldPath,
		targetDir,
		sentineldBinaryName,
	)

	if err != nil {

		return InstallResult{}, fmt.Errorf(
			"failed to stage sentineld: %w",
			err,
		)
	}

	defer os.Remove(
		stagedSentineld,
	)

	stagedSentinelctl, err := stageBinary(
		release.SentinelctlPath,
		targetDir,
		sentinelctlBinaryName,
	)

	if err != nil {

		return InstallResult{}, fmt.Errorf(
			"failed to stage sentinelctl: %w",
			err,
		)
	}

	defer os.Remove(
		stagedSentinelctl,
	)

	// Create durable backups before replacing anything.
	if err := copyFileAtomic(
		sentineldTarget,
		sentineldBackup,
		0755,
	); err != nil {

		return InstallResult{}, fmt.Errorf(
			"failed to backup sentineld: %w",
			err,
		)
	}

	if err := copyFileAtomic(
		sentinelctlTarget,
		sentinelctlBackup,
		0755,
	); err != nil {

		return InstallResult{}, fmt.Errorf(
			"failed to backup sentinelctl: %w",
			err,
		)
	}

	if err := os.Rename(
		stagedSentineld,
		sentineldTarget,
	); err != nil {

		return InstallResult{}, fmt.Errorf(
			"failed to install sentineld: %w",
			err,
		)
	}

	if err := os.Rename(
		stagedSentinelctl,
		sentinelctlTarget,
	); err != nil {

		rollbackErr := RestorePrevious(
			targetDir,
		)

		if rollbackErr != nil {

			return InstallResult{}, fmt.Errorf(
				"failed to install sentinelctl: %v; rollback failed: %v",
				err,
				rollbackErr,
			)
		}

		return InstallResult{}, fmt.Errorf(
			"failed to install sentinelctl: %w; previous release restored",
			err,
		)
	}

	if err := syncDirectory(
		targetDir,
	); err != nil {

		rollbackErr := RestorePrevious(
			targetDir,
		)

		if rollbackErr != nil {

			return InstallResult{}, fmt.Errorf(
				"install directory sync failed: %v; rollback failed: %v",
				err,
				rollbackErr,
			)
		}

		return InstallResult{}, fmt.Errorf(
			"install directory sync failed: %w; previous release restored",
			err,
		)
	}

	return InstallResult{
		SentineldPath: sentineldTarget,

		SentinelctlPath: sentinelctlTarget,

		SentineldBackupPath: sentineldBackup,

		SentinelctlBackupPath: sentinelctlBackup,
	}, nil
}

// RestorePrevious restores both binaries from .previous backups.
//
// Restoration itself is also atomic for each binary.
func RestorePrevious(
	targetDir string,
) error {

	sentineldTarget := filepath.Join(
		targetDir,
		sentineldBinaryName,
	)

	sentinelctlTarget := filepath.Join(
		targetDir,
		sentinelctlBinaryName,
	)

	sentineldBackup := sentineldTarget +
		previousSuffix

	sentinelctlBackup := sentinelctlTarget +
		previousSuffix

	if err := validateInstalledBinary(
		sentineldBackup,
	); err != nil {

		return fmt.Errorf(
			"sentineld backup validation failed: %w",
			err,
		)
	}

	if err := validateInstalledBinary(
		sentinelctlBackup,
	); err != nil {

		return fmt.Errorf(
			"sentinelctl backup validation failed: %w",
			err,
		)
	}

	if err := copyFileAtomic(
		sentineldBackup,
		sentineldTarget,
		0755,
	); err != nil {

		return fmt.Errorf(
			"failed to restore sentineld: %w",
			err,
		)
	}

	if err := copyFileAtomic(
		sentinelctlBackup,
		sentinelctlTarget,
		0755,
	); err != nil {

		return fmt.Errorf(
			"failed to restore sentinelctl: %w",
			err,
		)
	}

	if err := syncDirectory(
		targetDir,
	); err != nil {

		return fmt.Errorf(
			"failed to sync restored installation: %w",
			err,
		)
	}

	return nil
}

func validateInstallSource(
	path string,
	expectedName string,
) error {

	if path == "" {

		return fmt.Errorf(
			"%s source path is empty",
			expectedName,
		)
	}

	info, err := os.Lstat(
		path,
	)

	if err != nil {

		return fmt.Errorf(
			"%s source unavailable: %w",
			expectedName,
			err,
		)
	}

	if !info.Mode().IsRegular() {

		return fmt.Errorf(
			"%s source is not a regular file",
			expectedName,
		)
	}

	return nil
}

func validateInstalledBinary(
	path string,
) error {

	info, err := os.Lstat(
		path,
	)

	if err != nil {
		return err
	}

	if !info.Mode().IsRegular() {

		return fmt.Errorf(
			"not a regular file: %s",
			path,
		)
	}

	return nil
}

func stageBinary(
	source string,
	targetDir string,
	name string,
) (string, error) {

	tempFile, err := os.CreateTemp(
		targetDir,
		"."+name+"-*.new",
	)

	if err != nil {
		return "", err
	}

	tempPath := tempFile.Name()

	cleanup := func() {

		tempFile.Close()

		os.Remove(
			tempPath,
		)
	}

	sourceFile, err := os.Open(
		source,
	)

	if err != nil {

		cleanup()

		return "", err
	}

	_, copyErr := io.Copy(
		tempFile,
		sourceFile,
	)

	closeSourceErr := sourceFile.Close()

	if copyErr != nil {

		cleanup()

		return "", copyErr
	}

	if closeSourceErr != nil {

		cleanup()

		return "", closeSourceErr
	}

	if err := tempFile.Chmod(
		0755,
	); err != nil {

		cleanup()

		return "", err
	}

	if err := tempFile.Sync(); err != nil {

		cleanup()

		return "", err
	}

	if err := tempFile.Close(); err != nil {

		os.Remove(
			tempPath,
		)

		return "", err
	}

	return tempPath, nil
}

func copyFileAtomic(
	source string,
	target string,
	mode os.FileMode,
) error {

	dir := filepath.Dir(
		target,
	)

	tempFile, err := os.CreateTemp(
		dir,
		"."+filepath.Base(target)+"-*.tmp",
	)

	if err != nil {
		return err
	}

	tempPath := tempFile.Name()

	cleanup := func() {

		tempFile.Close()

		os.Remove(
			tempPath,
		)
	}

	sourceFile, err := os.Open(
		source,
	)

	if err != nil {

		cleanup()

		return err
	}

	_, copyErr := io.Copy(
		tempFile,
		sourceFile,
	)

	closeSourceErr := sourceFile.Close()

	if copyErr != nil {

		cleanup()

		return copyErr
	}

	if closeSourceErr != nil {

		cleanup()

		return closeSourceErr
	}

	if err := tempFile.Chmod(
		mode,
	); err != nil {

		cleanup()

		return err
	}

	if err := tempFile.Sync(); err != nil {

		cleanup()

		return err
	}

	if err := tempFile.Close(); err != nil {

		os.Remove(
			tempPath,
		)

		return err
	}

	if err := os.Rename(
		tempPath,
		target,
	); err != nil {

		os.Remove(
			tempPath,
		)

		return err
	}

	return nil
}

func syncDirectory(
	path string,
) error {

	dir, err := os.Open(
		path,
	)

	if err != nil {
		return err
	}

	defer dir.Close()

	return dir.Sync()
}
