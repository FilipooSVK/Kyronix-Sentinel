package updater

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"kyronix/sentinel/internal/logging"
)

type testReleaseChecker struct {
	check func(
		ctx context.Context,
		currentVersion string,
	) (CheckResult, error)
}

func (c *testReleaseChecker) Check(
	ctx context.Context,
	currentVersion string,
) (CheckResult, error) {

	return c.check(
		ctx,
		currentVersion,
	)
}

func TestMonitorRejectsInvalidInterval(
	t *testing.T,
) {

	var output bytes.Buffer

	logger := logging.New(
		&output,
		"info",
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			return CheckResult{}, nil
		},
	}

	monitor := NewMonitor(
		checker,
		"0.1.1",
		0,
		logger,
	)

	err := monitor.Run(
		context.Background(),
	)

	if err == nil {
		t.Fatal(
			"expected invalid interval error",
		)
	}
}

func TestMonitorChecksImmediately(
	t *testing.T,
) {

	var output bytes.Buffer

	logger := logging.New(
		&output,
		"info",
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			cancel()

			return CheckResult{
				CurrentVersion: "v0.1.1",

				LatestVersion: "v0.1.1",

				UpdateAvailable: false,
			}, nil
		},
	}

	monitor := NewMonitor(
		checker,
		"0.1.1",
		time.Hour,
		logger,
	)

	if err := monitor.Run(
		ctx,
	); err != nil {

		t.Fatal(err)
	}

	if !strings.Contains(
		output.String(),
		"automatic update check complete",
	) {

		t.Fatalf(
			"expected successful update check log, got %s",
			output.String(),
		)
	}
}

func TestMonitorLogsAvailableUpdate(
	t *testing.T,
) {

	var output bytes.Buffer

	logger := logging.New(
		&output,
		"info",
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			cancel()

			return CheckResult{
				CurrentVersion: "v0.1.1",

				LatestVersion: "v0.1.2",

				UpdateAvailable: true,
			}, nil
		},
	}

	monitor := NewMonitor(
		checker,
		"0.1.1",
		time.Hour,
		logger,
	)

	if err := monitor.Run(
		ctx,
	); err != nil {

		t.Fatal(err)
	}

	logOutput := output.String()

	if !strings.Contains(
		logOutput,
		"automatic update available",
	) {

		t.Fatalf(
			"expected update available log, got %s",
			logOutput,
		)
	}

	if !strings.Contains(
		logOutput,
		"v0.1.2",
	) {

		t.Fatalf(
			"expected latest version in log, got %s",
			logOutput,
		)
	}
}

func TestMonitorContinuesAfterCheckFailure(
	t *testing.T,
) {

	var output bytes.Buffer

	logger := logging.New(
		&output,
		"info",
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	var calls atomic.Int32

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			call := calls.Add(
				1,
			)

			if call == 1 {

				return CheckResult{}, errors.New(
					"temporary github failure",
				)
			}

			cancel()

			return CheckResult{
				CurrentVersion: "v0.1.1",

				LatestVersion: "v0.1.1",

				UpdateAvailable: false,
			}, nil
		},
	}

	monitor := NewMonitor(
		checker,
		"0.1.1",
		10*time.Millisecond,
		logger,
	)

	if err := monitor.Run(
		ctx,
	); err != nil {

		t.Fatal(err)
	}

	if calls.Load() < 2 {

		t.Fatalf(
			"expected at least two checks, got %d",
			calls.Load(),
		)
	}

	logOutput := output.String()

	if !strings.Contains(
		logOutput,
		"automatic update check failed",
	) {

		t.Fatalf(
			"expected failed check log, got %s",
			logOutput,
		)
	}

	if !strings.Contains(
		logOutput,
		"automatic update check complete",
	) {

		t.Fatalf(
			"expected recovery check log, got %s",
			logOutput,
		)
	}
}
