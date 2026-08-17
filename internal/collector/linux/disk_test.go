package linux

import (
	"context"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestDiskCollectorCollectDisks(t *testing.T) {
	procRoot := t.TempDir()
	selfDir := filepath.Join(procRoot, "self")

	if err := os.MkdirAll(selfDir, 0o755); err != nil {
		t.Fatalf("failed to create self directory: %v", err)
	}

	mountinfo := `36 25 0:32 / / rw,relatime - ext4 /dev/root rw
37 36 0:5 / /proc rw,nosuid,nodev,noexec,relatime - proc proc rw
38 36 0:6 / /sys rw,nosuid,nodev,noexec,relatime - sysfs sysfs rw
39 36 8:1 / /data rw,relatime - ext4 /dev/sda1 rw
`

	if err := os.WriteFile(
		filepath.Join(selfDir, "mountinfo"),
		[]byte(mountinfo),
		0o600,
	); err != nil {
		t.Fatalf("failed to create mountinfo fixture: %v", err)
	}

	collector := &DiskCollector{
		procRoot: procRoot,
		statfs: func(path string, fs *syscall.Statfs_t) error {
			switch path {
			case "/":
				fs.Bsize = 4096
				fs.Blocks = 1000
				fs.Bfree = 400
				fs.Bavail = 350

			case "/data":
				fs.Bsize = 4096
				fs.Blocks = 2000
				fs.Bfree = 1000
				fs.Bavail = 900

			default:
				t.Fatalf("unexpected statfs path: %s", path)
			}

			return nil
		},
	}

	stats, err := collector.CollectDisks(context.Background())
	if err != nil {
		t.Fatalf("CollectDisks returned error: %v", err)
	}

	if len(stats) != 2 {
		t.Fatalf(
			"filesystem count mismatch: got %d, want 2",
			len(stats),
		)
	}

	root := stats[0]
	data := stats[1]

	if root.MountPoint != "/" {
		t.Errorf(
			"root mount point mismatch: got %q, want %q",
			root.MountPoint,
			"/",
		)
	}

	if root.Device != "/dev/root" {
		t.Errorf(
			"root device mismatch: got %q, want %q",
			root.Device,
			"/dev/root",
		)
	}

	if root.FilesystemType != "ext4" {
		t.Errorf(
			"root filesystem type mismatch: got %q, want %q",
			root.FilesystemType,
			"ext4",
		)
	}

	if root.TotalBytes != 1000*4096 {
		t.Errorf(
			"root TotalBytes mismatch: got %d, want %d",
			root.TotalBytes,
			1000*4096,
		)
	}

	if root.UsedBytes != 600*4096 {
		t.Errorf(
			"root UsedBytes mismatch: got %d, want %d",
			root.UsedBytes,
			600*4096,
		)
	}

	if root.AvailableBytes != 350*4096 {
		t.Errorf(
			"root AvailableBytes mismatch: got %d, want %d",
			root.AvailableBytes,
			350*4096,
		)
	}

	expectedRootPercent := float64(600) / float64(600+350) * 100

	if root.UsedPercent != expectedRootPercent {
		t.Errorf(
			"root UsedPercent mismatch: got %.4f, want %.4f",
			root.UsedPercent,
			expectedRootPercent,
		)
	}

	if data.MountPoint != "/data" {
		t.Errorf(
			"data mount point mismatch: got %q, want %q",
			data.MountPoint,
			"/data",
		)
	}
}

func TestDiskCollectorPreservesRootOverlayFilesystem(t *testing.T) {
	procRoot := t.TempDir()
	selfDir := filepath.Join(procRoot, "self")

	if err := os.MkdirAll(selfDir, 0o755); err != nil {
		t.Fatalf("failed to create self directory: %v", err)
	}

	mountinfo := `36 25 0:32 / / rw,relatime - overlay overlay rw
`

	if err := os.WriteFile(
		filepath.Join(selfDir, "mountinfo"),
		[]byte(mountinfo),
		0o600,
	); err != nil {
		t.Fatalf("failed to create mountinfo fixture: %v", err)
	}

	collector := &DiskCollector{
		procRoot: procRoot,
		statfs: func(path string, fs *syscall.Statfs_t) error {
			if path != "/" {
				t.Fatalf("unexpected statfs path: %s", path)
			}

			fs.Bsize = 4096
			fs.Blocks = 1000
			fs.Bfree = 500
			fs.Bavail = 500

			return nil
		},
	}

	stats, err := collector.CollectDisks(context.Background())
	if err != nil {
		t.Fatalf("CollectDisks returned error: %v", err)
	}

	if len(stats) != 1 {
		t.Fatalf(
			"filesystem count mismatch: got %d, want 1",
			len(stats),
		)
	}

	if stats[0].FilesystemType != "overlay" {
		t.Errorf(
			"filesystem type mismatch: got %q, want %q",
			stats[0].FilesystemType,
			"overlay",
		)
	}
}

func TestParseMountInfoEscapedFields(t *testing.T) {
	line := `42 36 8:1 / /mnt/My\040Disk rw,relatime - ext4 /dev/disk\040one rw`

	mount, err := parseMountInfoLine(line)
	if err != nil {
		t.Fatalf("parseMountInfoLine returned error: %v", err)
	}

	if mount.mountPoint != "/mnt/My Disk" {
		t.Errorf(
			"mount point mismatch: got %q, want %q",
			mount.mountPoint,
			"/mnt/My Disk",
		)
	}

	if mount.device != "/dev/disk one" {
		t.Errorf(
			"device mismatch: got %q, want %q",
			mount.device,
			"/dev/disk one",
		)
	}
}

func TestDiskCollectorRejectsInvalidMountInfo(t *testing.T) {
	procRoot := t.TempDir()
	selfDir := filepath.Join(procRoot, "self")

	if err := os.MkdirAll(selfDir, 0o755); err != nil {
		t.Fatalf("failed to create self directory: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(selfDir, "mountinfo"),
		[]byte("invalid mountinfo record\n"),
		0o600,
	); err != nil {
		t.Fatalf("failed to create mountinfo fixture: %v", err)
	}

	collector := &DiskCollector{
		procRoot: procRoot,
		statfs:   syscall.Statfs,
	}

	_, err := collector.CollectDisks(context.Background())
	if err == nil {
		t.Fatal("expected CollectDisks to return an error")
	}
}
