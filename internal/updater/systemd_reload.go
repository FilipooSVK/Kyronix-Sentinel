package updater

import (
	"context"
	"fmt"
	"os/exec"
)

// SystemdReloader reloads systemd unit definitions after a release
// installs or restores Sentinel-managed unit files.
type SystemdReloader interface {
	DaemonReload(
		ctx context.Context,
	) error
}

// SystemdDaemonReloader is the production systemd implementation.
type SystemdDaemonReloader struct{}

func NewSystemdDaemonReloader() *SystemdDaemonReloader {

	return &SystemdDaemonReloader{}
}

func (r *SystemdDaemonReloader) DaemonReload(
	ctx context.Context,
) error {

	command := exec.CommandContext(
		ctx,
		"systemctl",
		"daemon-reload",
	)

	output, err := command.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"systemctl daemon-reload failed: %w: %s",
			err,
			string(output),
		)
	}

	return nil
}
