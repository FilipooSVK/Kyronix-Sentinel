package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	updateWorkerUnitName = "sentinel-update.service"

	unitPreviousSuffix = ".previous"
)

// UnitInstallResult describes the worker unit installation transaction.
type UnitInstallResult struct {
	TargetPath string

	BackupPath string

	PreviousExisted bool
}

// InstallWorkerUnit atomically installs the validated Sentinel update
// worker systemd unit.
//
// If an older unit already exists it is preserved as .previous.
// If the unit did not previously exist, rollback will remove the newly
// installed unit instead.
func InstallWorkerUnit(
	sourcePath string,
	targetPath string,
) (UnitInstallResult, error) {

	if sourcePath == "" {

		return UnitInstallResult{}, fmt.Errorf(
			"worker unit source path is empty",
		)
	}

	if targetPath == "" {

		return UnitInstallResult{}, fmt.Errorf(
			"worker unit target path is empty",
		)
	}

	if filepath.Base(
		sourcePath,
	) != updateWorkerUnitName {

		return UnitInstallResult{}, fmt.Errorf(
			"unexpected worker unit source name: %s",
			filepath.Base(sourcePath),
		)
	}

	if filepath.Base(
		targetPath,
	) != updateWorkerUnitName {

		return UnitInstallResult{}, fmt.Errorf(
			"unexpected worker unit target name: %s",
			filepath.Base(targetPath),
		)
	}

	sourceInfo, err := os.Lstat(
		sourcePath,
	)

	if err != nil {

		return UnitInstallResult{}, fmt.Errorf(
			"worker unit source unavailable: %w",
			err,
		)
	}

	if !sourceInfo.Mode().IsRegular() {

		return UnitInstallResult{}, fmt.Errorf(
			"worker unit source is not a regular file",
		)
	}

	targetDir := filepath.Dir(
		targetPath,
	)

	if err := os.MkdirAll(
		targetDir,
		0755,
	); err != nil {

		return UnitInstallResult{}, fmt.Errorf(
			"create systemd unit directory: %w",
			err,
		)
	}

	result := UnitInstallResult{
		TargetPath: targetPath,

		BackupPath: targetPath +
			unitPreviousSuffix,
	}

	targetInfo, err := os.Lstat(
		targetPath,
	)

	switch {

	case err == nil:

		if !targetInfo.Mode().IsRegular() {

			return UnitInstallResult{}, fmt.Errorf(
				"existing worker unit is not a regular file",
			)
		}

		result.PreviousExisted = true

		if err := copyFileAtomic(
			targetPath,
			result.BackupPath,
			0644,
		); err != nil {

			return UnitInstallResult{}, fmt.Errorf(
				"backup worker unit: %w",
				err,
			)
		}

	case os.IsNotExist(err):

		// A stale backup from an older transaction must not make a
		// first-time install look as though a previous unit existed.
		if err := os.Remove(
			result.BackupPath,
		); err != nil &&
			!os.IsNotExist(err) {

			return UnitInstallResult{}, fmt.Errorf(
				"remove stale worker unit backup: %w",
				err,
			)
		}

	default:

		return UnitInstallResult{}, fmt.Errorf(
			"inspect existing worker unit: %w",
			err,
		)
	}

	if err := copyFileAtomic(
		sourcePath,
		targetPath,
		0644,
	); err != nil {

		return UnitInstallResult{}, fmt.Errorf(
			"install worker unit: %w",
			err,
		)
	}

	if err := syncDirectory(
		targetDir,
	); err != nil {

		rollbackErr := RestoreWorkerUnit(
			result,
		)

		if rollbackErr != nil {

			return UnitInstallResult{}, fmt.Errorf(
				"sync worker unit directory failed: %v; rollback failed: %v",
				err,
				rollbackErr,
			)
		}

		return UnitInstallResult{}, fmt.Errorf(
			"sync worker unit directory failed: %w; previous unit restored",
			err,
		)
	}

	return result, nil
}

// RestoreWorkerUnit restores the worker unit state from before the
// transaction.
//
// When no previous worker unit existed, the newly installed unit is
// removed.
func RestoreWorkerUnit(
	result UnitInstallResult,
) error {

	if result.TargetPath == "" {

		return fmt.Errorf(
			"worker unit target path is empty",
		)
	}

	if result.PreviousExisted {

		info, err := os.Lstat(
			result.BackupPath,
		)

		if err != nil {

			return fmt.Errorf(
				"worker unit backup unavailable: %w",
				err,
			)
		}

		if !info.Mode().IsRegular() {

			return fmt.Errorf(
				"worker unit backup is not a regular file",
			)
		}

		if err := copyFileAtomic(
			result.BackupPath,
			result.TargetPath,
			0644,
		); err != nil {

			return fmt.Errorf(
				"restore worker unit: %w",
				err,
			)
		}

	} else {

		if err := os.Remove(
			result.TargetPath,
		); err != nil &&
			!os.IsNotExist(err) {

			return fmt.Errorf(
				"remove newly installed worker unit: %w",
				err,
			)
		}
	}

	if err := syncDirectory(
		filepath.Dir(
			result.TargetPath,
		),
	); err != nil {

		return fmt.Errorf(
			"sync worker unit directory after rollback: %w",
			err,
		)
	}

	return nil
}
