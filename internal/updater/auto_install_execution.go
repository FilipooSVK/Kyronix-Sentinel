package updater

import "fmt"

// AutoInstallExecutionMode controls what happens after an automatic
// installation policy decision.
//
// The zero-value behavior of Monitor remains observe-only through
// NewMonitor, so merely configuring an UpdateWorker cannot enable
// unattended installation.
type AutoInstallExecutionMode string

const (
	AutoInstallExecutionObserveOnly AutoInstallExecutionMode = "observe_only"

	AutoInstallExecutionWorkerEnabled AutoInstallExecutionMode = "worker_enabled"
)

// Valid reports whether the execution mode is supported.
func (m AutoInstallExecutionMode) Valid() bool {

	switch m {

	case AutoInstallExecutionObserveOnly,
		AutoInstallExecutionWorkerEnabled:

		return true

	default:

		return false
	}
}

// SetAutoInstallExecution configures what the update monitor may do
// after an ALLOW policy decision.
//
// worker_enabled always requires an explicit worker dependency.
// observe_only deliberately discards any supplied worker.
func (m *Monitor) SetAutoInstallExecution(
	mode AutoInstallExecutionMode,
	worker UpdateWorker,
) error {

	if m == nil {

		return fmt.Errorf(
			"update monitor is nil",
		)
	}

	if !mode.Valid() {

		return fmt.Errorf(
			"invalid automatic install execution mode: %q",
			mode,
		)
	}

	if mode ==
		AutoInstallExecutionWorkerEnabled &&
		worker == nil {

		return fmt.Errorf(
			"automatic install worker is nil",
		)
	}

	m.autoInstallMode = mode

	if mode ==
		AutoInstallExecutionObserveOnly {

		m.updateWorker = nil

		return nil
	}

	m.updateWorker = worker

	return nil
}
