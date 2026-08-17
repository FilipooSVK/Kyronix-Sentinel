package collector

import (
	"context"

	"kyronix/sentinel/internal/domain"
)

// KernelCollector collects kernel-level health information.
type KernelCollector interface {
	CollectKernel(ctx context.Context) (domain.KernelStats, error)
}
