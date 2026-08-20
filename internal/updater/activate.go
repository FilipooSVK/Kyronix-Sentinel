package updater

import (
	"context"
	"errors"
	"fmt"
)

// ActivationOptions configures release-managed auxiliary resources.
type ActivationOptions struct {
	WorkerUnitTarget string

	Reloader SystemdReloader
}

// ActivationResult describes the result of installing
// and activating a Sentinel release.
type ActivationResult struct {
	Install InstallResult

	WorkerUnit *UnitInstallResult

	InstalledVersion string

	RolledBack bool

	RollbackVerified bool
}

// ActivateRelease installs a validated release, restarts Sentinel,
// verifies the new version and automatically restores the previous
// release when activation fails.
//
// When ActivationOptions are supplied, the Sentinel update worker
// systemd unit participates in the same activation transaction.
func ActivateRelease(
	ctx context.Context,
	release ExtractedRelease,
	targetDir string,
	expectedVersion string,
	previousVersion string,
	service ServiceController,
	health VersionHealthChecker,
	options ...ActivationOptions,
) (ActivationResult, error) {

	if service == nil {

		return ActivationResult{}, fmt.Errorf(
			"service controller is nil",
		)
	}

	if health == nil {

		return ActivationResult{}, fmt.Errorf(
			"health checker is nil",
		)
	}

	if !IsValidVersion(
		expectedVersion,
	) {

		return ActivationResult{}, fmt.Errorf(
			"invalid expected version: %s",
			expectedVersion,
		)
	}

	if !IsValidVersion(
		previousVersion,
	) {

		return ActivationResult{}, fmt.Errorf(
			"invalid previous version: %s",
			previousVersion,
		)
	}

	activationOptions, workerEnabled, err :=
		resolveActivationOptions(
			options,
		)

	if err != nil {
		return ActivationResult{}, err
	}

	installResult, err := InstallRelease(
		release,
		targetDir,
	)

	if err != nil {
		return ActivationResult{}, err
	}

	result := ActivationResult{
		Install: installResult,

		InstalledVersion: NormalizeVersion(
			expectedVersion,
		),
	}

	var workerInstall *UnitInstallResult

	if workerEnabled {

		unitResult, unitErr := InstallWorkerUnit(
			release.WorkerUnitPath,
			activationOptions.WorkerUnitTarget,
		)

		if unitErr != nil {

			result.RolledBack = true

			rollbackErr := rollbackActivation(
				ctx,
				targetDir,
				previousVersion,
				nil,
				activationOptions.Reloader,
				service,
				health,
			)

			if rollbackErr != nil {

				return result, fmt.Errorf(
					"worker unit installation failed: %v; rollback failed: %v",
					unitErr,
					rollbackErr,
				)
			}

			result.RollbackVerified = true

			return result, fmt.Errorf(
				"worker unit installation failed: %w; previous release restored",
				unitErr,
			)
		}

		workerInstall = &unitResult

		result.WorkerUnit =
			workerInstall

		if reloadErr :=
			activationOptions.Reloader.DaemonReload(
				ctx,
			); reloadErr != nil {

			result.RolledBack = true

			rollbackErr := rollbackActivation(
				ctx,
				targetDir,
				previousVersion,
				workerInstall,
				activationOptions.Reloader,
				service,
				health,
			)

			if rollbackErr != nil {

				return result, fmt.Errorf(
					"systemd daemon reload failed after worker unit installation: %v; rollback failed: %v",
					reloadErr,
					rollbackErr,
				)
			}

			result.RollbackVerified = true

			return result, fmt.Errorf(
				"systemd daemon reload failed after worker unit installation: %w; previous release restored",
				reloadErr,
			)
		}
	}

	if err := service.Restart(
		ctx,
	); err != nil {

		result.RolledBack = true

		rollbackErr := rollbackActivation(
			ctx,
			targetDir,
			previousVersion,
			workerInstall,
			activationOptions.Reloader,
			service,
			health,
		)

		if rollbackErr != nil {

			return result, fmt.Errorf(
				"new release installed but service restart failed: %v; rollback failed: %v",
				err,
				rollbackErr,
			)
		}

		result.RollbackVerified = true

		return result, fmt.Errorf(
			"new release activation failed during service restart: %w; previous release restored",
			err,
		)
	}

	if err := health.WaitForVersion(
		ctx,
		expectedVersion,
	); err != nil {

		result.RolledBack = true

		rollbackErr := rollbackActivation(
			ctx,
			targetDir,
			previousVersion,
			workerInstall,
			activationOptions.Reloader,
			service,
			health,
		)

		if rollbackErr != nil {

			return result, fmt.Errorf(
				"new release health check failed: %v; rollback failed: %v",
				err,
				rollbackErr,
			)
		}

		result.RollbackVerified = true

		return result, fmt.Errorf(
			"new release health check failed: %w; previous release restored",
			err,
		)
	}

	return result, nil
}

func resolveActivationOptions(
	options []ActivationOptions,
) (
	ActivationOptions,
	bool,
	error,
) {

	if len(options) == 0 {

		return ActivationOptions{},
			false,
			nil
	}

	if len(options) > 1 {

		return ActivationOptions{},
			false,
			fmt.Errorf(
				"multiple activation option sets are not supported",
			)
	}

	option := options[0]

	if option.WorkerUnitTarget == "" {

		return ActivationOptions{},
			false,
			fmt.Errorf(
				"worker unit target path is empty",
			)
	}

	if option.Reloader == nil {

		return ActivationOptions{},
			false,
			fmt.Errorf(
				"systemd reloader is nil",
			)
	}

	return option,
		true,
		nil
}

func rollbackActivation(
	ctx context.Context,
	targetDir string,
	previousVersion string,
	workerInstall *UnitInstallResult,
	reloader SystemdReloader,
	service ServiceController,
	health VersionHealthChecker,
) error {

	var rollbackErrors []error

	binariesRestored := true

	if err := RestorePrevious(
		targetDir,
	); err != nil {

		binariesRestored = false

		rollbackErrors = append(
			rollbackErrors,
			fmt.Errorf(
				"restore previous binaries failed: %w",
				err,
			),
		)
	}

	if workerInstall != nil {

		if err := RestoreWorkerUnit(
			*workerInstall,
		); err != nil {

			rollbackErrors = append(
				rollbackErrors,
				fmt.Errorf(
					"restore previous worker unit failed: %w",
					err,
				),
			)
		}
	}

	if reloader != nil {

		if err := reloader.DaemonReload(
			ctx,
		); err != nil {

			rollbackErrors = append(
				rollbackErrors,
				fmt.Errorf(
					"systemd daemon reload after rollback failed: %w",
					err,
				),
			)
		}
	}

	// Only restart if the old binaries were actually restored.
	if binariesRestored {

		if err := service.Restart(
			ctx,
		); err != nil {

			rollbackErrors = append(
				rollbackErrors,
				fmt.Errorf(
					"restart after rollback failed: %w",
					err,
				),
			)

		} else if err := health.WaitForVersion(
			ctx,
			previousVersion,
		); err != nil {

			rollbackErrors = append(
				rollbackErrors,
				fmt.Errorf(
					"rollback health verification failed: %w",
					err,
				),
			)
		}
	}

	return errors.Join(
		rollbackErrors...,
	)
}
