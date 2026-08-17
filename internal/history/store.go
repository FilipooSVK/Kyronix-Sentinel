package history

import (
	"kyronix/sentinel/internal/domain"
)

// Store keeps historical Sentinel data.
type Store struct {
	results []domain.HealthResult

	snapshots []SnapshotEntry

	limit int
}

// NewStore creates history storage.
func NewStore(
	limit int,
) *Store {

	return &Store{

		results: make(
			[]domain.HealthResult,
			0,
		),

		snapshots: make(
			[]SnapshotEntry,
			0,
		),

		limit: limit,
	}
}

// Add stores a health result.
func (s *Store) Add(
	result domain.HealthResult,
) {

	s.results = append(
		s.results,
		result,
	)

	if len(s.results) > s.limit {

		s.results = s.results[len(s.results)-s.limit:]
	}
}

// AddSnapshot stores a historical snapshot.
func (s *Store) AddSnapshot(
	entry SnapshotEntry,
) {

	s.snapshots = append(
		s.snapshots,
		entry,
	)

	if len(s.snapshots) > s.limit {

		s.snapshots = s.snapshots[len(s.snapshots)-s.limit:]
	}
}

// GetAll returns stored health results.
func (s *Store) GetAll() []domain.HealthResult {

	return append(
		[]domain.HealthResult{},
		s.results...,
	)
}

// GetSnapshots returns historical snapshots.
func (s *Store) GetSnapshots() []SnapshotEntry {

	return append(
		[]SnapshotEntry{},
		s.snapshots...,
	)
}

// Latest returns newest result.
func (s *Store) Latest() (
	domain.HealthResult,
	bool,
) {

	if len(s.results) == 0 {

		return domain.HealthResult{}, false
	}

	return s.results[len(s.results)-1], true
}

// LatestSnapshot returns newest snapshot.
func (s *Store) LatestSnapshot() (
	SnapshotEntry,
	bool,
) {

	if len(s.snapshots) == 0 {

		return SnapshotEntry{}, false
	}

	return s.snapshots[len(s.snapshots)-1], true
}
