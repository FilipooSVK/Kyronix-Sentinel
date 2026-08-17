package app

import (
	"context"
	"time"

	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/logging"
	"kyronix/sentinel/internal/version"
)

// Daemon runs Sentinel engine periodically.
type Daemon struct {
	engine *Engine

	period time.Duration

	logger *logging.Logger

	statusServer *local.Server

	lastResult domain.HealthResult

	lastSnapshot domain.Snapshot
}

// NewDaemon creates Sentinel daemon.
func NewDaemon(
	engine *Engine,
	period time.Duration,
	logger *logging.Logger,
) *Daemon {

	return &Daemon{
		engine: engine,

		period: period,

		logger: logger,

		statusServer: local.NewServer(
			local.DefaultSocket,
			local.Status{
				Running: true,
				Version: version.Version,
			},
		),
	}
}

// Run starts daemon lifecycle.
func (d *Daemon) Run(
	ctx context.Context,
) error {

	err := d.statusServer.Start()

	if err != nil {
		return err
	}

	// Perform initial evaluation immediately after startup.
	d.evaluate()

	d.logger.Info(
		"sentineld started",
		nil,
	)

	ticker := time.NewTicker(
		d.period,
	)

	defer ticker.Stop()

	for {
		select {

		case <-ctx.Done():

			d.logger.Info(
				"sentineld stopping",
				nil,
			)

			return nil

		case <-ticker.C:

			d.evaluate()
		}
	}
}
