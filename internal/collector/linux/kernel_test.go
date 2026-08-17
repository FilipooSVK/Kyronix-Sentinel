package linux

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestKernelCollectorWithSystemAndCgroupOOM(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()

	vmstat := `nr_free_pages 1000
oom_kill 3
pgfault 123456
`

	if err := os.WriteFile(
		filepath.Join(procRoot, "vmstat"),
		[]byte(vmstat),
		0o600,
	); err != nil {
		t.Fatalf("failed to create vmstat fixture: %v", err)
	}

	memoryEvents := `low 0
high 2
max 4
oom 5
oom_kill 2
oom_group_kill 0
`

	if err := os.WriteFile(
		filepath.Join(cgroupRoot, "memory.events"),
		[]byte(memoryEvents),
		0o600,
	); err != nil {
		t.Fatalf("failed to create memory.events fixture: %v", err)
	}

	collector := &KernelCollector{
		procRoot:   procRoot,
		cgroupRoot: cgroupRoot,
	}

	stats, err := collector.CollectKernel(context.Background())
	if err != nil {
		t.Fatalf("CollectKernel returned error: %v", err)
	}

	if stats.OOM.SystemKillCount == nil {
		t.Fatal("SystemKillCount unexpectedly nil")
	}

	if *stats.OOM.SystemKillCount != 3 {
		t.Errorf(
			"SystemKillCount mismatch: got %d, want 3",
			*stats.OOM.SystemKillCount,
		)
	}

	if stats.OOM.CgroupOOMCount == nil {
		t.Fatal("CgroupOOMCount unexpectedly nil")
	}

	if *stats.OOM.CgroupOOMCount != 5 {
		t.Errorf(
			"CgroupOOMCount mismatch: got %d, want 5",
			*stats.OOM.CgroupOOMCount,
		)
	}

	if stats.OOM.CgroupKillCount == nil {
		t.Fatal("CgroupKillCount unexpectedly nil")
	}

	if *stats.OOM.CgroupKillCount != 2 {
		t.Errorf(
			"CgroupKillCount mismatch: got %d, want 2",
			*stats.OOM.CgroupKillCount,
		)
	}
}

func TestKernelCollectorWithSystemOOMOnly(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()

	vmstat := `nr_free_pages 1000
oom_kill 7
pgfault 123456
`

	if err := os.WriteFile(
		filepath.Join(procRoot, "vmstat"),
		[]byte(vmstat),
		0o600,
	); err != nil {
		t.Fatalf("failed to create vmstat fixture: %v", err)
	}

	collector := &KernelCollector{
		procRoot:   procRoot,
		cgroupRoot: filepath.Join(cgroupRoot, "missing"),
	}

	stats, err := collector.CollectKernel(context.Background())
	if err != nil {
		t.Fatalf("CollectKernel returned error: %v", err)
	}

	if stats.OOM.SystemKillCount == nil {
		t.Fatal("SystemKillCount unexpectedly nil")
	}

	if *stats.OOM.SystemKillCount != 7 {
		t.Errorf(
			"SystemKillCount mismatch: got %d, want 7",
			*stats.OOM.SystemKillCount,
		)
	}

	if stats.OOM.CgroupOOMCount != nil {
		t.Errorf(
			"expected CgroupOOMCount to be nil, got %d",
			*stats.OOM.CgroupOOMCount,
		)
	}

	if stats.OOM.CgroupKillCount != nil {
		t.Errorf(
			"expected CgroupKillCount to be nil, got %d",
			*stats.OOM.CgroupKillCount,
		)
	}
}

func TestKernelCollectorWithCgroupOOMOnly(t *testing.T) {
	procRoot := t.TempDir()
	cgroupRoot := t.TempDir()

	memoryEvents := `low 0
high 0
max 0
oom 4
oom_kill 1
oom_group_kill 0
`

	if err := os.WriteFile(
		filepath.Join(cgroupRoot, "memory.events"),
		[]byte(memoryEvents),
		0o600,
	); err != nil {
		t.Fatalf("failed to create memory.events fixture: %v", err)
	}

	collector := &KernelCollector{
		procRoot:   filepath.Join(procRoot, "missing"),
		cgroupRoot: cgroupRoot,
	}

	stats, err := collector.CollectKernel(context.Background())
	if err != nil {
		t.Fatalf("CollectKernel returned error: %v", err)
	}

	if stats.OOM.SystemKillCount != nil {
		t.Errorf(
			"expected SystemKillCount to be nil, got %d",
			*stats.OOM.SystemKillCount,
		)
	}

	if stats.OOM.CgroupOOMCount == nil {
		t.Fatal("CgroupOOMCount unexpectedly nil")
	}

	if *stats.OOM.CgroupOOMCount != 4 {
		t.Errorf(
			"CgroupOOMCount mismatch: got %d, want 4",
			*stats.OOM.CgroupOOMCount,
		)
	}

	if stats.OOM.CgroupKillCount == nil {
		t.Fatal("CgroupKillCount unexpectedly nil")
	}

	if *stats.OOM.CgroupKillCount != 1 {
		t.Errorf(
			"CgroupKillCount mismatch: got %d, want 1",
			*stats.OOM.CgroupKillCount,
		)
	}
}

func TestKernelCollectorUnavailable(t *testing.T) {
	collector := &KernelCollector{
		procRoot:   filepath.Join(t.TempDir(), "missing-proc"),
		cgroupRoot: filepath.Join(t.TempDir(), "missing-cgroup"),
	}

	_, err := collector.CollectKernel(context.Background())
	if err == nil {
		t.Fatal("expected CollectKernel to return an error")
	}

	if !errors.Is(err, ErrKernelTelemetryUnavailable) {
		t.Fatalf(
			"expected ErrKernelTelemetryUnavailable, got %v",
			err,
		)
	}
}

func TestKernelCollectorRejectsInvalidSystemOOMValue(t *testing.T) {
	procRoot := t.TempDir()

	vmstat := `nr_free_pages 1000
oom_kill invalid
`

	if err := os.WriteFile(
		filepath.Join(procRoot, "vmstat"),
		[]byte(vmstat),
		0o600,
	); err != nil {
		t.Fatalf("failed to create vmstat fixture: %v", err)
	}

	collector := &KernelCollector{
		procRoot:   procRoot,
		cgroupRoot: filepath.Join(t.TempDir(), "missing"),
	}

	_, err := collector.CollectKernel(context.Background())
	if err == nil {
		t.Fatal("expected CollectKernel to return an error")
	}
}

func TestKernelCollectorRejectsInvalidCgroupOOMValue(t *testing.T) {
	cgroupRoot := t.TempDir()

	memoryEvents := `low 0
high 0
max 0
oom invalid
oom_kill 0
`

	if err := os.WriteFile(
		filepath.Join(cgroupRoot, "memory.events"),
		[]byte(memoryEvents),
		0o600,
	); err != nil {
		t.Fatalf("failed to create memory.events fixture: %v", err)
	}

	collector := &KernelCollector{
		procRoot:   filepath.Join(t.TempDir(), "missing"),
		cgroupRoot: cgroupRoot,
	}

	_, err := collector.CollectKernel(context.Background())
	if err == nil {
		t.Fatal("expected CollectKernel to return an error")
	}
}
