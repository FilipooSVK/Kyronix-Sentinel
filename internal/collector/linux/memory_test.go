package linux

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestMemoryCollectorCollectMemory(t *testing.T) {
	procRoot := t.TempDir()

	meminfo := `MemTotal:        8000000 kB
MemFree:         1000000 kB
MemAvailable:    5000000 kB
Buffers:          100000 kB
Cached:          2000000 kB
SwapCached:            0 kB
SwapTotal:       2000000 kB
SwapFree:        1500000 kB
`

	err := os.WriteFile(
		filepath.Join(procRoot, "meminfo"),
		[]byte(meminfo),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create meminfo fixture: %v", err)
	}

	collector := &MemoryCollector{
		procRoot: procRoot,
	}

	stats, err := collector.CollectMemory(context.Background())
	if err != nil {
		t.Fatalf("CollectMemory returned error: %v", err)
	}

	const kb = uint64(1024)

	if stats.TotalBytes != 8000000*kb {
		t.Errorf(
			"TotalBytes mismatch: got %d, want %d",
			stats.TotalBytes,
			8000000*kb,
		)
	}

	if stats.AvailableBytes != 5000000*kb {
		t.Errorf(
			"AvailableBytes mismatch: got %d, want %d",
			stats.AvailableBytes,
			5000000*kb,
		)
	}

	if stats.UsedBytes != 3000000*kb {
		t.Errorf(
			"UsedBytes mismatch: got %d, want %d",
			stats.UsedBytes,
			3000000*kb,
		)
	}

	if stats.UsedPercent != 37.5 {
		t.Errorf(
			"UsedPercent mismatch: got %.2f, want 37.50",
			stats.UsedPercent,
		)
	}

	if stats.SwapUsedBytes != 500000*kb {
		t.Errorf(
			"SwapUsedBytes mismatch: got %d, want %d",
			stats.SwapUsedBytes,
			500000*kb,
		)
	}

	if stats.SwapPercent != 25 {
		t.Errorf(
			"SwapPercent mismatch: got %.2f, want 25.00",
			stats.SwapPercent,
		)
	}
}

func TestMemoryCollectorWithoutSwap(t *testing.T) {
	procRoot := t.TempDir()

	meminfo := `MemTotal:        4000000 kB
MemAvailable:    3000000 kB
SwapTotal:             0 kB
SwapFree:              0 kB
`

	err := os.WriteFile(
		filepath.Join(procRoot, "meminfo"),
		[]byte(meminfo),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create meminfo fixture: %v", err)
	}

	collector := &MemoryCollector{
		procRoot: procRoot,
	}

	stats, err := collector.CollectMemory(context.Background())
	if err != nil {
		t.Fatalf("CollectMemory returned error: %v", err)
	}

	if stats.SwapUsedBytes != 0 {
		t.Errorf(
			"SwapUsedBytes mismatch: got %d, want 0",
			stats.SwapUsedBytes,
		)
	}

	if stats.SwapPercent != 0 {
		t.Errorf(
			"SwapPercent mismatch: got %.2f, want 0",
			stats.SwapPercent,
		)
	}
}

func TestMemoryCollectorRejectsMissingField(t *testing.T) {
	procRoot := t.TempDir()

	meminfo := `MemTotal:        8000000 kB
MemAvailable:    5000000 kB
SwapTotal:       2000000 kB
`

	err := os.WriteFile(
		filepath.Join(procRoot, "meminfo"),
		[]byte(meminfo),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create meminfo fixture: %v", err)
	}

	collector := &MemoryCollector{
		procRoot: procRoot,
	}

	_, err = collector.CollectMemory(context.Background())
	if err == nil {
		t.Fatal("expected CollectMemory to return an error")
	}
}
