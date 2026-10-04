package app

import (
	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/version"
)

// evaluate executes one health evaluation cycle
// and updates runtime state.
func (d *Daemon) evaluate() {

	result := d.engine.RunOnce()

	d.lastResult = result

	d.lastSnapshot = d.engine.LastSnapshot()

	// Update current host information.
	d.statusServer.UpdateHost(
		BuildHost(
			d.lastSnapshot,
		),
	)

	// Update predictive runtime state.
	d.UpdatePrediction()

	// Update current health status.
	d.statusServer.Update(
		local.Status{
			Running:     true,
			HealthScore: result.HealthScore,
			FreezeRisk:  string(result.FreezeRisk),
			Version:     version.Version,
		},
	)

	// Update collector diagnostics.
	d.statusServer.UpdateDiagnostics(
		BuildDiagnostics(
			d.lastSnapshot,
			result,
		),
	)

	// Update Prometheus-compatible runtime metrics.
	d.UpdateMetrics()

	LogHealthResult(
		d.logger,
		result,
	)
}
