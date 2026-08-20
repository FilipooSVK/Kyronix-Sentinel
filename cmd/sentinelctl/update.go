package main

import (
	"context"
	"fmt"
	"os"

	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/updater"
	"kyronix/sentinel/internal/version"
)

const (
	sentinelSystemConfig = "/etc/sentinel/sentinel.yaml"

	sentinelInstallDir = "/usr/local/bin"

	sentinelServiceName = "sentineld.service"

	sentinelWorkerUnitTarget = "/etc/systemd/system/sentinel-update.service"
)

func runUpdateCommand(
	args []string,
) int {

	if len(args) == 0 {

		fmt.Println("usage: sentinelctl update <command>")
		fmt.Println()
		fmt.Println("commands:")
		fmt.Println("  check")
		fmt.Println("  status")
		fmt.Println("  install")
		fmt.Println("  policy")
		fmt.Println("  quarantine [clear]")

		return updateExitUsage
	}

	switch args[0] {

	case "check":

		runUpdateCheck()

		return updateExitSuccess

	case "status":

		runUpdateStatus()

		return updateExitSuccess

	case "install":

		return runUpdateInstall()

	case "policy":

		runUpdatePolicy()

		return updateExitSuccess

	case "quarantine":

		runUpdateQuarantine(
			args[1:],
		)

		return updateExitSuccess

	default:

		fmt.Println(
			"unknown update command:",
			args[0],
		)

		return updateExitUsage
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

func runUpdateInstall() int {

	fmt.Println(
		"Kyronix Sentinel Update",
	)

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

		return updateExitPrecondition
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

		return updateExitPrecondition
	}

	if !cfg.Update.Enabled {

		fmt.Println(
			"Status: UPDATE SYSTEM DISABLED",
		)

		return updateExitPrecondition
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

		return updateExitPrecondition
	}

	client := updater.NewGitHubClient(
		cfg.Update.Owner,
		cfg.Update.Repository,
	)

	service := updater.NewSystemdController(
		sentinelServiceName,
	)

	reloader := updater.NewSystemdDaemonReloader()

	health := updater.NewUnixHealthChecker(
		local.DefaultSocket,
	)

	executor := updater.NewInstallExecutor(
		client,
		updater.InstallExecutorConfig{
			CurrentVersion: version.Version,

			CheckInterval: cfg.Update.CheckInterval,

			StatePath: cfg.Update.StatePath,

			InstallDir: sentinelInstallDir,

			WorkerUnitTarget: sentinelWorkerUnitTarget,
		},
		service,
		reloader,
		health,
	)

	result, err := executor.Execute(
		context.Background(),
		printUpdateExecutorEvent,
	)

	if err != nil {

		printUpdateExecutorFailure(
			result,
			err,
		)

		printUpdateExecutorWarnings(
			result,
		)

		return updateInstallExitCode(
			result,
			err,
		)
	}

	if result.UpToDate {

		fmt.Println(
			"Status: UP TO DATE",
		)

		printUpdateExecutorWarnings(
			result,
		)

		return updateInstallExitCode(
			result,
			nil,
		)
	}

	fmt.Println()

	fmt.Println(
		"Status: UPDATE SUCCESSFUL",
	)

	fmt.Println(
		"Installed:",
		result.Activation.InstalledVersion,
	)

	fmt.Println(
		"Backup:",
		result.Activation.Install.SentineldBackupPath,
	)

	fmt.Println(
		"Backup:",
		result.Activation.Install.SentinelctlBackupPath,
	)

	printUpdateExecutorWarnings(
		result,
	)

	return updateInstallExitCode(
		result,
		nil,
	)
}

func updateRepositoryConfigured(
	cfg config.Config,
) bool {

	return cfg.Update.Owner != "" &&
		cfg.Update.Repository != ""
}
