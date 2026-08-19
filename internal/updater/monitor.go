package updater

import (
	"context"
	"fmt"
	"time"

	"kyronix/sentinel/internal/logging"
)

// ReleaseChecker checks whether a newer Sentinel release exists.
type ReleaseChecker interface {
	Check(
		ctx context.Context,
		currentVersion string,
	) (CheckResult, error)
}

// Monitor periodically checks for Sentinel software updates.
type Monitor struct {
	checker ReleaseChecker

	currentVersion string

	interval time.Duration

	logger *logging.Logger

	stateStore *StateStore

	autoInstallPolicy AutoInstallPolicy

	autoInstallMode AutoInstallExecutionMode

	updateWorker UpdateWorker
}

// NewMonitor creates a periodic Sentinel update monitor.
func NewMonitor(
	checker ReleaseChecker,
	currentVersion string,
	interval time.Duration,
	logger *logging.Logger,
	stateStore *StateStore,
) *Monitor {

	return &Monitor{
		checker: checker,

		currentVersion: currentVersion,

		interval: interval,

		logger: logger,

		stateStore: stateStore,

		autoInstallMode: AutoInstallExecutionObserveOnly,
	}
}

// SetAutoInstallPolicy configures unattended-install eligibility.
//
// The monitor currently evaluates this policy in observe-only mode.
// It does not download or install releases.
func (m *Monitor) SetAutoInstallPolicy(
	policy AutoInstallPolicy,
) {

	if m == nil {
		return
	}

	m.autoInstallPolicy = policy
}

// Run starts the periodic update check loop.
//
// An initial check is performed immediately. Subsequent checks
// run according to the configured interval.
//
// Update check failures are logged but do not stop the monitor.
func (m *Monitor) Run(
	ctx context.Context,
) error {

	if m.checker == nil {

		return fmt.Errorf(
			"update checker is nil",
		)
	}

	if m.logger == nil {

		return fmt.Errorf(
			"update monitor logger is nil",
		)
	}

	if m.stateStore == nil {

		return fmt.Errorf(
			"update state store is nil",
		)
	}

	if !IsValidVersion(
		m.currentVersion,
	) {

		return fmt.Errorf(
			"invalid current version: %s",
			m.currentVersion,
		)
	}

	if m.interval <= 0 {

		return fmt.Errorf(
			"update check interval must be greater than zero",
		)
	}

	if m.autoInstallPolicy.MinReleaseAge < 0 {

		return fmt.Errorf(
			"automatic install minimum release age cannot be negative",
		)
	}

	if !m.autoInstallMode.Valid() {

		return fmt.Errorf(
			"invalid automatic install execution mode: %q",
			m.autoInstallMode,
		)
	}

	if m.autoInstallMode ==
		AutoInstallExecutionWorkerEnabled &&
		m.updateWorker == nil {

		return fmt.Errorf(
			"automatic install worker is nil",
		)
	}

	m.checkOnce(
		ctx,
	)

	ticker := time.NewTicker(
		m.interval,
	)

	defer ticker.Stop()

	for {

		select {

		case <-ctx.Done():

			return nil

		case <-ticker.C:

			m.checkOnce(
				ctx,
			)
		}
	}
}

