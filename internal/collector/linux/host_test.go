package linux

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHostCollectorCollectHost(t *testing.T) {
	procRoot := t.TempDir()

	err := os.WriteFile(
		filepath.Join(procRoot, "uptime"),
		[]byte("12345.67 54321.00\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create uptime fixture: %v", err)
	}

	collector := &HostCollector{
		procRoot: procRoot,
		hostname: func() (string, error) {
			return "kyronix-test-host", nil
		},
	}

	stats, err := collector.CollectHost(context.Background())
	if err != nil {
		t.Fatalf("CollectHost returned error: %v", err)
	}

	if stats.Hostname != "kyronix-test-host" {
		t.Errorf(
			"hostname mismatch: got %q, want %q",
			stats.Hostname,
			"kyronix-test-host",
		)
	}

	expectedUptime := 12345*time.Second + 670*time.Millisecond

	if stats.Uptime != expectedUptime {
		t.Errorf(
			"uptime mismatch: got %s, want %s",
			stats.Uptime,
			expectedUptime,
		)
	}
}

func TestHostCollectorRejectsInvalidUptime(t *testing.T) {
	procRoot := t.TempDir()

	err := os.WriteFile(
		filepath.Join(procRoot, "uptime"),
		[]byte("invalid\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create uptime fixture: %v", err)
	}

	collector := &HostCollector{
		procRoot: procRoot,
		hostname: func() (string, error) {
			return "kyronix-test-host", nil
		},
	}

	_, err = collector.CollectHost(context.Background())
	if err == nil {
		t.Fatal("expected CollectHost to return an error")
	}
}
