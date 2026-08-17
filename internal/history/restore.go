package history

// RestoreSnapshots replaces the current in-memory history with
// previously persisted snapshot entries.
//
// Both snapshot history and HealthResult history are restored so
// all existing consumers see a consistent state.
func (s *Store) RestoreSnapshots(
	entries []SnapshotEntry,
) {

	s.results = nil
	s.snapshots = nil

	for _, entry := range entries {

		s.Add(
			entry.Health,
		)

		s.AddSnapshot(
			entry,
		)
	}
}
