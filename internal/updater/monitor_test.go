package updater

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
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

func newTestStateStore(
	t *testing.T,
) *StateStore {

	t.Helper()

	return NewStateStore(
		filepath.Join(
			t.TempDir(),
			"update-state.json",
		),
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
		newTestStateStore(t),
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

func TestMonitorChecksImmediatelyAndPersistsState(
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

	store := newTestStateStore(
		t,
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
		store,
	)

	if err := monitor.Run(
		ctx,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.CurrentVersion != "v0.1.1" {

		t.Fatalf(
			"unexpected current version: %s",
			state.CurrentVersion,
		)
	}

	if state.LatestVersion != "v0.1.1" {

		t.Fatalf(
			"unexpected latest version: %s",
			state.LatestVersion,
		)
	}

	if state.UpdateAvailable {

		t.Fatal(
			"unexpected available update",
		)
	}

	if state.LastResult != UpdateResultSuccess {

		t.Fatalf(
			"unexpected result: %s",
			state.LastResult,
		)
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

func TestMonitorPersistsAvailableUpdate(
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

	store := newTestStateStore(
		t,
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
		store,
	)

	if err := monitor.Run(
		ctx,
	); err != nil {

		t.Fatal(err)
	}

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if !state.UpdateAvailable {

		t.Fatal(
			"expected update available state",
		)
	}

	if state.LatestVersion != "v0.1.2" {

		t.Fatalf(
			"unexpected latest version: %s",
			state.LatestVersion,
		)
	}

	if !strings.Contains(
		output.String(),
		"automatic update available",
	) {

		t.Fatalf(
			"expected update available log, got %s",
			output.String(),
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

	store := newTestStateStore(
		t,
	)

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
		store,
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

	state, err := store.Load()

	if err != nil {
		t.Fatal(err)
	}

	if state.LastResult != UpdateResultSuccess {

		t.Fatalf(
			"expected recovered successful state, got %s",
			state.LastResult,
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

func TestMonitorEvaluatesAllowedPolicyInObserveOnlyMode(
	t *testing.T,
) {

	var output bytes.Buffer

	logger := logging.New(
		&output,
		"info",
	)

	ctx, cancel :=
		context.WithCancel(
			context.Background(),
		)

	store := newTestStateStore(
		t,
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

				Release: Release{
					TagName: "v0.1.2",

					PublishedAt: time.Now().
						UTC().
						Add(
							-48 *
								time.Hour,
						),
				},
			}, nil
		},
	}

	monitor := NewMonitor(
		checker,
		"0.1.1",
		time.Hour,
		logger,
		store,
	)

	monitor.SetAutoInstallPolicy(
		AutoInstallPolicy{
			Enabled: true,

			MinReleaseAge: 24 * time.Hour,

			PatchOnly: true,
		},
	)

	if err := monitor.Run(
		ctx,
	); err != nil {

		t.Fatal(err)
	}

	logOutput := output.String()

	if !strings.Contains(
		logOutput,
		"automatic update policy evaluated",
	) {

		t.Fatalf(
			"expected policy evaluation log, got %s",
			logOutput,
		)
	}

	if !strings.Contains(
		logOutput,
		`"decision":"allow"`,
	) {

		t.Fatalf(
			"expected allow decision, got %s",
			logOutput,
		)
	}

	if !strings.Contains(
		logOutput,
		`"mode":"observe_only"`,
	) {

		t.Fatalf(
			"expected observe-only mode, got %s",
			logOutput,
		)
	}
}

func TestMonitorRejectsNegativeAutoInstallReleaseAge(
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
		time.Hour,
		logger,
		newTestStateStore(t),
	)

	monitor.SetAutoInstallPolicy(
		AutoInstallPolicy{
			Enabled: true,

			MinReleaseAge: -time.Hour,

			PatchOnly: true,
		},
	)

	err := monitor.Run(
		context.Background(),
	)

	if err == nil {

		t.Fatal(
			"expected invalid policy error",
		)
	}
}

type testUpdateWorker struct {
	starts atomic.Int32

	err error
}

func (w *testUpdateWorker) Start(
	ctx context.Context,
) error {

	w.starts.Add(
		1,
	)

	return w.err
}

func allowedAutoInstallCheckResult() CheckResult {

	return CheckResult{
		CurrentVersion: "v0.1.1",

		LatestVersion: "v0.1.2",

		UpdateAvailable: true,

		Release: Release{
			TagName: "v0.1.2",

			PublishedAt: time.Now().
				UTC().
				Add(
					-48 * time.Hour,
				),
		},
	}
}

func newAutoInstallExecutionMonitor(
	t *testing.T,
	output *bytes.Buffer,
	policy AutoInstallPolicy,
) *Monitor {

	t.Helper()

	logger := logging.New(
		output,
		"info",
	)

	checker := &testReleaseChecker{
		check: func(
			ctx context.Context,
			currentVersion string,
		) (CheckResult, error) {

			return allowedAutoInstallCheckResult(),
				nil
		},
	}

	monitor := NewMonitor(
		checker,
		"0.1.1",
		time.Hour,
		logger,
		newTestStateStore(t),
	)

	monitor.SetAutoInstallPolicy(
		policy,
	)

	return monitor
}

func TestMonitorObserveOnlyNeverStartsWorker(
	t *testing.T,
) {

	var output bytes.Buffer

	monitor :=
		newAutoInstallExecutionMonitor(
			t,
			&output,
			AutoInstallPolicy{
				Enabled: true,

				MinReleaseAge: 24 * time.Hour,

				PatchOnly: true,
			},
		)

	worker := &testUpdateWorker{}

	if err := monitor.SetAutoInstallExecution(
		AutoInstallExecutionObserveOnly,
		worker,
	); err != nil {

		t.Fatal(err)
	}

	monitor.checkOnce(
		context.Background(),
	)

	if worker.starts.Load() != 0 {

		t.Fatalf(
			"observe-only mode unexpectedly started worker %d time(s)",
			worker.starts.Load(),
		)
	}

	if !strings.Contains(
		output.String(),
		`"mode":"observe_only"`,
	) {

		t.Fatalf(
			"expected observe-only log, got %s",
			output.String(),
		)
	}
}

func TestMonitorWorkerEnabledStartsWorkerOnAllow(
	t *testing.T,
) {

	var output bytes.Buffer

	monitor :=
		newAutoInstallExecutionMonitor(
			t,
			&output,
			AutoInstallPolicy{
				Enabled: true,

				MinReleaseAge: 24 * time.Hour,

				PatchOnly: true,
			},
		)

	worker := &testUpdateWorker{}

	if err := monitor.SetAutoInstallExecution(
		AutoInstallExecutionWorkerEnabled,
		worker,
	); err != nil {

		t.Fatal(err)
	}

	monitor.checkOnce(
		context.Background(),
	)

	if worker.starts.Load() != 1 {

		t.Fatalf(
			"expected worker to start once, got %d",
			worker.starts.Load(),
		)
	}

	if !strings.Contains(
		output.String(),
		"automatic update worker start requested",
	) {

		t.Fatalf(
			"expected worker start log, got %s",
			output.String(),
		)
	}

	if !strings.Contains(
		output.String(),
		`"mode":"worker_enabled"`,
	) {

		t.Fatalf(
			"expected worker-enabled mode log, got %s",
			output.String(),
		)
	}
}

func TestMonitorWorkerEnabledDoesNotStartWorkerOnDeny(
	t *testing.T,
) {

	var output bytes.Buffer

	monitor :=
		newAutoInstallExecutionMonitor(
			t,
			&output,
			AutoInstallPolicy{
				Enabled: false,

				MinReleaseAge: 24 * time.Hour,

				PatchOnly: true,
			},
		)

	worker := &testUpdateWorker{}

	if err := monitor.SetAutoInstallExecution(
		AutoInstallExecutionWorkerEnabled,
		worker,
	); err != nil {

		t.Fatal(err)
	}

	monitor.checkOnce(
		context.Background(),
	)

	if worker.starts.Load() != 0 {

		t.Fatalf(
			"denied policy unexpectedly started worker %d time(s)",
			worker.starts.Load(),
		)
	}

	if !strings.Contains(
		output.String(),
		`"decision":"deny"`,
	) {

		t.Fatalf(
			"expected deny decision, got %s",
			output.String(),
		)
	}
}

func TestMonitorLogsWorkerStartFailure(
	t *testing.T,
) {

	var output bytes.Buffer

	monitor :=
		newAutoInstallExecutionMonitor(
			t,
			&output,
			AutoInstallPolicy{
				Enabled: true,

				MinReleaseAge: 24 * time.Hour,

				PatchOnly: true,
			},
		)

	worker := &testUpdateWorker{
		err: errors.New(
			"worker launch failed",
		),
	}

	if err := monitor.SetAutoInstallExecution(
		AutoInstallExecutionWorkerEnabled,
		worker,
	); err != nil {

		t.Fatal(err)
	}

	monitor.checkOnce(
		context.Background(),
	)

	if worker.starts.Load() != 1 {

		t.Fatalf(
			"expected one worker start attempt, got %d",
			worker.starts.Load(),
		)
	}

	if !strings.Contains(
		output.String(),
		"automatic update worker start failed",
	) {

		t.Fatalf(
			"expected worker failure log, got %s",
			output.String(),
		)
	}
}

func TestMonitorWorkerEnabledRequiresWorker(
	t *testing.T,
) {

	var output bytes.Buffer

	monitor :=
		newAutoInstallExecutionMonitor(
			t,
			&output,
			AutoInstallPolicy{},
		)

	err := monitor.SetAutoInstallExecution(
		AutoInstallExecutionWorkerEnabled,
		nil,
	)

	if err == nil {

		t.Fatal(
			"expected worker-enabled mode to require worker",
		)
	}
}

func TestMonitorRejectsInvalidExecutionMode(
	t *testing.T,
) {

	var output bytes.Buffer

	monitor :=
		newAutoInstallExecutionMonitor(
			t,
			&output,
			AutoInstallPolicy{},
		)

	err := monitor.SetAutoInstallExecution(
		AutoInstallExecutionMode(
			"invalid",
		),
		nil,
	)

	if err == nil {

		t.Fatal(
			"expected invalid execution mode error",
		)
	}
}
