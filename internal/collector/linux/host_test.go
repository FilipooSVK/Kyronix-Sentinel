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

	kernelDir := filepath.Join(
		procRoot,
		"sys",
		"kernel",
	)

	if err := os.MkdirAll(kernelDir, 0o755); err != nil {
		t.Fatalf("failed to create kernel fixture directory: %v", err)
	}

	err = os.WriteFile(
		filepath.Join(kernelDir, "osrelease"),
		[]byte("6.12.47+rpt-rpi-v8\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create kernel fixture: %v", err)
	}

	osReleasePath := filepath.Join(
		t.TempDir(),
		"os-release",
	)

	err = os.WriteFile(
		osReleasePath,
		[]byte(
			"ID=debian\n"+
				"NAME=\"Debian GNU/Linux\"\n"+
				"PRETTY_NAME=\"Debian GNU/Linux 13 (trixie)\"\n"+
				"VERSION_ID=\"13\"\n",
		),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create os-release fixture: %v", err)
	}

	collector := &HostCollector{
		procRoot:      procRoot,
		osReleasePath: osReleasePath,
		hostname: func() (string, error) {
			return "kyronix-test-host", nil
		},
		architecture: func() string {
			return "arm64"
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

	if stats.Architecture != "arm64" {
		t.Errorf(
			"architecture mismatch: got %q, want %q",
			stats.Architecture,
			"arm64",
		)
	}

	if stats.OSID != "debian" {
		t.Errorf(
			"os id mismatch: got %q, want %q",
			stats.OSID,
			"debian",
		)
	}

	if stats.OSName != "Debian GNU/Linux 13 (trixie)" {
		t.Errorf(
			"os name mismatch: got %q",
			stats.OSName,
		)
	}

	if stats.OSVersion != "13" {
		t.Errorf(
			"os version mismatch: got %q, want %q",
			stats.OSVersion,
			"13",
		)
	}

	if stats.KernelVersion != "6.12.47+rpt-rpi-v8" {
		t.Errorf(
			"kernel version mismatch: got %q",
			stats.KernelVersion,
		)
	}
}

func TestHostCollectorMetadataIsBestEffort(t *testing.T) {
	procRoot := t.TempDir()

	err := os.WriteFile(
		filepath.Join(procRoot, "uptime"),
		[]byte("100.00 200.00\n"),
		0o600,
	)
	if err != nil {
		t.Fatalf("failed to create uptime fixture: %v", err)
	}

	collector := &HostCollector{
		procRoot: procRoot,
		osReleasePath: filepath.Join(
			t.TempDir(),
			"missing-os-release",
		),
		hostname: func() (string, error) {
			return "minimal-linux-host", nil
		},
		architecture: func() string {
			return "amd64"
		},
	}

	stats, err := collector.CollectHost(context.Background())
	if err != nil {
		t.Fatalf(
			"optional metadata failure made host collection fail: %v",
			err,
		)
	}

	if stats.Hostname != "minimal-linux-host" {
		t.Errorf(
			"hostname mismatch: got %q",
			stats.Hostname,
		)
	}

	if stats.Architecture != "amd64" {
		t.Errorf(
			"architecture mismatch: got %q",
			stats.Architecture,
		)
	}

	if stats.OSID != "" {
		t.Errorf(
			"expected empty OS ID, got %q",
			stats.OSID,
		)
	}

	if stats.KernelVersion != "" {
		t.Errorf(
			"expected empty kernel version, got %q",
			stats.KernelVersion,
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
