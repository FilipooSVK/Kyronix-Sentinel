package app

import (
	"context"

	"kyronix/sentinel/internal/analyzer"
	"kyronix/sentinel/internal/collector"
	"kyronix/sentinel/internal/collector/linux"
	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

// NewDefaultEngine creates production Sentinel engine
// using provided configuration.
func NewDefaultEngine(
	cfg config.Config,
) *Engine {

	snapshotManager := collector.NewSnapshotManager(
		linux.NewHostCollector(),
		linux.NewCPUCollector(),
		linux.NewMemoryCollector(),
		linux.NewPressureCollector(),
		linux.NewDiskCollector(),
		linux.NewKernelCollector(),
	)

	return NewEngine(
		snapshotManagerAdapter{
			manager: snapshotManager,
		},
		analyzer.NewDefaultAnalyzer(),
		history.NewStore(
			cfg.History.Size,
		),
	)
}

// snapshotManagerAdapter adapts SnapshotManager
// to Engine SnapshotCollector interface.
type snapshotManagerAdapter struct {
	manager *collector.SnapshotManager
}

// Collect executes snapshot collection.
func (s snapshotManagerAdapter) Collect() domain.Snapshot {

	return s.manager.Collect(
		context.Background(),
	)
}
