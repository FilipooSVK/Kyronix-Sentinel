package collector

import (
	"context"

	"kyronix/sentinel/internal/domain"
)

// HostCollector collects general host information.
type HostCollector interface {
	CollectHost(ctx context.Context) (domain.HostStats, error)
}
