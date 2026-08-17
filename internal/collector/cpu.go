package collector

import (
	"context"

	"kyronix/sentinel/internal/domain"
)

// CPUCollector collects CPU utilization and load information.
type CPUCollector interface {
	CollectCPU(ctx context.Context) (domain.CPUStats, error)
}