func (m *Monitor) checkOnce(
	ctx context.Context,
) {

	if ctx.Err() != nil {
		return
	}

	result, err := m.checker.Check(
		ctx,
		m.currentVersion,
	)

	checkedAt := time.Now().UTC()

	nextCheck := checkedAt.Add(
		m.interval,
	)

	if err != nil {

		if ctx.Err() != nil {
			return
		}

		state := UpdateState{
			LastCheck: checkedAt,

			NextCheck: nextCheck,

			CurrentVersion: NormalizeVersion(
				m.currentVersion,
			),

			LastResult: UpdateResultError,

			LastError: err.Error(),
		}

		previous, loadErr := m.stateStore.Load()

		if loadErr == nil {

			state.LatestVersion =
				previous.LatestVersion

			state.UpdateAvailable =
				previous.UpdateAvailable
		}

		m.persistState(
			state,
		)

		m.logger.Error(
			"automatic update check failed",
			map[string]interface{}{
				"current_version": state.CurrentVersion,
				"error":           err.Error(),
			},
		)

		return
	}

	state := UpdateState{
		LastCheck: checkedAt,

		NextCheck: nextCheck,

		CurrentVersion: result.CurrentVersion,

		LatestVersion: result.LatestVersion,

		UpdateAvailable: result.UpdateAvailable,

		LastResult: UpdateResultSuccess,
	}

	m.persistState(
		state,
	)

	if result.UpdateAvailable {

		m.logger.Info(
			"automatic update available",
			map[string]interface{}{
				"current_version": result.CurrentVersion,
				"latest_version":  result.LatestVersion,
			},
		)

		m.evaluateAutoInstallPolicy(
			ctx,
			result,
			checkedAt,
		)

		return
	}

	m.logger.Info(
		"automatic update check complete",
		map[string]interface{}{
			"current_version": result.CurrentVersion,
			"latest_version":  result.LatestVersion,
			"status":          "up_to_date",
		},
	)
}

func (m *Monitor) persistState(
	state UpdateState,
) {

	if err := m.stateStore.SaveCheck(
		state,
	); err != nil {

		m.logger.Error(
			"update state persistence failed",
			map[string]interface{}{
				"path":  m.stateStore.Path(),
				"error": err.Error(),
			},
		)
	}
}

func (m *Monitor) evaluateAutoInstallPolicy(
	ctx context.Context,
	result CheckResult,
	now time.Time,
) {

	state, err := m.stateStore.Load()

	if err != nil {

		m.logger.Error(
			"automatic update policy evaluation failed",
			map[string]interface{}{
				"current_version": result.CurrentVersion,
				"latest_version":  result.LatestVersion,
				"mode":            string(m.autoInstallMode),
				"error":           err.Error(),
			},
		)

		return
	}

	decision, err :=
		EvaluateAutoInstallPolicy(
			m.autoInstallPolicy,
			AutoInstallPolicyInput{
				Check: result,

				State: state,

				Now: now.UTC(),
			},
		)

	if err != nil {

		m.logger.Error(
			"automatic update policy evaluation failed",
			map[string]interface{}{
				"current_version": result.CurrentVersion,
				"latest_version":  result.LatestVersion,
				"mode":            string(m.autoInstallMode),
				"error":           err.Error(),
			},
		)

		return
	}

	decisionName := "deny"

	if decision.Allowed {
		decisionName = "allow"
	}

	fields := map[string]interface{}{
		"current_version": result.CurrentVersion,

		"latest_version": result.LatestVersion,

		"decision": decisionName,

		"reasons": decision.Reasons,

		"mode": string(m.autoInstallMode),

		"auto_install": m.autoInstallPolicy.Enabled,

		"patch_only": m.autoInstallPolicy.PatchOnly,

		"min_release_age": m.autoInstallPolicy.MinReleaseAge.String(),
	}

	if decision.ReleaseAge > 0 {

		fields["release_age"] =
			decision.ReleaseAge.String()
	}

	m.logger.Info(
		"automatic update policy evaluated",
		fields,
	)

	if !decision.Allowed {

		return
	}

	if m.autoInstallMode !=
		AutoInstallExecutionWorkerEnabled {

		return
	}

	if ctx.Err() != nil {

		return
	}

	if m.updateWorker == nil {

		m.logger.Error(
			"automatic update worker start failed",
			map[string]interface{}{
				"current_version": result.CurrentVersion,

				"latest_version": result.LatestVersion,

				"mode": string(m.autoInstallMode),

				"error": "automatic install worker is nil",
			},
		)

		return
	}

	if err := m.updateWorker.Start(
		ctx,
	); err != nil {

		m.logger.Error(
			"automatic update worker start failed",
			map[string]interface{}{
				"current_version": result.CurrentVersion,

				"latest_version": result.LatestVersion,

				"mode": string(m.autoInstallMode),

				"error": err.Error(),
			},
		)

		return
	}

	m.logger.Info(
		"automatic update worker start requested",
		map[string]interface{}{
			"current_version": result.CurrentVersion,

			"latest_version": result.LatestVersion,

			"mode": string(m.autoInstallMode),
		},
	)
}
