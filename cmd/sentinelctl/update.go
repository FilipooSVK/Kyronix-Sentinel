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

		return
	}

	switch args[0] {

	case "check":

		runUpdateCheck()

	case "status":

		runUpdateStatus()

	case "install":

		runUpdateInstall()

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

	if !check.UpdateAvailable {

		fmt.Println(
			"Status: UP TO DATE",
		)

		return
	}

	fmt.Println(
		"Status: UPDATE AVAILABLE",
	)

	assets, err := updater.SelectCurrentPlatformAssets(
		check.Release,
	)

	if err != nil {

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
