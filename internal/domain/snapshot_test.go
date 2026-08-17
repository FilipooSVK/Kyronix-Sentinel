package domain

import (
	"testing"
	"time"
)

func TestSnapshotStoresHostState(t *testing.T) {
	cpuUsage := 14.2

	snapshot := Snapshot{
		Timestamp: time.Now(),
		Host: HostStats{
			Hostname: "kyronix-stratus-os",
			Uptime:   24 * time.Hour,
		},
		CPU: CPUStats{
			UsagePercent: &cpuUsage,
			CoreCount:    4,
			Load1:        0.32,
			Load5:        0.28,
			Load15:       0.21,
		},
		Memory: MemoryStats{
			TotalBytes:     8 * 1024 * 1024 * 1024,
			AvailableBytes: 5 * 1024 * 1024 * 1024,
		},
		Collection: CollectionStatus{
			CPU: CollectorStatus{
				State: CollectorOK,
			},
			Memory: CollectorStatus{
				State: CollectorOK,
			},
		},
	}

	if snapshot.Host.Hostname != "kyronix-stratus-os" {
		t.Errorf(
			"Hostname mismatch: got %q",
			snapshot.Host.Hostname,
		)
	}

	if snapshot.CPU.UsagePercent == nil {
		t.Fatal("CPU usage unexpectedly nil")
	}

	if *snapshot.CPU.UsagePercent != cpuUsage {
		t.Errorf(
			"CPU usage mismatch: got %.2f, want %.2f",
			*snapshot.CPU.UsagePercent,
			cpuUsage,
		)
	}

	if snapshot.Collection.CPU.State != CollectorOK {
		t.Errorf(
			"CPU collector state mismatch: got %q",
			snapshot.Collection.CPU.State,
		)
	}
}
