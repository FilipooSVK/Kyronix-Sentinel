package main

import (
	"errors"
	"testing"

	"kyronix/sentinel/internal/updater"
)

func TestUpdateInstallExitCodeSuccess(
	t *testing.T,
) {

	got := updateInstallExitCode(
		updater.InstallExecutorResult{},
		nil,
	)

	if got != updateExitSuccess {
		t.Fatalf(
			"got %d want %d",
			got,
			updateExitSuccess,
		)
	}
}

func TestUpdateInstallExitCodeLocked(
	t *testing.T,
) {

	got := updateInstallExitCode(
		updater.InstallExecutorResult{
			Stage: updater.ExecutorStageLock,
		},
		updater.ErrUpdateOperationLocked,
	)

	if got != updateExitLocked {
		t.Fatalf(
			"got %d want %d",
			got,
			updateExitLocked,
		)
	}
}

func TestUpdateInstallExitCodeQuarantined(
	t *testing.T,
) {

	got := updateInstallExitCode(
		updater.InstallExecutorResult{
			Stage: updater.ExecutorStageQuarantine,
		},
		updater.ErrReleaseQuarantined,
	)

	if got != updateExitQuarantined {
		t.Fatalf(
			"got %d want %d",
			got,
			updateExitQuarantined,
		)
	}
}

func TestUpdateInstallExitCodeStages(
	t *testing.T,
) {

	testErr := errors.New(
		"test failure",
	)

	tests := []struct {
		name string

		result updater.InstallExecutorResult

		want int
	}{
		{
			name: "check",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageCheck,
			},

			want: updateExitCheckFailed,
		},
		{
			name: "state",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageLifecycle,
			},

			want: updateExitStateFailed,
		},
		{
			name: "assets",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageAssets,
			},

			want: updateExitReleaseFailed,
		},
		{
			name: "download",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageDownload,
			},

			want: updateExitDownloadFailed,
		},
		{
			name: "validation",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageValidation,
			},

			want: updateExitValidationFailed,
		},
		{
			name: "activation",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageActivation,
			},

			want: updateExitActivationFailed,
		},
		{
			name: "rollback unsafe",

			result: updater.InstallExecutorResult{
				Stage: updater.ExecutorStageActivation,

				Activation: updater.ActivationResult{
					RolledBack: true,

					RollbackVerified: false,
				},
			},

			want: updateExitRollbackUnsafe,
		},
	}

	for _, tt := range tests {

		t.Run(
			tt.name,
			func(
				t *testing.T,
			) {

				got :=
					updateInstallExitCode(
						tt.result,
						testErr,
					)

				if got != tt.want {

					t.Fatalf(
						"got %d want %d",
						got,
						tt.want,
					)
				}
			},
		)
	}
}

func TestUpdateInstallExitCodeStateRecordFailure(
	t *testing.T,
) {

	got := updateInstallExitCode(
		updater.InstallExecutorResult{
			StateRecordError: errors.New(
				"state failure",
			),
		},
		nil,
	)

	if got != updateExitStateFailed {

		t.Fatalf(
			"got %d want %d",
			got,
			updateExitStateFailed,
		)
	}
}
