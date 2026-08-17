package updater

import (
	"context"
	"fmt"
)

// ActivationResult describes the result of installing
// and activating a Sentinel release.
type ActivationResult struct {
	Install InstallResult

	InstalledVersion string

	RolledBack bool

	RollbackVerified bool
}

// ActivateRelease installs a validated release, restarts Sentinel,
// verifies the new version and automatically restores the previous
// release when activation fails.
func ActivateRelease(
	ctx context.Context,
	release ExtractedRelease,
	targetDir string,
	expectedVersion string,
	previousVersion string,
	service ServiceController,
	health VersionHealthChecker,
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

	if err := service.Restart(
		ctx,
	); err != nil {

		result.RolledBack = true

		rollbackErr := rollbackActivation(
			ctx,
			targetDir,
			previousVersion,
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

func rollbackActivation(
	ctx context.Context,
	targetDir string,
	previousVersion string,
	service ServiceController,
	health VersionHealthChecker,
) error {

	if err := RestorePrevious(
		targetDir,
	); err != nil {

		return fmt.Errorf(
			"restore previous binaries failed: %w",
			err,
		)
	}

	if err := service.Restart(
		ctx,
	); err != nil {

		return fmt.Errorf(
			"restart after rollback failed: %w",
			err,
		)
	}

	if err := health.WaitForVersion(
		ctx,
		previousVersion,
	); err != nil {

		return fmt.Errorf(
			"rollback health verification failed: %w",
			err,
		)
	}

	return nil
}
