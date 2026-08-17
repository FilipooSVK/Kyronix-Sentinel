package app

import (
	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/logging"
	"kyronix/sentinel/internal/version"
)

// LogHealthResult writes runtime health diagnostics.
func LogHealthResult(
	logger *logging.Logger,
	result domain.HealthResult,
) {

	logger.Info(
		"health evaluation completed",
		map[string]interface{}{
			"health_score": result.HealthScore,
			"freeze_risk":  result.FreezeRisk,
			"findings":     len(result.Findings),
		},
	)
}

// BuildDiagnostics creates API diagnostics response.
func BuildDiagnostics(
	snapshot domain.Snapshot,
	result domain.HealthResult,
) local.Diagnostics {

	return local.Diagnostics{

		Running: true,

		HealthScore: result.HealthScore,

		FreezeRisk: string(result.FreezeRisk),

		Version: version.Version,

		Collectors: []local.CollectorStatus{

			FromCollectorStatus(
				"Host",
				snapshot.Collection.Host,
			),

			FromCollectorStatus(
				"CPU",
				snapshot.Collection.CPU,
			),

			FromCollectorStatus(
				"Memory",
				snapshot.Collection.Memory,
			),

			FromCollectorStatus(
				"Pressure",
				snapshot.Collection.Pressure,
			),

			FromCollectorStatus(
				"Disk",
				snapshot.Collection.Disk,
			),

			FromCollectorStatus(
				"Kernel",
				snapshot.Collection.Kernel,
			),
		},
	}
}

func FromCollectorStatus(
	name string,
	status domain.CollectorStatus,
) local.CollectorStatus {

	state := "UNKNOWN"

	switch status.State {

	case domain.CollectorOK:

		state = "OK"

	case domain.CollectorError:

		state = "ERROR"

	case domain.CollectorUnavailable:

		state = "UNAVAILABLE"
	}

	return local.CollectorStatus{

		Name: name,

		State: state,

		Message: status.Message,

		CollectionMS: status.CollectionMS,

		LastSuccess: status.LastSuccess,
	}
}
