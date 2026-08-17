package linux

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kyronix/sentinel/internal/domain"
)

// ErrKernelTelemetryUnavailable indicates that no supported kernel
// telemetry source is available.
var ErrKernelTelemetryUnavailable = errors.New("kernel telemetry unavailable")

// KernelCollector collects kernel-level health information from Linux.
type KernelCollector struct {
	procRoot   string
	cgroupRoot string
}

// NewKernelCollector creates a Linux kernel collector using the real
// system interfaces.
func NewKernelCollector() *KernelCollector {
	return &KernelCollector{
		procRoot:   defaultProcRoot,
		cgroupRoot: "/sys/fs/cgroup",
	}
}

// CollectKernel collects kernel OOM telemetry.
//
// System-wide OOM kill information is read from /proc/vmstat.
// Cgroup-specific OOM information is read from cgroup v2 memory.events
// when available.
func (c *KernelCollector) CollectKernel(
	ctx context.Context,
) (domain.KernelStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.KernelStats{}, err
	}

	var stats domain.KernelStats
	var available bool

	systemKills, err := c.readSystemOOMKills()
	if err == nil {
		stats.OOM.SystemKillCount = &systemKills
		available = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return domain.KernelStats{}, err
	}

	cgroupOOM, cgroupKills, err := c.readCgroupOOMEvents()
	if err == nil {
		stats.OOM.CgroupOOMCount = &cgroupOOM
		stats.OOM.CgroupKillCount = &cgroupKills
		available = true
	} else if !errors.Is(err, os.ErrNotExist) {
		return domain.KernelStats{}, err
	}

	if !available {
		return domain.KernelStats{}, ErrKernelTelemetryUnavailable
	}

	return stats, nil
}

func (c *KernelCollector) readSystemOOMKills() (uint64, error) {
	path := filepath.Join(c.procRoot, "vmstat")

	file, err := os.Open(path)
	if err != nil {
		return 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}

		if fields[0] != "oom_kill" {
			continue
		}

		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return 0, fmt.Errorf(
				"parse %s oom_kill value %q: %w",
				path,
				fields[1],
				err,
			)
		}

		return value, nil
	}

	if err := scanner.Err(); err != nil {
		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	return 0, fmt.Errorf("parse %s: oom_kill field missing", path)
}

func (c *KernelCollector) readCgroupOOMEvents() (uint64, uint64, error) {
	path := filepath.Join(c.cgroupRoot, "memory.events")

	file, err := os.Open(path)
	if err != nil {
		return 0, 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var (
		oom      uint64
		oomKill  uint64
		haveOOM  bool
		haveKill bool
	)

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}

		switch fields[0] {
		case "oom":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf(
					"parse %s oom value %q: %w",
					path,
					fields[1],
					err,
				)
			}

			oom = value
			haveOOM = true

		case "oom_kill":
			value, err := strconv.ParseUint(fields[1], 10, 64)
			if err != nil {
				return 0, 0, fmt.Errorf(
					"parse %s oom_kill value %q: %w",
					path,
					fields[1],
					err,
				)
			}

			oomKill = value
			haveKill = true
		}
	}

	if err := scanner.Err(); err != nil {
		return 0, 0, fmt.Errorf("read %s: %w", path, err)
	}

	if !haveOOM {
		return 0, 0, fmt.Errorf("parse %s: oom field missing", path)
	}

	if !haveKill {
		return 0, 0, fmt.Errorf("parse %s: oom_kill field missing", path)
	}

	return oom, oomKill, nil
}
