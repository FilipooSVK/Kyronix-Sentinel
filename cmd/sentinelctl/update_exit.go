package main

import (
	"errors"

	"kyronix/sentinel/internal/updater"
)

const (
	updateExitSuccess = 0
	updateExitUsage   = 2

	updateExitLocked       = 10
	updateExitQuarantined  = 11
	updateExitPrecondition = 12

	updateExitCheckFailed      = 20
	updateExitStateFailed      = 21
	updateExitReleaseFailed    = 22
	updateExitDownloadFailed   = 23
	updateExitValidationFailed = 24
	updateExitActivationFailed = 25
	updateExitRollbackUnsafe   = 26

	updateExitInternal = 30
)

func updateInstallExitCode(
	result updater.InstallExecutorResult,
	executorErr error,
) int {

	if executorErr == nil {

		if result.StateRecordError != nil {
			return updateExitStateFailed
		}

		if result.CleanupError != nil {
			return updateExitInternal
		}

		return updateExitSuccess
	}

	if errors.Is(
		executorErr,
		updater.ErrUpdateOperationLocked,
	) {
		return updateExitLocked
	}

	if errors.Is(
		executorErr,
		updater.ErrReleaseQuarantined,
	) {
		return updateExitQuarantined
	}

	switch result.Stage {

	case updater.ExecutorStageCheck:

		return updateExitCheckFailed

	case updater.ExecutorStagePersistCheck,
		updater.ExecutorStageQuarantine,
		updater.ExecutorStageLifecycle:

		return updateExitStateFailed

	case updater.ExecutorStageAssets:

		return updateExitReleaseFailed

	case updater.ExecutorStageDownload:

		return updateExitDownloadFailed

	case updater.ExecutorStageValidation:

		return updateExitValidationFailed

	case updater.ExecutorStageActivation:

		if result.Activation.RolledBack &&
			!result.Activation.RollbackVerified {

			return updateExitRollbackUnsafe
		}

		return updateExitActivationFailed

	case updater.ExecutorStageLock,
		updater.ExecutorStageWorkDir:

		return updateExitInternal

	default:

		return updateExitInternal
	}
}
