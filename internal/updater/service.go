package updater

import (
	"context"
	"fmt"
	"os/exec"
)

// ServiceController controls the Sentinel daemon lifecycle
// during software updates.
type ServiceController interface {
	Restart(
		ctx context.Context,
	) error
}

// SystemdController controls sentineld through systemd.
type SystemdController struct {
	Service string
}

// NewSystemdController creates a production service controller.
func NewSystemdController(
	service string,
) *SystemdController {

	return &SystemdController{
		Service: service,
	}
}

// Restart restarts the configured systemd service.
func (c *SystemdController) Restart(
	ctx context.Context,
) error {

	if c.Service == "" {

		return fmt.Errorf(
			"systemd service name is empty",
		)
	}

	command := exec.CommandContext(
		ctx,
		"systemctl",
		"restart",
		c.Service,
	)

	output, err := command.CombinedOutput()

	if err != nil {

		return fmt.Errorf(
			"systemctl restart %s failed: %w: %s",
			c.Service,
			err,
			string(output),
		)
	}

	return nil
}
