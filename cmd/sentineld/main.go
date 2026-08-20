package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"kyronix/sentinel/internal/app"
	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/logging"
	"kyronix/sentinel/internal/updater"
	"kyronix/sentinel/internal/version"
)

const sentinelUpdateServiceName = "sentinel-update.service"

func main() {

	configPath := flag.String(
		"config",
		"configs/sentinel.yaml",
		"path to Sentinel configuration file",
	)

	flag.Parse()

	cfg, err := config.Load(
		*configPath,
	)

	if err != nil {
		panic(err)
	}

	logger := logging.New(
		os.Stdout,
		cfg.Logging.Level,
	)

	if err := bootstrapUpdateWorker(
		cfg,
		logger,
	); err != nil {

		logger.Error(
			"update worker bootstrap failed",
			map[string]interface{}{
				"error": err.Error(),
			},
		)

		os.Exit(1)
	}

	logger.Info(
		"Kyronix Sentinel daemon started",
		map[string]interface{}{
			"version": version.Version,
		},
	)

	engine := app.NewDefaultEngine(
		cfg,
	)

	if cfg.History.Persistence {

		err := engine.ConfigureHistoryPersistence(
			cfg.History.Path,
			cfg.History.Size,
		)

		if err != nil {

			logger.Error(
				"persistent history restore failed",
				map[string]interface{}{
					"path":  cfg.History.Path,
					"error": err.Error(),
				},
			)

		} else {

			logger.Info(
				"persistent history restored",
				map[string]interface{}{
					"path":    cfg.History.Path,
					"entries": len(engine.History()),
				},
			)
		}

	} else {

		logger.Info(
			"persistent history disabled",
			nil,
		)
	}

	daemon := app.NewDaemon(
		engine,
		cfg.Daemon.Interval,
		logger,
	)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	defer stop()

	startAutomaticUpdateChecker(
		ctx,
		cfg,
		logger,
	)

	if err := daemon.Run(
		ctx,
	); err != nil {

		logger.Error(
			"daemon failed",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}
}

func startAutomaticUpdateChecker(
	ctx context.Context,
	cfg config.Config,
	logger *logging.Logger,
) {

	if !cfg.Update.Enabled {

		logger.Info(
			"software update system disabled",
			nil,
		)

		return
	}

	if !cfg.Update.AutoCheck {

		logger.Info(
			"automatic update check disabled",
			nil,
		)

		return
	}

	if cfg.Update.Owner == "" ||
		cfg.Update.Repository == "" {

		logger.Error(
			"automatic update check not started",
			map[string]interface{}{
				"error": "update repository is not configured",
			},
		)

		return
	}

	if cfg.Update.CheckInterval <= 0 {

		logger.Error(
			"automatic update check not started",
			map[string]interface{}{
				"error": "check interval must be greater than zero",
			},
		)

		return
	}

	if cfg.Update.AutoInstallPolicy.MinReleaseAge < 0 {

		logger.Error(
			"automatic update checker not started",
			map[string]interface{}{
				"error": "automatic install minimum release age cannot be negative",
			},
		)

		return
	}

	client := updater.NewGitHubClient(
		cfg.Update.Owner,
		cfg.Update.Repository,
	)

	stateStore := updater.NewStateStore(
		cfg.Update.StatePath,
	)

	monitor := updater.NewMonitor(
		client,
		version.Version,
		cfg.Update.CheckInterval,
		logger,
		stateStore,
	)

	monitor.SetAutoInstallPolicy(
		updater.AutoInstallPolicy{
			Enabled: cfg.Update.AutoInstall,

			MinReleaseAge: cfg.Update.AutoInstallPolicy.MinReleaseAge,

			PatchOnly: cfg.Update.AutoInstallPolicy.PatchOnly,
		},
	)

	executionMode := updater.AutoInstallExecutionMode(
		cfg.Update.AutoInstallMode,
	)

	var updateWorker updater.UpdateWorker

	if executionMode ==
		updater.AutoInstallExecutionWorkerEnabled {

		updateWorker =
			updater.NewSystemdUpdateWorker(
				sentinelUpdateServiceName,
			)
	}

	if err := monitor.SetAutoInstallExecution(
		executionMode,
		updateWorker,
	); err != nil {

		logger.Error(
			"automatic update checker not started",
			map[string]interface{}{
				"error": err.Error(),

				"auto_install_mode": cfg.Update.AutoInstallMode,
			},
		)

		return
	}

	logger.Info(
		"automatic update checker started",
		map[string]interface{}{
			"owner":          cfg.Update.Owner,
			"repository":     cfg.Update.Repository,
			"check_interval": cfg.Update.CheckInterval.String(),
			"state_path":     cfg.Update.StatePath,

			"auto_install": cfg.Update.AutoInstall,

			"policy_mode": string(executionMode),

			"min_release_age": cfg.Update.AutoInstallPolicy.MinReleaseAge.String(),

			"patch_only": cfg.Update.AutoInstallPolicy.PatchOnly,
		},
	)

	go func() {

		if err := monitor.Run(
			ctx,
		); err != nil {

			logger.Error(
				"automatic update checker stopped",
				map[string]interface{}{
					"error": err.Error(),
				},
			)
		}
	}()
}
