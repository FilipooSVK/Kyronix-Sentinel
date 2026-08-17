package predictor

import (
	"time"

	"kyronix/sentinel/internal/history"
)

// AnalyzeKernelEventsWindow analyzes kernel incidents that occurred
// inside the requested recent time window.
//
// Events are calculated between consecutive snapshots so repeated
// incidents and counter resets can be detected correctly.
func AnalyzeKernelEventsWindow(
	entries []history.SnapshotEntry,
	window time.Duration,
) KernelEvents {

	if len(entries) < 2 || window <= 0 {
		return KernelEvents{}
	}

	lastTimestamp := entries[len(entries)-1].Timestamp

	cutoff := lastTimestamp.Add(
		-window,
	)

	// Find the first sample inside the window.
	start := 0

	for i, entry := range entries {

		if !entry.Timestamp.Before(cutoff) {
			start = i
			break
		}
	}

	// Keep one sample immediately before the window boundary
	// so the first delta inside the window can still be calculated.
	if start > 0 {
		start--
	}

	var events KernelEvents

	for i := start + 1; i < len(entries); i++ {

		currentEntry := entries[i]

		if currentEntry.Timestamp.Before(cutoff) {
			continue
		}

		previous := entries[i-1].Snapshot.Kernel
		current := currentEntry.Snapshot.Kernel

		events.SystemOOMKills += counterDelta(
			previous.OOM.SystemKillCount,
			current.OOM.SystemKillCount,
		)

		events.CgroupOOMKills += counterDelta(
			previous.OOM.CgroupKillCount,
			current.OOM.CgroupKillCount,
		)

		events.CgroupOOMEvents += counterDelta(
			previous.OOM.CgroupOOMCount,
			current.OOM.CgroupOOMCount,
		)

		events.FilesystemErrors += uintCounterDelta(
			previous.FilesystemErrors,
			current.FilesystemErrors,
		)
	}

	return events
}
