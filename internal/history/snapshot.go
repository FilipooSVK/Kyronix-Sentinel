package history

import (
	"time"

	"kyronix/sentinel/internal/domain"
)

// SnapshotEntry represents historical system state.
type SnapshotEntry struct {
	Timestamp time.Time

	Snapshot domain.Snapshot

	Health domain.HealthResult
}
