package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
)

type fakeHostCollector struct {
	err error
}

func (f fakeHostCollector) CollectHost(context.Context) (domain.HostStats, error) {
	return domain.HostStats{
		Hostname: "test-host",
		Uptime:   100 * time.Second,
	}, f.err
}

type fakeCPUCollector struct {
	err error
}

func (f fakeCPUCollector) CollectCPU(context.Context) (domain.CPUStats, error) {
	value := 25.0

	return domain.CPUStats{
		UsagePercent: &value,
		CoreCount:    4,
		Load1:        0.5,
	}, f.err
}

type fakeMemoryCollector struct {
	err error
}

func (f fakeMemoryCollector) CollectMemory(context.Context) (domain.MemoryStats, error) {
	return domain.MemoryStats{
		TotalBytes:     1024,
		AvailableBytes: 512,
	}, f.err
}

type fakePressureCollector struct {
	err error
}

func (f fakePressureCollector) CollectPressure(context.Context) (domain.PressureStats, error) {
	return domain.PressureStats{}, f.err
}

type fakeDiskCollector struct {
	err error
}

func (f fakeDiskCollector) CollectDisks(context.Context) ([]domain.DiskStats, error) {
	return []domain.DiskStats{
		{
			MountPoint: "/",
		},
	}, f.err
}

type fakeKernelCollector struct {
	err error
}

func (f fakeKernelCollector) CollectKernel(context.Context) (domain.KernelStats, error) {
	return domain.KernelStats{}, f.err
}

func TestSnapshotManagerCollect(t *testing.T) {
	manager := NewSnapshotManager(
		fakeHostCollector{},
		fakeCPUCollector{},
		fakeMemoryCollector{},
		fakePressureCollector{},
		fakeDiskCollector{},
		fakeKernelCollector{},
	)

	snapshot := manager.Collect(context.Background())

	if snapshot.Timestamp.IsZero() {
		t.Fatal("snapshot timestamp is empty")
	}

	if snapshot.Host.Uptime != 100*time.Second {
		t.Errorf(
			"host uptime mismatch: got %s",
			snapshot.Host.Uptime,
		)
	}

	if snapshot.CPU.CoreCount != 4 {
		t.Errorf(
			"cpu core count mismatch: got %d",
			snapshot.CPU.CoreCount,
		)
	}

	if len(snapshot.Disks) != 1 {
		t.Errorf(
			"disk count mismatch: got %d",
			len(snapshot.Disks),
		)
	}

	if snapshot.Collection.Host.State != domain.CollectorOK {
		t.Errorf(
			"host state mismatch: got %s",
			snapshot.Collection.Host.State,
		)
	}

	if snapshot.Collection.Pressure.State != domain.CollectorOK {
		t.Errorf(
			"pressure state mismatch: got %s",
			snapshot.Collection.Pressure.State,
		)
	}
}

func TestSnapshotManagerHandlesUnavailableCollectors(t *testing.T) {
	manager := NewSnapshotManager(
		fakeHostCollector{},
		fakeCPUCollector{},
		fakeMemoryCollector{},
		fakePressureCollector{
			err: errors.New("pressure unavailable"),
		},
		fakeDiskCollector{},
		fakeKernelCollector{
			err: errors.New("kernel unavailable"),
		},
	)

	snapshot := manager.Collect(context.Background())

	if snapshot.Collection.Pressure.State != domain.CollectorUnavailable {
		t.Errorf(
			"pressure state mismatch: got %s",
			snapshot.Collection.Pressure.State,
		)
	}

	if snapshot.Collection.Kernel.State != domain.CollectorUnavailable {
		t.Errorf(
			"kernel state mismatch: got %s",
			snapshot.Collection.Kernel.State,
		)
	}

	if snapshot.Collection.Memory.State != domain.CollectorOK {
		t.Errorf(
			"memory should remain OK, got %s",
			snapshot.Collection.Memory.State,
		)
	}
}

func TestSnapshotManagerContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	manager := NewSnapshotManager(
		fakeHostCollector{},
		fakeCPUCollector{},
		fakeMemoryCollector{},
		fakePressureCollector{},
		fakeDiskCollector{},
		fakeKernelCollector{},
	)

	start := time.Now()

	_ = manager.Collect(ctx)

	if time.Since(start) > time.Second {
		t.Fatal("collector manager ignored context cancellation")
	}
}
