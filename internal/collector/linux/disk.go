package linux

import (
	"bufio"
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"kyronix/sentinel/internal/domain"
)

type mountInfo struct {
	mountPoint     string
	device         string
	filesystemType string
}

type statfsFunc func(string, *syscall.Statfs_t) error

// DiskCollector collects filesystem capacity information from Linux.
type DiskCollector struct {
	procRoot string
	statfs   statfsFunc
}

// NewDiskCollector creates a Linux disk collector using the real system
// interfaces.
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{
		procRoot: defaultProcRoot,
		statfs:   syscall.Statfs,
	}
}

// CollectDisks collects capacity information for monitored local filesystems.
func (c *DiskCollector) CollectDisks(
	ctx context.Context,
) ([]domain.DiskStats, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	mounts, err := c.readMountInfo()
	if err != nil {
		return nil, err
	}

	stats := make([]domain.DiskStats, 0, len(mounts))

	for _, mount := range mounts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		var fs syscall.Statfs_t

		if err := c.statfs(mount.mountPoint, &fs); err != nil {
			return nil, fmt.Errorf(
				"statfs %s: %w",
				mount.mountPoint,
				err,
			)
		}

		disk, err := buildDiskStats(mount, fs)
		if err != nil {
			return nil, fmt.Errorf(
				"filesystem %s: %w",
				mount.mountPoint,
				err,
			)
		}

		stats = append(stats, disk)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].MountPoint < stats[j].MountPoint
	})

	return stats, nil
}

func (c *DiskCollector) readMountInfo() ([]mountInfo, error) {
	path := filepath.Join(c.procRoot, "self", "mountinfo")

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var mounts []mountInfo
	seen := make(map[string]struct{})

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		mount, err := parseMountInfoLine(scanner.Text())
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		if !shouldMonitorFilesystem(mount) {
			continue
		}

		if _, ok := seen[mount.mountPoint]; ok {
			continue
		}

		seen[mount.mountPoint] = struct{}{}
		mounts = append(mounts, mount)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	if len(mounts) == 0 {
		return nil, fmt.Errorf(
			"parse %s: no monitored filesystems found",
			path,
		)
	}

	return mounts, nil
}

func parseMountInfoLine(line string) (mountInfo, error) {
	parts := strings.SplitN(line, " - ", 2)
	if len(parts) != 2 {
		return mountInfo{}, fmt.Errorf(
			"mountinfo separator missing",
		)
	}

	left := strings.Fields(parts[0])
	right := strings.Fields(parts[1])

	if len(left) < 5 {
		return mountInfo{}, fmt.Errorf(
			"mountinfo record has too few fields before separator",
		)
	}

	if len(right) < 2 {
		return mountInfo{}, fmt.Errorf(
			"mountinfo record has too few fields after separator",
		)
	}

	mountPoint, err := decodeMountInfoField(left[4])
	if err != nil {
		return mountInfo{}, fmt.Errorf(
			"decode mount point %q: %w",
			left[4],
			err,
		)
	}

	device, err := decodeMountInfoField(right[1])
	if err != nil {
		return mountInfo{}, fmt.Errorf(
			"decode device %q: %w",
			right[1],
			err,
		)
	}

	return mountInfo{
		mountPoint:     mountPoint,
		device:         device,
		filesystemType: right[0],
	}, nil
}

func decodeMountInfoField(value string) (string, error) {
	var builder strings.Builder

	for i := 0; i < len(value); {
		if value[i] != '\\' {
			builder.WriteByte(value[i])
			i++
			continue
		}

		if i+3 >= len(value) {
			return "", fmt.Errorf("incomplete escape sequence")
		}

		escaped := value[i+1 : i+4]

		decoded, err := strconv.ParseUint(escaped, 8, 8)
		if err != nil {
			return "", fmt.Errorf(
				"invalid escape sequence \\%s",
				escaped,
			)
		}

		builder.WriteByte(byte(decoded))
		i += 4
	}

	return builder.String(), nil
}

func shouldMonitorFilesystem(mount mountInfo) bool {
	// The root filesystem is always important, including container
	// environments where it may be backed by overlayfs.
	if mount.mountPoint == "/" {
		return true
	}

	switch mount.filesystemType {
	case "ext2",
		"ext3",
		"ext4",
		"xfs",
		"btrfs",
		"zfs",
		"f2fs",
		"vfat",
		"exfat",
		"ntfs",
		"ntfs3":
		return true
	default:
		return false
	}
}

func buildDiskStats(
	mount mountInfo,
	fs syscall.Statfs_t,
) (domain.DiskStats, error) {
	if fs.Bsize <= 0 {
		return domain.DiskStats{}, fmt.Errorf(
			"invalid filesystem block size %d",
			fs.Bsize,
		)
	}

	if fs.Bfree > fs.Blocks {
		return domain.DiskStats{}, fmt.Errorf(
			"free blocks exceed total blocks",
		)
	}

	if fs.Bavail > fs.Blocks {
		return domain.DiskStats{}, fmt.Errorf(
			"available blocks exceed total blocks",
		)
	}

	blockSize := uint64(fs.Bsize)
	usedBlocks := fs.Blocks - fs.Bfree

	totalBytes, err := blocksToBytes(fs.Blocks, blockSize)
	if err != nil {
		return domain.DiskStats{}, fmt.Errorf(
			"calculate total bytes: %w",
			err,
		)
	}

	usedBytes, err := blocksToBytes(usedBlocks, blockSize)
	if err != nil {
		return domain.DiskStats{}, fmt.Errorf(
			"calculate used bytes: %w",
			err,
		)
	}

	availableBytes, err := blocksToBytes(fs.Bavail, blockSize)
	if err != nil {
		return domain.DiskStats{}, fmt.Errorf(
			"calculate available bytes: %w",
			err,
		)
	}

	var usedPercent float64

	usableBlocks := usedBlocks + fs.Bavail
	if usableBlocks > 0 {
		usedPercent = float64(usedBlocks) /
			float64(usableBlocks) * 100
	}

	return domain.DiskStats{
		MountPoint:     mount.mountPoint,
		Device:         mount.device,
		FilesystemType: mount.filesystemType,
		TotalBytes:     totalBytes,
		UsedBytes:      usedBytes,
		AvailableBytes: availableBytes,
		UsedPercent:    usedPercent,
	}, nil
}

func blocksToBytes(blocks, blockSize uint64) (uint64, error) {
	if blockSize != 0 && blocks > math.MaxUint64/blockSize {
		return 0, fmt.Errorf("byte count overflows uint64")
	}

	return blocks * blockSize, nil
}
