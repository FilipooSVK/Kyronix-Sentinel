package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/updater"
	"kyronix/sentinel/internal/version"
)

const (
	sentinelSystemConfig = "/etc/sentinel/sentinel.yaml"

	sentinelInstallDir = "/usr/local/bin"

	sentinelServiceName = "sentineld.service"
)

func runUpdateCommand(
	args []string,
) {

	if len(args) == 0 {

		fmt.Println("usage: sentinelctl update <command>")
		fmt.Println()
		fmt.Println("commands:")
		fmt.Println("  check")
		fmt.Println("  status")
		fmt.Println("  install")
		fmt.Println("  policy")
		fmt.Println("  quarantine [clear]")

		return
	}

	switch args[0] {

	case "check":

		runUpdateCheck()

	case "status":

		runUpdateStatus()

	case "install":

		runUpdateInstall()

	case "policy":

		runUpdatePolicy()

	case "quarantine":

		runUpdateQuarantine(
			args[1:],
		)

	default:

		fmt.Println(
			"unknown update command:",
			args[0],
		)
	}
}

func runUpdateStatus() {

	cfg, err := config.Load(
		sentinelSystemConfig,
	)

	if err != nil {

		fmt.Println(
			"Unable to load Sentinel configuration:",
			err,
		)

		return
	}

	fmt.Println(
		"Kyronix Sentinel Update Status",
	)

	fmt.Println()

	fmt.Println(
		"Current:",
		updater.NormalizeVersion(
			version.Version,
		),
	)

	store := updater.NewStateStore(
		cfg.Update.StatePath,
	)

	state, err := store.Load()

	if err != nil {

		if os.IsNotExist(
			err,
		) {

			fmt.Println(
				"State: NO UPDATE STATE AVAILABLE",
			)

			fmt.Println()

			fmt.Println(
				"No automatic update check has been persisted yet.",
			)

			return
		}

		fmt.Println(
			"State: READ FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println(
		"Latest:",
		state.LatestVersion,
	)

	if state.UpdateAvailable {

		fmt.Println(
			"Update available: yes",
		)

	} else {

		fmt.Println(
			"Update available: no",
		)
	}

	fmt.Println(
		"Last check:",
		state.LastCheck.UTC().Format(
			"2006-01-02 15:04:05 UTC",
		),
	)

	fmt.Println(
		"Next check:",
		state.NextCheck.UTC().Format(
			"2006-01-02 15:04:05 UTC",
		),
	)

	switch state.LastResult {

	case updater.UpdateResultSuccess:

		fmt.Println(
			"Last result: SUCCESS",
		)

	case updater.UpdateResultError:

		fmt.Println(
			"Last result: ERROR",
		)

		fmt.Println(
			"Last error:",
			state.LastError,
		)

	default:

		fmt.Println(
			"Last result:",
			state.LastResult,
		)
	}

	fmt.Println()
	fmt.Println(
		"Installation:",
	)

	if state.LastInstallAttempt == nil {

		fmt.Println(
			"  Last attempt: none",
		)

		printUpdateQuarantineSummary(
			state,
		)

		return
	}

	fmt.Println(
		"  Last attempt:",
		state.LastInstallAttempt.UTC().Format(
			"2006-01-02 15:04:05 UTC",
		),
	)

	fmt.Println(
		"  From:",
		state.LastInstallFromVersion,
	)

	fmt.Println(
		"  Target:",
		state.LastInstallTarget,
	)

	switch state.LastInstallResult {

	case updater.InstallResultInProgress:

		fmt.Println(
			"  Result: IN PROGRESS",
		)

	case updater.InstallResultSuccess:

		fmt.Println(
			"  Result: SUCCESS",
		)

		fmt.Println(
			"  Installed:",
			state.LastInstalledVersion,
		)

	case updater.InstallResultFailed:

		fmt.Println(
			"  Result: FAILED",
		)

		fmt.Println(
			"  Error:",
			state.LastInstallError,
		)

		if state.LastRollback {

			if state.LastRollbackVerified {

				fmt.Println(
					"  Rollback: VERIFIED",
				)

				fmt.Println(
					"  Recovered:",
					state.RecoveredVersion,
				)

			} else {

				fmt.Println(
					"  Rollback: NOT VERIFIED",
				)
			}

		} else {

			fmt.Println(
				"  Rollback: not required",
			)
		}

	default:

		fmt.Println(
			"  Result:",
			state.LastInstallResult,
		)
	}

	printUpdateQuarantineSummary(
		state,
	)
}

func runUpdateCheck() {

	cfg, err := config.Load(
		sentinelSystemConfig,
	)

	if err != nil {

		fmt.Println(
			"Unable to load Sentinel configuration:",
			err,
		)

		return
	}

	fmt.Println("Kyronix Sentinel Update")
	fmt.Println()

	fmt.Println(
		"Current:",
		updater.NormalizeVersion(
			version.Version,
		),
	)

	if !cfg.Update.Enabled {

		fmt.Println(
			"Status: UPDATE CHECK DISABLED",
		)

		return
	}

	if !updateRepositoryConfigured(
		cfg,
	) {

		fmt.Println(
			"Status: UPDATE REPOSITORY NOT CONFIGURED",
		)

		fmt.Println()

		fmt.Println(
			"Configure update.owner and update.repository in",
			sentinelSystemConfig,
		)

		return
	}

	client := updater.NewGitHubClient(
		cfg.Update.Owner,
		cfg.Update.Repository,
	)

	result, err := client.Check(
		context.Background(),
		version.Version,
	)

	if err != nil {

		fmt.Println(
			"Status: CHECK FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println(
		"Latest:",
		result.LatestVersion,
	)

	if result.UpdateAvailable {

		fmt.Println(
			"Status: UPDATE AVAILABLE",
		)

		return
	}

	fmt.Println(
		"Status: UP TO DATE",
	)
}

func runUpdateInstall() {

	fmt.Println("Kyronix Sentinel Update")
	fmt.Println()

	fmt.Println(
		"Current:",
		updater.NormalizeVersion(
			version.Version,
		),
	)

	if os.Geteuid() != 0 {

		fmt.Println(
			"Status: ROOT PRIVILEGES REQUIRED",
		)

		fmt.Println()

		fmt.Println(
			"Run:",
			"sudo sentinelctl update install",
		)

		return
	}

	cfg, err := config.Load(
		sentinelSystemConfig,
	)

	if err != nil {

		fmt.Println(
			"Status: CONFIGURATION ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	if !cfg.Update.Enabled {

		fmt.Println(
			"Status: UPDATE SYSTEM DISABLED",
		)

		return
	}

	if !updateRepositoryConfigured(
		cfg,
	) {

		fmt.Println(
			"Status: UPDATE REPOSITORY NOT CONFIGURED",
		)

		fmt.Println()

		fmt.Println(
			"Configure update.owner and update.repository in",
			sentinelSystemConfig,
		)

		return
	}

	fmt.Println(
		"Checking GitHub release...",
	)

	client := updater.NewGitHubClient(
		cfg.Update.Owner,
		cfg.Update.Repository,
	)

	ctx := context.Background()

	check, err := client.Check(
		ctx,
		version.Version,
	)

	if err != nil {

		fmt.Println(
			"Status: UPDATE CHECK FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println(
		"Latest:",
		check.LatestVersion,
	)

	stateStore, err := persistUpdateCheckState(
		cfg,
		check,
	)

	if err != nil {

		fmt.Println(
			"Status: UPDATE STATE ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	if !check.UpdateAvailable {

		fmt.Println(
			"Status: UP TO DATE",
		)

		return
	}

	quarantined, err :=
		stateStore.IsVersionQuarantined(
			check.LatestVersion,
		)

	if err != nil {

		fmt.Println(
			"Status: UPDATE STATE ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	if quarantined {

		state, err :=
			stateStore.Load()

		if err != nil {

			fmt.Println(
				"Status: UPDATE STATE ERROR",
			)

			fmt.Println(
				"Error:",
				err,
			)

			return
		}

		fmt.Println(
			"Status: RELEASE QUARANTINED",
		)

		fmt.Println(
			"Version:",
			check.LatestVersion,
		)

		fmt.Println(
			"Failures:",
			state.QuarantineFailureCount,
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

	fmt.Println(
		"Status: UPDATE AVAILABLE",
	)

	if err := startUpdateInstallLifecycle(
		stateStore,
		version.Version,
		check.LatestVersion,
	); err != nil {

		fmt.Println(
			"Status: UPDATE STATE ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	assets, err := updater.SelectCurrentPlatformAssets(
		check.Release,
	)

	if err != nil {

		recordUpdateInstallFailure(
			stateStore,
			err,
			false,
			false,
		)

		fmt.Println(
			"Status: RELEASE ASSET ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Printf(
		"Platform: %s/%s\n",
		assets.OS,
		assets.Arch,
	)

	fmt.Println(
		"Package:",
		assets.Package.Name,
	)

	fmt.Println(
		"Checksum:",
		assets.Checksum.Name,
	)

	workDir, err := os.MkdirTemp(
		"",
		"sentinel-update-*",
	)

	if err != nil {

		recordUpdateInstallFailure(
			stateStore,
			err,
			false,
			false,
		)

		fmt.Println(
			"Status: WORK DIRECTORY ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	defer os.RemoveAll(
		workDir,
	)

	downloadDir := filepath.Join(
		workDir,
		"download",
	)

	fmt.Println(
		"Downloading and verifying release...",
	)

	artifact, err := updater.DownloadAndVerify(
		ctx,
		assets,
		downloadDir,
	)

	if err != nil {

		recordUpdateInstallFailure(
			stateStore,
			err,
			false,
			false,
		)

		fmt.Println(
			"Status: DOWNLOAD VERIFICATION FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println(
		"SHA256:",
		artifact.SHA256,
	)

	fmt.Println(
		"Release package verified.",
	)

	extractDir := filepath.Join(
		workDir,
		"extract",
	)

	fmt.Println(
		"Validating release manifest...",
	)

	release, err := updater.ExtractAndValidateRelease(
		artifact,
		extractDir,
		check.LatestVersion,
		assets.OS,
		assets.Arch,
	)

	if err != nil {

		recordUpdateInstallFailure(
			stateStore,
			err,
			false,
			false,
		)

		fmt.Println(
			"Status: RELEASE VALIDATION FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println(
		"Manifest verified.",
	)

	fmt.Println(
		"Installing release...",
	)

	service := updater.NewSystemdController(
		sentinelServiceName,
	)

	health := updater.NewUnixHealthChecker(
		local.DefaultSocket,
	)

	activation, err := updater.ActivateRelease(
		ctx,
		release,
		sentinelInstallDir,
		check.LatestVersion,
		version.Version,
		service,
		health,
	)

	if err != nil {

		recordUpdateInstallFailure(
			stateStore,
			err,
			activation.RolledBack,
			activation.RollbackVerified,
		)

		fmt.Println(
			"Status: UPDATE FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		if activation.RolledBack {

			if activation.RollbackVerified {

				fmt.Println(
					"Rollback: VERIFIED",
				)

			} else {

				fmt.Println(
					"Rollback: NOT VERIFIED",
				)
			}
		}

		return
	}

	recordUpdateInstallSuccess(
		stateStore,
	)

	fmt.Println()
	fmt.Println(
		"Status: UPDATE SUCCESSFUL",
	)

	fmt.Println(
		"Installed:",
		activation.InstalledVersion,
	)

	fmt.Println(
		"Backup:",
		activation.Install.SentineldBackupPath,
	)

	fmt.Println(
		"Backup:",
		activation.Install.SentinelctlBackupPath,
	)
}

func updateRepositoryConfigured(
	cfg config.Config,
) bool {

	return cfg.Update.Owner != "" &&
		cfg.Update.Repository != ""
}
