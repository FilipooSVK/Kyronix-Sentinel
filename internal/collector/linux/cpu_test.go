package linux

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCPUCollectorCollectCPU(t *testing.T) {
	procRoot := t.TempDir()

	firstStat := `cpu  100 0 100 700 100 0 0 0 0 0
cpu0 50 0 50 350 50 0 0 0 0 0
cpu1 50 0 50 350 50 0 0 0 0 0
`

	secondStat := `cpu  200 0 200 800 200 0 0 0 0 0
cpu0 100 0 100 400 100 0 0 0 0 0
cpu1 100 0 100 400 100 0 0 0 0 0
`

	loadavg := "0.42 0.31 0.25 1/123 456\n"

	if err := os.WriteFile(
		filepath.Join(procRoot, "stat"),
		[]byte(firstStat),
		0o600,
	); err != nil {
		t.Fatalf("failed to create stat fixture: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(procRoot, "loadavg"),
		[]byte(loadavg),
		0o600,
	); err != nil {
		t.Fatalf("failed to create loadavg fixture: %v", err)
	}

	collector := &CPUCollector{
		procRoot: procRoot,
	}

	first, err := collector.CollectCPU(context.Background())
	if err != nil {
		t.Fatalf("first CollectCPU returned error: %v", err)
	}

	if first.UsagePercent != nil {
		t.Fatalf(
			"expected first CPU usage to be nil, got %.2f",
			*first.UsagePercent,
		)
	}

	if first.CoreCount != 2 {
		t.Errorf(
			"CoreCount mismatch: got %d, want 2",
			first.CoreCount,
		)
	}

	if first.Load1 != 0.42 {
		t.Errorf(
			"Load1 mismatch: got %.2f, want 0.42",
			first.Load1,
		)
	}

	if first.Load5 != 0.31 {
		t.Errorf(
			"Load5 mismatch: got %.2f, want 0.31",
			first.Load5,
		)
	}

	if first.Load15 != 0.25 {
		t.Errorf(
			"Load15 mismatch: got %.2f, want 0.25",
			first.Load15,
		)
	}

	if err := os.WriteFile(
		filepath.Join(procRoot, "stat"),
		[]byte(secondStat),
		0o600,
	); err != nil {
		t.Fatalf("failed to update stat fixture: %v", err)
	}

	second, err := collector.CollectCPU(context.Background())
	if err != nil {
		t.Fatalf("second CollectCPU returned error: %v", err)
	}

	if second.UsagePercent == nil {
		t.Fatal("expected second CPU usage to be available")
	}

	if *second.UsagePercent != 50 {
		t.Errorf(
			"CPU usage mismatch: got %.2f, want 50.00",
			*second.UsagePercent,
		)
	}
}

func TestCPUCollectorRejectsMissingAggregateCPU(t *testing.T) {
	procRoot := t.TempDir()

	stat := `cpu0 100 0 100 700 100 0 0 0
cpu1 100 0 100 700 100 0 0 0
`

	if err := os.WriteFile(
		filepath.Join(procRoot, "stat"),
		[]byte(stat),
		0o600,
	); err != nil {
		t.Fatalf("failed to create stat fixture: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(procRoot, "loadavg"),
		[]byte("0.10 0.20 0.30 1/100 123\n"),
		0o600,
	); err != nil {
		t.Fatalf("failed to create loadavg fixture: %v", err)
	}

	collector := &CPUCollector{
		procRoot: procRoot,
	}

	_, err := collector.CollectCPU(context.Background())
	if err == nil {
		t.Fatal("expected CollectCPU to return an error")
	}
}

func TestParseCPUSampleDoesNotDoubleCountGuestTime(t *testing.T) {
	fields := []string{
		"100",
		"20",
		"50",
		"700",
		"30",
		"10",
		"5",
		"2",
		"40",
		"10",
	}

	sample, err := parseCPUSample(fields)
	if err != nil {
		t.Fatalf("parseCPUSample returned error: %v", err)
	}

	const expectedTotal uint64 = 917
	const expectedIdle uint64 = 730

	if sample.total != expectedTotal {
		t.Errorf(
			"total mismatch: got %d, want %d",
			sample.total,
			expectedTotal,
		)
	}

	if sample.idle != expectedIdle {
		t.Errorf(
			"idle mismatch: got %d, want %d",
			sample.idle,
			expectedIdle,
		)
	}
}
