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
	"kyronix/sentinel/internal/version"
)

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

	if err := daemon.Run(ctx); err != nil {

		logger.Error(
			"daemon failed",
			map[string]interface{}{
				"error": err.Error(),
			},
		)
	}
}
