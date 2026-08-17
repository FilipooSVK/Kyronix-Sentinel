package linux

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"kyronix/sentinel/internal/domain"
)

type cpuSample struct {
	total uint64
	idle  uint64
}

// CPUCollector collects CPU utilization and load information from Linux /proc.
//
// CPU utilization is calculated from the difference between consecutive
// /proc/stat samples. Therefore the first collection does not contain a
// utilization percentage.
type CPUCollector struct {
	procRoot string

	mu       sync.Mutex
	previous *cpuSample
}

// NewCPUCollector creates a Linux CPU collector using the real /proc filesystem.
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{
		procRoot: defaultProcRoot,
	}
}

// CollectCPU collects CPU utilization, logical CPU count and load averages.
func (c *CPUCollector) CollectCPU(ctx context.Context) (domain.CPUStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.CPUStats{}, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	current, coreCount, err := c.readCPUStat()
	if err != nil {
		return domain.CPUStats{}, err
	}

	load1, load5, load15, err := c.readLoadAverage()
	if err != nil {
		return domain.CPUStats{}, err
	}

	var usage *float64

	if c.previous != nil {
		if current.total > c.previous.total && current.idle >= c.previous.idle {
			totalDelta := current.total - c.previous.total
			idleDelta := current.idle - c.previous.idle

			if idleDelta <= totalDelta {
				value := float64(totalDelta-idleDelta) / float64(totalDelta) * 100
				usage = &value
			}
		}
	}

	c.previous = &current

	return domain.CPUStats{
		UsagePercent: usage,
		CoreCount:    coreCount,
		Load1:        load1,
		Load5:        load5,
		Load15:       load15,
	}, nil
}

func (c *CPUCollector) readCPUStat() (cpuSample, int, error) {
	path := filepath.Join(c.procRoot, "stat")

	file, err := os.Open(path)
	if err != nil {
		return cpuSample{}, 0, fmt.Errorf("open %s: %w", path, err)
	}
	defer file.Close()

	var aggregate cpuSample
	var haveAggregate bool
	var coreCount int

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}

		if fields[0] == "cpu" {
			sample, err := parseCPUSample(fields[1:])
			if err != nil {
				return cpuSample{}, 0, fmt.Errorf(
					"parse aggregate CPU statistics: %w",
					err,
				)
			}

			aggregate = sample
			haveAggregate = true
			continue
		}

		if isCPUCore(fields[0]) {
			coreCount++
		}
	}

	if err := scanner.Err(); err != nil {
		return cpuSample{}, 0, fmt.Errorf("read %s: %w", path, err)
	}

	if !haveAggregate {
		return cpuSample{}, 0, fmt.Errorf(
			"parse %s: aggregate CPU record missing",
			path,
		)
	}

	if coreCount == 0 {
		return cpuSample{}, 0, fmt.Errorf(
			"parse %s: no logical CPUs found",
			path,
		)
	}

	return aggregate, coreCount, nil
}

func parseCPUSample(fields []string) (cpuSample, error) {
	if len(fields) < 4 {
		return cpuSample{}, fmt.Errorf(
			"CPU record has %d fields, need at least 4",
			len(fields),
		)
	}

	values := make([]uint64, len(fields))

	for i, field := range fields {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			return cpuSample{}, fmt.Errorf(
				"invalid CPU counter %q: %w",
				field,
				err,
			)
		}

		values[i] = value
	}

	// Linux /proc/stat fields:
	//
	// user nice system idle iowait irq softirq steal guest guest_nice
	//
	// guest and guest_nice are already included in user and nice,
	// therefore they must not be added again.
	fieldCount := len(values)
	if fieldCount > 8 {
		fieldCount = 8
	}

	var total uint64
	for i := 0; i < fieldCount; i++ {
		total += values[i]
	}

	idle := values[3]

	if len(values) > 4 {
		idle += values[4]
	}

	return cpuSample{
		total: total,
		idle:  idle,
	}, nil
}

func isCPUCore(value string) bool {
	if !strings.HasPrefix(value, "cpu") || len(value) <= 3 {
		return false
	}

	for _, r := range value[3:] {
		if r < '0' || r > '9' {
			return false
		}
	}

	return true
}

func (c *CPUCollector) readLoadAverage() (float64, float64, float64, error) {
	path := filepath.Join(c.procRoot, "loadavg")

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("read %s: %w", path, err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf(
			"parse %s: expected at least 3 fields",
			path,
		)
	}

	load1, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse load1: %w", err)
	}

	load5, err := strconv.ParseFloat(fields[1], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse load5: %w", err)
	}

	load15, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("parse load15: %w", err)
	}

	return load1, load5, load15, nil
}
