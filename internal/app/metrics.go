package app

import (
	"time"

	"kyronix/sentinel/internal/domain"
	sentinelmetrics "kyronix/sentinel/internal/metrics"
)

// UpdateMetrics updates the Prometheus-compatible runtime metrics state.
func (d *Daemon) UpdateMetrics() {
	if d.metricsServer == nil {
		return
	}

	prediction := d.engine.LastPrediction()

	snapshot := d.lastSnapshot

	result := d.lastResult

	d.metricsServer.Update(
		sentinelmetrics.State{
			Up: true,

			HealthScore: result.HealthScore,

			FreezeRisk: string(
				result.FreezeRisk,
			),

			PredictionScore: prediction.Score,

			PredictionConfidence: prediction.Confidence,

			ActiveSignals: prediction.ActiveSignals,

			PersistentSignals: prediction.PersistentSignals,

			KernelEvidence: prediction.KernelEvidence,

			Host: sentinelmetrics.HostInfo{
				Hostname: snapshot.Host.Hostname,

				Environment: snapshot.Identity.Environment,

				Role: snapshot.Identity.Role,

				Architecture: snapshot.Host.Architecture,

				Virtualization: snapshot.Host.Virtualization,

				Platform: snapshot.Host.Platform,

				OSID: snapshot.Host.OSID,

				OSVersion: snapshot.Host.OSVersion,
			},

			Collectors: []sentinelmetrics.Collector{
				buildCollectorMetric(
					"host",
					snapshot.Collection.Host,
				),

				buildCollectorMetric(
					"cpu",
					snapshot.Collection.CPU,
				),

				buildCollectorMetric(
					"memory",
					snapshot.Collection.Memory,
				),

				buildCollectorMetric(
					"pressure",
					snapshot.Collection.Pressure,
				),

				buildCollectorMetric(
					"disk",
					snapshot.Collection.Disk,
				),

				buildCollectorMetric(
					"kernel",
					snapshot.Collection.Kernel,
				),
			},
		},
	)
}

func buildCollectorMetric(
	name string,
	status domain.CollectorStatus,
) sentinelmetrics.Collector {
	return sentinelmetrics.Collector{
		Name: name,

		Up: status.State == domain.CollectorOK,

		CollectionDuration: time.Duration(
			status.CollectionMS,
		) * time.Millisecond,
	}
}
