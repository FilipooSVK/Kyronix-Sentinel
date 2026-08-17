package linux

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kyronix/sentinel/internal/domain"
)

// MemoryCollector collects memory information from Linux /proc.
type MemoryCollector struct {
	procRoot string
}

// NewMemoryCollector creates a Linux memory collector using the real
// system interfaces.
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{
		procRoot: defaultProcRoot,
	}
}

// CollectMemory collects physical memory and swap utilization.
func (c *MemoryCollector) CollectMemory(ctx context.Context) (domain.MemoryStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.MemoryStats{}, err
	}

	path := filepath.Join(c.procRoot, "meminfo")

	file, err := os.Open(path)
	if err != nil {
		return domain.MemoryStats{}, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	stats, err := parseMeminfo(file)
	if err != nil {
		return domain.MemoryStats{}, fmt.Errorf("parse %s: %w", path, err)
	}

	return stats, nil
}

func parseMeminfo(r io.Reader) (domain.MemoryStats, error) {
	values := make(map[string]uint64)

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		key := strings.TrimSuffix(fields[0], ":")

		switch key {
		case "MemTotal", "MemAvailable", "SwapTotal", "SwapFree":
		default:
			continue
		}

		valueKB, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return domain.MemoryStats{}, fmt.Errorf(
				"invalid value for %s: %q",
				key,
				fields[1],
			)
		}

		if valueKB > ^uint64(0)/1024 {
			return domain.MemoryStats{}, fmt.Errorf(
				"value for %s overflows uint64",
				key,
			)
		}

		values[key] = valueKB * 1024
	}

	if err := scanner.Err(); err != nil {
		return domain.MemoryStats{}, fmt.Errorf("read meminfo: %w", err)
	}

	required := []string{
		"MemTotal",
		"MemAvailable",
		"SwapTotal",
		"SwapFree",
	}

	for _, key := range required {
		if _, ok := values[key]; !ok {
			return domain.MemoryStats{}, fmt.Errorf(
				"required field %s missing",
				key,
			)
		}
	}

	total := values["MemTotal"]
	available := values["MemAvailable"]
	swapTotal := values["SwapTotal"]
	swapFree := values["SwapFree"]

	if total == 0 {
		return domain.MemoryStats{}, fmt.Errorf("MemTotal must be greater than zero")
	}

	if available > total {
		return domain.MemoryStats{}, fmt.Errorf(
			"MemAvailable exceeds MemTotal",
		)
	}

	if swapFree > swapTotal {
		return domain.MemoryStats{}, fmt.Errorf(
			"SwapFree exceeds SwapTotal",
		)
	}

	used := total - available
	usedPercent := float64(used) / float64(total) * 100

	swapUsed := swapTotal - swapFree

	var swapPercent float64
	if swapTotal > 0 {
		swapPercent = float64(swapUsed) / float64(swapTotal) * 100
	}

	return domain.MemoryStats{
		TotalBytes:     total,
		AvailableBytes: available,
		UsedBytes:      used,
		UsedPercent:    usedPercent,

		SwapTotalBytes: swapTotal,
		SwapUsedBytes:  swapUsed,
		SwapPercent:    swapPercent,
	}, nil
}
