package collector

import (
	"context"

	"kyronix/sentinel/internal/domain"
)

// PressureCollector collects Linux Pressure Stall Information.
type PressureCollector interface {
	CollectPressure(ctx context.Context) (domain.PressureStats, error)
}
