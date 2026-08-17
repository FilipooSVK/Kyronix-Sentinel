package collector

import (
	"context"
	"time"

	"kyronix/sentinel/internal/domain"
)

// SnapshotManager executes all collectors and builds a system snapshot.
type SnapshotManager struct {
	Host     HostCollector
	CPU      CPUCollector
	Memory   MemoryCollector
	Pressure PressureCollector
	Disk     DiskCollector
	Kernel   KernelCollector
}

// NewSnapshotManager creates a snapshot manager.
func NewSnapshotManager(
	host HostCollector,
	cpu CPUCollector,
	memory MemoryCollector,
	pressure PressureCollector,
	disk DiskCollector,
	kernel KernelCollector,
) *SnapshotManager {
	return &SnapshotManager{
		Host:     host,
		CPU:      cpu,
		Memory:   memory,
		Pressure: pressure,
		Disk:     disk,
		Kernel:   kernel,
	}
}

// Collect executes all collectors and returns one system snapshot.
//
// Individual collector failures do not invalidate the complete snapshot.
// The failure is recorded in CollectionStatus.
func (m *SnapshotManager) Collect(
	ctx context.Context,
) domain.Snapshot {
	snapshot := domain.Snapshot{
		Timestamp: time.Now(),
	}

	collectHost(ctx, m.Host, &snapshot)
	collectCPU(ctx, m.CPU, &snapshot)
	collectMemory(ctx, m.Memory, &snapshot)
	collectPressure(ctx, m.Pressure, &snapshot)
	collectDisk(ctx, m.Disk, &snapshot)
	collectKernel(ctx, m.Kernel, &snapshot)

	return snapshot
}

func collectHost(
	ctx context.Context,
	collector HostCollector,
	snapshot *domain.Snapshot,
) {
	start := time.Now()

	value, err := collector.CollectHost(ctx)

	status := domain.CollectorStatus{
		CollectionMS: time.Since(start).Milliseconds(),
	}

	if err != nil {
		status.State = domain.CollectorError
		status.Message = err.Error()
		snapshot.Collection.Host = status
		return
	}

	status.State = domain.CollectorOK
	now := time.Now()
	status.LastSuccess = &now

	snapshot.Host = value
	snapshot.Collection.Host = status
}

func collectCPU(
	ctx context.Context,
	collector CPUCollector,
	snapshot *domain.Snapshot,
) {
	start := time.Now()

	value, err := collector.CollectCPU(ctx)

	status := domain.CollectorStatus{
		CollectionMS: time.Since(start).Milliseconds(),
	}

	if err != nil {
		status.State = domain.CollectorError
		status.Message = err.Error()
		snapshot.Collection.CPU = status
		return
	}

	status.State = domain.CollectorOK
	now := time.Now()
	status.LastSuccess = &now

	snapshot.CPU = value
	snapshot.Collection.CPU = status
}

func collectMemory(
	ctx context.Context,
	collector MemoryCollector,
	snapshot *domain.Snapshot,
) {
	start := time.Now()

	value, err := collector.CollectMemory(ctx)

	status := domain.CollectorStatus{
		CollectionMS: time.Since(start).Milliseconds(),
	}

	if err != nil {
		status.State = domain.CollectorError
		status.Message = err.Error()
		snapshot.Collection.Memory = status
		return
	}

	status.State = domain.CollectorOK
	now := time.Now()
	status.LastSuccess = &now

	snapshot.Memory = value
	snapshot.Collection.Memory = status
}

func collectPressure(
	ctx context.Context,
	collector PressureCollector,
	snapshot *domain.Snapshot,
) {
	start := time.Now()

	value, err := collector.CollectPressure(ctx)

	status := domain.CollectorStatus{
		CollectionMS: time.Since(start).Milliseconds(),
	}

	if err != nil {
		status.State = domain.CollectorUnavailable
		status.Message = err.Error()
		snapshot.Collection.Pressure = status
		return
	}

	status.State = domain.CollectorOK
	now := time.Now()
	status.LastSuccess = &now

	snapshot.Pressure = value
	snapshot.Collection.Pressure = status
}

func collectDisk(
	ctx context.Context,
	collector DiskCollector,
	snapshot *domain.Snapshot,
) {
	start := time.Now()

	value, err := collector.CollectDisks(ctx)

	status := domain.CollectorStatus{
		CollectionMS: time.Since(start).Milliseconds(),
	}

	if err != nil {
		status.State = domain.CollectorError
		status.Message = err.Error()
		snapshot.Collection.Disk = status
		return
	}

	status.State = domain.CollectorOK
	now := time.Now()
	status.LastSuccess = &now

	snapshot.Disks = value
	snapshot.Collection.Disk = status
}

func collectKernel(
	ctx context.Context,
	collector KernelCollector,
	snapshot *domain.Snapshot,
) {
	start := time.Now()

	value, err := collector.CollectKernel(ctx)

	status := domain.CollectorStatus{
		CollectionMS: time.Since(start).Milliseconds(),
	}

	if err != nil {
		status.State = domain.CollectorUnavailable
		status.Message = err.Error()
		snapshot.Collection.Kernel = status
		return
	}

	status.State = domain.CollectorOK
	now := time.Now()
	status.LastSuccess = &now

	snapshot.Kernel = value
	snapshot.Collection.Kernel = status
}
