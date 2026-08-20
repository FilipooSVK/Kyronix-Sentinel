package updater

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const updateWorkerUnitContent = `[Unit]
Description=Kyronix Sentinel Update Worker
Documentation=https://github.com/FilipooSVK/Kyronix-Sentinel
After=network-online.target
Wants=network-online.target

[Service]
Type=oneshot

User=root
Group=root

ExecStart=/usr/local/bin/sentinelctl update install

TimeoutStartSec=5min

StandardOutput=journal
StandardError=journal
`

// EnsureUpdateWorkerUnit provisions the Sentinel update worker unit when
// it is missing.
//
// Existing regular unit files are intentionally left untouched. Normal
// release activation is responsible for upgrading an already installed
// unit. This function exists primarily as the bootstrap bridge for hosts
// upgrading from releases that did not yet install the worker unit.
func EnsureUpdateWorkerUnit(
	ctx context.Context,
	targetPath string,
	reloader SystemdReloader,
) (bool, error) {

	if targetPath == "" {

		return false, fmt.Errorf(
			"worker unit target path is empty",
		)
	}

	if filepath.Base(
		targetPath,
	) != updateWorkerUnitName {

		return false, fmt.Errorf(
			"unexpected worker unit target name: %s",
			filepath.Base(targetPath),
		)
	}

	if reloader == nil {

		return false, fmt.Errorf(
			"systemd reloader is nil",
		)
	}

	info, err := os.Lstat(
		targetPath,
	)

	switch {

	case err == nil:

		if !info.Mode().IsRegular() {

			return false, fmt.Errorf(
				"existing worker unit is not a regular file",
			)
		}

		return false, nil

	case !os.IsNotExist(err):

		return false, fmt.Errorf(
			"inspect worker unit: %w",
			err,
		)
	}

	targetDir := filepath.Dir(
		targetPath,
	)

	if err := os.MkdirAll(
		targetDir,
		0755,
	); err != nil {

		return false, fmt.Errorf(
			"create worker unit directory: %w",
			err,
		)
	}

	tempFile, err := os.CreateTemp(
		targetDir,
		"."+updateWorkerUnitName+"-*.bootstrap",
	)

	if err != nil {

		return false, fmt.Errorf(
			"create worker unit staging file: %w",
			err,
		)
	}

	tempPath := tempFile.Name()

	cleanupTemp := func() {

		_ = tempFile.Close()

		_ = os.Remove(
			tempPath,
		)
	}

	defer cleanupTemp()

	if _, err := tempFile.WriteString(
		updateWorkerUnitContent,
	); err != nil {

		return false, fmt.Errorf(
			"write worker unit staging file: %w",
			err,
		)
	}

	if err := tempFile.Chmod(
		0644,
	); err != nil {

		return false, fmt.Errorf(
			"set worker unit permissions: %w",
			err,
		)
	}

	if err := tempFile.Sync(); err != nil {

		return false, fmt.Errorf(
			"sync worker unit staging file: %w",
			err,
		)
	}

	if err := tempFile.Close(); err != nil {

		return false, fmt.Errorf(
			"close worker unit staging file: %w",
			err,
		)
	}

	// Link rather than rename so an existing unit can never be
	// overwritten by a startup race.
	if err := os.Link(
		tempPath,
		targetPath,
	); err != nil {

		existing, statErr := os.Lstat(
			targetPath,
		)

		if statErr == nil &&
			existing.Mode().IsRegular() {

			return false, nil
		}

		return false, fmt.Errorf(
			"install bootstrap worker unit: %w",
			err,
		)
	}

	if err := syncDirectory(
		targetDir,
	); err != nil {

		removeErr := os.Remove(
			targetPath,
		)

		return false, errors.Join(
			fmt.Errorf(
				"sync worker unit directory: %w",
				err,
			),
			removeErr,
		)
	}

	if err := reloader.DaemonReload(
		ctx,
	); err != nil {

		removeErr := os.Remove(
			targetPath,
		)

		syncErr := syncDirectory(
			targetDir,
		)

		// Reload again after restoring the pre-bootstrap filesystem
		// state. This protects against a partially successful reload.
		rollbackReloadErr :=
			reloader.DaemonReload(
				ctx,
			)

		return false, errors.Join(
			fmt.Errorf(
				"systemd daemon reload after worker bootstrap failed: %w",
				err,
			),
			removeErr,
			syncErr,
			rollbackReloadErr,
		)
	}

	return true, nil
}
