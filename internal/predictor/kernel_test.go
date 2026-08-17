package predictor

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

func uint64Ptr(
	value uint64,
) *uint64 {

	return &value
}

func TestAnalyzeKernelEventsDetectsNewIncidents(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(2),
						CgroupKillCount: uint64Ptr(4),
						CgroupOOMCount:  uint64Ptr(5),
					},
					FilesystemErrors: 1,
				},
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(3),
						CgroupKillCount: uint64Ptr(6),
						CgroupOOMCount:  uint64Ptr(8),
					},
					FilesystemErrors: 5,
				},
			},
		},
	}

	events := AnalyzeKernelEvents(
		entries,
	)

	if events.SystemOOMKills != 1 {

		t.Fatalf(
			"expected 1 system OOM kill, got %d",
			events.SystemOOMKills,
		)
	}

	if events.CgroupOOMKills != 2 {

		t.Fatalf(
			"expected 2 cgroup OOM kills, got %d",
			events.CgroupOOMKills,
		)
	}

	if events.CgroupOOMEvents != 3 {

		t.Fatalf(
			"expected 3 cgroup OOM events, got %d",
			events.CgroupOOMEvents,
		)
	}

	if events.FilesystemErrors != 4 {

		t.Fatalf(
			"expected 4 filesystem errors, got %d",
			events.FilesystemErrors,
		)
	}

	if events.Total() != 10 {

		t.Fatalf(
			"expected 10 total events, got %d",
			events.Total(),
		)
	}

	if !events.HasEvents() {

		t.Fatal(
			"expected kernel events to be detected",
		)
	}
}

func TestAnalyzeKernelEventsStableCounters(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(2),
						CgroupKillCount: uint64Ptr(3),
						CgroupOOMCount:  uint64Ptr(4),
					},
					FilesystemErrors: 5,
				},
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(2),
						CgroupKillCount: uint64Ptr(3),
						CgroupOOMCount:  uint64Ptr(4),
					},
					FilesystemErrors: 5,
				},
			},
		},
	}

	events := AnalyzeKernelEvents(
		entries,
	)

	if events.Total() != 0 {

		t.Fatalf(
			"expected no new events, got %d",
			events.Total(),
		)
	}

	if events.HasEvents() {

		t.Fatal(
			"expected no kernel events",
		)
	}
}

func TestAnalyzeKernelEventsUnavailableOOMCounters(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					FilesystemErrors: 0,
				},
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					FilesystemErrors: 0,
				},
			},
		},
	}

	events := AnalyzeKernelEvents(
		entries,
	)

	if events.Total() != 0 {

		t.Fatalf(
			"expected no events for unavailable counters, got %d",
			events.Total(),
		)
	}
}

func TestAnalyzeKernelEventsCounterReset(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(7),
					},
					FilesystemErrors: 9,
				},
			},
		},
		{
			Timestamp: start.Add(
				time.Hour,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(1),
					},
					FilesystemErrors: 2,
				},
			},
		},
	}

	events := AnalyzeKernelEvents(
		entries,
	)

	if events.SystemOOMKills != 1 {

		t.Fatalf(
			"expected reset system OOM delta 1, got %d",
			events.SystemOOMKills,
		)
	}

	if events.FilesystemErrors != 2 {

		t.Fatalf(
			"expected reset filesystem delta 2, got %d",
			events.FilesystemErrors,
		)
	}
}

func TestAnalyzeKernelEventsInsufficientHistory(
	t *testing.T,
) {

	events := AnalyzeKernelEvents(
		nil,
	)

	if events.Total() != 0 {

		t.Fatalf(
			"expected no events, got %d",
			events.Total(),
		)
	}
}
