package updater

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// UpdateWorker starts the out-of-process Sentinel update worker.
//
// Start only requests worker execution. It does not wait for the
// update transaction itself to complete.
type UpdateWorker interface {
	Start(
		ctx context.Context,
	) error
}

type workerCommandRunner func(
	ctx context.Context,
	command string,
	args ...string,
) ([]byte, error)

// SystemdUpdateWorker requests execution of the dedicated Sentinel
// update worker through systemd.
type SystemdUpdateWorker struct {
	Service string

	run workerCommandRunner
}

// NewSystemdUpdateWorker creates a production systemd update worker.
func NewSystemdUpdateWorker(
	service string,
) *SystemdUpdateWorker {

	return &SystemdUpdateWorker{
		Service: strings.TrimSpace(
			service,
		),

		run: runWorkerCommand,
	}
}

// Start queues the configured update worker in systemd.
//
// --no-block is intentional. sentineld must not remain blocked in a
// child systemctl process because the update worker may restart
// sentineld as part of activation.
func (w *SystemdUpdateWorker) Start(
	ctx context.Context,
) error {

	if w == nil {
		return fmt.Errorf(
			"systemd update worker is nil",
		)
	}

	if ctx == nil {
		return fmt.Errorf(
			"update worker context is nil",
		)
	}

	service := strings.TrimSpace(
		w.Service,
	)

	if service == "" {
		return fmt.Errorf(
			"systemd update worker service name is empty",
		)
	}

	if w.run == nil {
		return fmt.Errorf(
			"systemd update worker command runner is nil",
		)
	}

	output, err := w.run(
		ctx,
		"systemctl",
		"start",
		"--no-block",
		service,
	)

	if err != nil {

		message := strings.TrimSpace(
			string(output),
		)

		if message == "" {

			return fmt.Errorf(
				"systemctl start --no-block %s failed: %w",
				service,
				err,
			)
		}

		return fmt.Errorf(
			"systemctl start --no-block %s failed: %w: %s",
			service,
			err,
			message,
		)
	}

	return nil
}

func runWorkerCommand(
	ctx context.Context,
	command string,
	args ...string,
) ([]byte, error) {

	cmd := exec.CommandContext(
		ctx,
		command,
		args...,
	)

	return cmd.CombinedOutput()
}
