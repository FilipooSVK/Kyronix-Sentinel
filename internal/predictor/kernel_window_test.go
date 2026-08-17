package predictor

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

func TestAnalyzeKernelEventsWindowDetectsRepeatedEvents(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(0),
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				5 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(1),
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				10 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(2),
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				15 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(3),
					},
				},
			},
		},
	}

	events := AnalyzeKernelEventsWindow(
		entries,
		15*time.Minute,
	)

	if events.SystemOOMKills != 3 {

		t.Fatalf(
			"expected 3 system OOM kills, got %d",
			events.SystemOOMKills,
		)
	}
}

func TestAnalyzeKernelEventsWindowIgnoresOldEvents(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(0),
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				10 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(3),
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				50 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(3),
					},
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
					},
				},
			},
		},
	}

	events := AnalyzeKernelEventsWindow(
		entries,
		15*time.Minute,
	)

	if events.SystemOOMKills != 0 {

		t.Fatalf(
			"expected old OOM events to be ignored, got %d",
			events.SystemOOMKills,
		)
	}
}

func TestAnalyzeKernelEventsWindowCounterReset(
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
				},
			},
		},
		{
			Timestamp: start.Add(
				5 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(0),
					},
				},
			},
		},
		{
			Timestamp: start.Add(
				10 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					OOM: domain.OOMStats{
						SystemKillCount: uint64Ptr(2),
					},
				},
			},
		},
	}

	events := AnalyzeKernelEventsWindow(
		entries,
		15*time.Minute,
	)

	if events.SystemOOMKills != 2 {

		t.Fatalf(
			"expected 2 post-reset OOM kills, got %d",
			events.SystemOOMKills,
		)
	}
}

func TestAnalyzeKernelEventsWindowFilesystemErrors(
	t *testing.T,
) {

	start := time.Now()

	entries := []history.SnapshotEntry{
		{
			Timestamp: start,
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					FilesystemErrors: 2,
				},
			},
		},
		{
			Timestamp: start.Add(
				5 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					FilesystemErrors: 4,
				},
			},
		},
		{
			Timestamp: start.Add(
				10 * time.Minute,
			),
			Snapshot: domain.Snapshot{
				Kernel: domain.KernelStats{
					FilesystemErrors: 7,
				},
			},
		},
	}

	events := AnalyzeKernelEventsWindow(
		entries,
		15*time.Minute,
	)

	if events.FilesystemErrors != 5 {

		t.Fatalf(
			"expected 5 filesystem errors, got %d",
			events.FilesystemErrors,
		)
	}
}

func TestAnalyzeKernelEventsWindowInsufficientHistory(
	t *testing.T,
) {

	events := AnalyzeKernelEventsWindow(
		nil,
		15*time.Minute,
	)

	if events.Total() != 0 {

		t.Fatalf(
			"expected no events, got %d",
			events.Total(),
		)
	}
}
