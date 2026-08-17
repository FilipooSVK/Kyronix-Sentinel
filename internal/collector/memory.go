package collector

import (
	"context"

	"kyronix/sentinel/internal/domain"
)

// MemoryCollector collects physical memory and swap information.
type MemoryCollector interface {
	CollectMemory(ctx context.Context) (domain.MemoryStats, error)
}
