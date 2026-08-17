package updater

import (
	"context"
	"fmt"
	"time"

	"kyronix/sentinel/internal/api/local"
)

const (
	defaultHealthTimeout  = 15 * time.Second
	defaultHealthInterval = 500 * time.Millisecond
)

// VersionHealthChecker verifies that Sentinel is running
// and reports the expected software version.
type VersionHealthChecker interface {
	WaitForVersion(
		ctx context.Context,
		version string,
	) error
}

// UnixHealthChecker performs update health checks through
// the Sentinel local Unix socket API.
type UnixHealthChecker struct {
	Socket string

	Timeout time.Duration

	Interval time.Duration
}

// NewUnixHealthChecker creates the production Sentinel health checker.
func NewUnixHealthChecker(
	socket string,
) *UnixHealthChecker {

	return &UnixHealthChecker{
		Socket: socket,

		Timeout: defaultHealthTimeout,

		Interval: defaultHealthInterval,
	}
}

// WaitForVersion waits until Sentinel reports Running=true
// and the expected version through its Unix API.
func (c *UnixHealthChecker) WaitForVersion(
	ctx context.Context,
	expectedVersion string,
) error {

	if !IsValidVersion(
		expectedVersion,
	) {

		return fmt.Errorf(
			"invalid expected version: %s",
			expectedVersion,
		)
	}

	timeout := c.Timeout

	if timeout <= 0 {
		timeout = defaultHealthTimeout
	}

	interval := c.Interval

	if interval <= 0 {
		interval = defaultHealthInterval
	}

	checkCtx, cancel := context.WithTimeout(
		ctx,
		timeout,
	)

	defer cancel()

	ticker := time.NewTicker(
		interval,
	)

	defer ticker.Stop()

	for {

		status, err := local.GetStatus(
			c.Socket,
		)

		if err == nil &&
			status.Running &&
			IsValidVersion(status.Version) &&
			CompareVersions(
				status.Version,
				expectedVersion,
			) == 0 {

			return nil
		}

		select {

		case <-checkCtx.Done():

			if err != nil {

				return fmt.Errorf(
					"Sentinel health check failed: %w",
					err,
				)
			}

			return fmt.Errorf(
				"Sentinel did not report expected version %s before timeout",
				expectedVersion,
			)

		case <-ticker.C:
		}
	}
}
