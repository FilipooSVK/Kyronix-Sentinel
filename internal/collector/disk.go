package collector

import (
	"context"

	"kyronix/sentinel/internal/domain"
)

// DiskCollector collects filesystem capacity information.
type DiskCollector interface {
	CollectDisks(ctx context.Context) ([]domain.DiskStats, error)
}
