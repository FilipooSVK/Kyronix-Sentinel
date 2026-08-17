package predictor

import "kyronix/sentinel/internal/history"

// KernelEvents represents new kernel-level incidents
// detected during the analyzed history window.
type KernelEvents struct {
	SystemOOMKills uint64

	CgroupOOMKills uint64

	CgroupOOMEvents uint64

	FilesystemErrors uint64
}

// Total returns total number of newly detected kernel incidents.
func (e KernelEvents) Total() uint64 {

	return e.SystemOOMKills +
		e.CgroupOOMKills +
		e.CgroupOOMEvents +
		e.FilesystemErrors
}

// HasEvents reports whether any new kernel incident was detected.
func (e KernelEvents) HasEvents() bool {

	return e.Total() > 0
}

// AnalyzeKernelEvents compares the oldest and newest snapshot
// and returns newly observed kernel incidents.
func AnalyzeKernelEvents(
	entries []history.SnapshotEntry,
) KernelEvents {

	if len(entries) < 2 {

		return KernelEvents{}
	}

	first := entries[0].Snapshot.Kernel

	last := entries[len(entries)-1].Snapshot.Kernel

	return KernelEvents{
		SystemOOMKills: counterDelta(
			first.OOM.SystemKillCount,
			last.OOM.SystemKillCount,
		),

		CgroupOOMKills: counterDelta(
			first.OOM.CgroupKillCount,
			last.OOM.CgroupKillCount,
		),

		CgroupOOMEvents: counterDelta(
			first.OOM.CgroupOOMCount,
			last.OOM.CgroupOOMCount,
		),

		FilesystemErrors: uintCounterDelta(
			first.FilesystemErrors,
			last.FilesystemErrors,
		),
	}
}

// counterDelta calculates the increase between two optional counters.
//
// Nil counters represent unavailable data and therefore do not
// generate predictive events.
func counterDelta(
	previous *uint64,
	current *uint64,
) uint64 {

	if previous == nil || current == nil {

		return 0
	}

	return uintCounterDelta(
		*previous,
		*current,
	)
}

// uintCounterDelta calculates the increase of a monotonic counter.
//
// If the counter decreases, for example because of a host reboot,
// the new value is treated as the beginning of a new counter epoch.
func uintCounterDelta(
	previous uint64,
	current uint64,
) uint64 {

	if current >= previous {

		return current - previous
	}

	return current
}
