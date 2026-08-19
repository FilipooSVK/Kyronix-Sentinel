package main

import (
	"errors"
	"fmt"

	"kyronix/sentinel/internal/updater"
)

func printUpdateExecutorEvent(
	event updater.ExecutorEvent,
) {

	switch event.Type {

	case updater.ExecutorEventChecking:

		fmt.Println(
			"Checking GitHub release...",
		)

	case updater.ExecutorEventChecked:

		fmt.Println(
			"Latest:",
			event.Check.LatestVersion,
		)

	case updater.ExecutorEventUpdateAvailable:

		fmt.Println(
			"Status: UPDATE AVAILABLE",
		)

	case updater.ExecutorEventAssetsSelected:

		fmt.Printf(
			"Platform: %s/%s\n",
			event.Assets.OS,
			event.Assets.Arch,
		)

		fmt.Println(
			"Package:",
			event.Assets.Package.Name,
		)

		fmt.Println(
			"Checksum:",
			event.Assets.Checksum.Name,
		)

	case updater.ExecutorEventDownloading:

		fmt.Println(
			"Downloading and verifying release...",
		)

	case updater.ExecutorEventDownloaded:

		fmt.Println(
			"SHA256:",
			event.Artifact.SHA256,
		)

		fmt.Println(
			"Release package verified.",
		)

	case updater.ExecutorEventValidating:

		fmt.Println(
			"Validating release manifest...",
		)

	case updater.ExecutorEventValidated:

		fmt.Println(
			"Manifest verified.",
		)

	case updater.ExecutorEventInstalling:

		fmt.Println(
			"Installing release...",
		)
	}
}

func printUpdateExecutorFailure(
	result updater.InstallExecutorResult,
	executorErr error,
) {

	if errors.Is(
		executorErr,
		updater.ErrUpdateOperationLocked,
	) {

		fmt.Println(
			"Status: UPDATE ALREADY IN PROGRESS",
		)

		fmt.Println()

		fmt.Println(
			"Another Sentinel update operation currently holds the global update lock.",
		)

		return
	}

	if errors.Is(
		executorErr,
		updater.ErrReleaseQuarantined,
	) {

		fmt.Println(
			"Status: RELEASE QUARANTINED",
		)

		fmt.Println(
			"Version:",
			result.Check.LatestVersion,
		)

		fmt.Println(
			"Failures:",
			result.QuarantineFailures,
		)

		fmt.Println()

		fmt.Println(
			"Installation refused.",
		)

		fmt.Println(
			"To explicitly clear quarantine:",
		)

		fmt.Println(
			"sudo sentinelctl update quarantine clear",
		)

		return
	}

	switch result.Stage {

	case updater.ExecutorStageLock:

		fmt.Println(
			"Status: UPDATE LOCK ERROR",
		)

	case updater.ExecutorStageCheck:

		fmt.Println(
			"Status: UPDATE CHECK FAILED",
		)

	case updater.ExecutorStagePersistCheck,
		updater.ExecutorStageQuarantine,
		updater.ExecutorStageLifecycle:

		fmt.Println(
			"Status: UPDATE STATE ERROR",
		)

	case updater.ExecutorStageAssets:

		fmt.Println(
			"Status: RELEASE ASSET ERROR",
		)

	case updater.ExecutorStageWorkDir:

		fmt.Println(
			"Status: WORK DIRECTORY ERROR",
		)

	case updater.ExecutorStageDownload:

		fmt.Println(
			"Status: DOWNLOAD VERIFICATION FAILED",
		)

	case updater.ExecutorStageValidation:

		fmt.Println(
			"Status: RELEASE VALIDATION FAILED",
		)

	case updater.ExecutorStageActivation:

		fmt.Println(
			"Status: UPDATE FAILED",
		)

	default:

		fmt.Println(
			"Status: UPDATE FAILED",
		)
	}

	fmt.Println(
		"Error:",
		executorErr,
	)

	if result.Stage ==
		updater.ExecutorStageActivation &&
		result.Activation.RolledBack {

		if result.Activation.RollbackVerified {

			fmt.Println(
				"Rollback: VERIFIED",
			)

		} else {

			fmt.Println(
				"Rollback: NOT VERIFIED",
			)
		}
	}
}

func printUpdateExecutorWarnings(
	result updater.InstallExecutorResult,
) {

	if result.StateRecordError != nil {

		fmt.Println()

		fmt.Println(
			"Update state: RECORD FAILED",
		)

		fmt.Println(
			"State error:",
			result.StateRecordError,
		)
	}

	if result.CleanupError != nil {

		fmt.Println()

		fmt.Println(
			"Warning: failed to release update operation lock:",
			result.CleanupError,
		)
	}
}
