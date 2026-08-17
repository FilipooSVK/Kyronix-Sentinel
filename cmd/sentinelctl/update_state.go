package main

import (
	"fmt"
	"time"

	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/updater"
)

// persistUpdateCheckState stores the GitHub release check performed
// as part of a manual update installation.
func persistUpdateCheckState(
	cfg config.Config,
	check updater.CheckResult,
) (*updater.StateStore, error) {

	if cfg.Update.CheckInterval <= 0 {

		return nil, fmt.Errorf(
			"update check interval must be greater than zero",
		)
	}

	store := updater.NewStateStore(
		cfg.Update.StatePath,
	)

	checkedAt := time.Now().UTC()

	state := updater.UpdateState{
		LastCheck: checkedAt,

		NextCheck: checkedAt.Add(
			cfg.Update.CheckInterval,
		),

		CurrentVersion: check.CurrentVersion,

		LatestVersion: check.LatestVersion,

		UpdateAvailable: check.UpdateAvailable,

		LastResult: updater.UpdateResultSuccess,
	}

	if err := store.SaveCheck(
		state,
	); err != nil {

		return store, err
	}

	return store, nil
}

// startUpdateInstallLifecycle records that Sentinel is about
// to begin modifying the installed software.
func startUpdateInstallLifecycle(
	store *updater.StateStore,
	fromVersion string,
	targetVersion string,
) error {

	if store == nil {

		return fmt.Errorf(
			"update state store is nil",
		)
	}

	return store.RecordInstallStarted(
		time.Now().UTC(),
		fromVersion,
		targetVersion,
	)
}

// recordUpdateInstallFailure persists a failed update attempt.
//
// Persistence failure is reported to the operator but does not hide
// the original update failure.
func recordUpdateInstallFailure(
	store *updater.StateStore,
	installErr error,
	rolledBack bool,
	rollbackVerified bool,
) {

	if store == nil ||
		installErr == nil {

		return
	}

	if err := store.RecordInstallFailure(
		installErr.Error(),
		rolledBack,
		rollbackVerified,
	); err != nil {

		fmt.Println()
		fmt.Println(
			"Update state: RECORD FAILED",
		)

		fmt.Println(
			"State error:",
			err,
		)
	}
}

// recordUpdateInstallSuccess persists a successful activation.
//
// The software update itself remains successful even if the
// post-activation audit state cannot be persisted.
func recordUpdateInstallSuccess(
	store *updater.StateStore,
) {

	if store == nil {
		return
	}

	if err := store.RecordInstallSuccess(); err != nil {

		fmt.Println()
		fmt.Println(
			"Update state: RECORD FAILED",
		)

		fmt.Println(
			"State error:",
			err,
		)
	}
}
