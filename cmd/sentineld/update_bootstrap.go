package main

import (
	"context"
	"fmt"
	"time"

	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/logging"
	"kyronix/sentinel/internal/updater"
)

const sentinelUpdateWorkerUnitTarget = "/etc/systemd/system/sentinel-update.service"

func bootstrapUpdateWorker(
	cfg config.Config,
	logger *logging.Logger,
) error {

	if !cfg.Update.Enabled {
		return nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)

	defer cancel()

	installed, err :=
		updater.EnsureUpdateWorkerUnit(
			ctx,
			sentinelUpdateWorkerUnitTarget,
			updater.NewSystemdDaemonReloader(),
		)

	if err != nil {

		return fmt.Errorf(
			"provision update worker unit: %w",
			err,
		)
	}

	if installed {

		logger.Info(
			"update worker unit provisioned",
			map[string]interface{}{
				"path": sentinelUpdateWorkerUnitTarget,
			},
		)
	}

	return nil
}
