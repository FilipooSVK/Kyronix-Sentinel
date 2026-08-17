package linux

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kyronix/sentinel/internal/domain"
)

// ErrPressureUnavailable indicates that Linux PSI is not available
// on the running system.
var ErrPressureUnavailable = errors.New("pressure stall information unavailable")

// PressureCollector collects Linux Pressure Stall Information.
type PressureCollector struct {
	procRoot string
}

// NewPressureCollector creates a PSI collector using the real /proc filesystem.
func NewPressureCollector() *PressureCollector {
	return &PressureCollector{
		procRoot: defaultProcRoot,
	}
}

// CollectPressure collects CPU, memory and IO pressure information.
func (c *PressureCollector) CollectPressure(
	ctx context.Context,
) (domain.PressureStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.PressureStats{}, err
	}

	pressureRoot := filepath.Join(c.procRoot, "pressure")

	if _, err := os.Stat(pressureRoot); err != nil {
		if os.IsNotExist(err) {
			return domain.PressureStats{}, ErrPressureUnavailable
		}

		return domain.PressureStats{}, fmt.Errorf(
			"stat %s: %w",
			pressureRoot,
			err,
		)
	}

	cpu, err := c.readResource(ctx, filepath.Join(pressureRoot, "cpu"))
	if err != nil {
		return domain.PressureStats{}, err
	}

	memory, err := c.readResource(ctx, filepath.Join(pressureRoot, "memory"))
	if err != nil {
		return domain.PressureStats{}, err
	}

	ioStats, err := c.readResource(ctx, filepath.Join(pressureRoot, "io"))
	if err != nil {
		return domain.PressureStats{}, err
	}

	return domain.PressureStats{
		CPU:    cpu,
		Memory: memory,
		IO:     ioStats,
	}, nil
}

func (c *PressureCollector) readResource(
	ctx context.Context,
	path string,
) (domain.ResourcePressureStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.ResourcePressureStats{}, err
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return domain.ResourcePressureStats{}, ErrPressureUnavailable
		}

		return domain.ResourcePressureStats{}, fmt.Errorf(
			"open %s: %w",
			path,
			err,
		)
	}
	defer file.Close()

	stats, err := parsePressure(file)
	if err != nil {
		return domain.ResourcePressureStats{}, fmt.Errorf(
			"parse %s: %w",
			path,
			err,
		)
	}

	return stats, nil
}

func parsePressure(r io.Reader) (domain.ResourcePressureStats, error) {
	var result domain.ResourcePressureStats

	var haveSome bool
	var haveFull bool

	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}

		sample, err := parsePressureSample(fields[1:])
		if err != nil {
			return domain.ResourcePressureStats{}, err
		}

		switch fields[0] {
		case "some":
			result.Some = sample
			haveSome = true

		case "full":
			result.Full = sample
			haveFull = true
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.ResourcePressureStats{}, fmt.Errorf(
			"read PSI data: %w",
			err,
		)
	}

	if !haveSome {
		return domain.ResourcePressureStats{}, errors.New(
			"PSI some record missing",
		)
	}

	// Some kernels/resources may not expose a meaningful full record.
	// Leaving it at zero is preferable to failing the complete snapshot.
	if !haveFull {
		result.Full = domain.PressureSample{}
	}

	return result, nil
}

func parsePressureSample(fields []string) (domain.PressureSample, error) {
	values := make(map[string]string)

	for _, field := range fields {
		key, value, ok := strings.Cut(field, "=")
		if !ok {
			return domain.PressureSample{}, fmt.Errorf(
				"invalid PSI field %q",
				field,
			)
		}

		values[key] = value
	}

	required := []string{"avg10", "avg60", "avg300", "total"}

	for _, key := range required {
		if _, ok := values[key]; !ok {
			return domain.PressureSample{}, fmt.Errorf(
				"required PSI field %s missing",
				key,
			)
		}
	}

	avg10, err := strconv.ParseFloat(values["avg10"], 64)
	if err != nil {
		return domain.PressureSample{}, fmt.Errorf(
			"parse avg10: %w",
			err,
		)
	}

	avg60, err := strconv.ParseFloat(values["avg60"], 64)
	if err != nil {
		return domain.PressureSample{}, fmt.Errorf(
			"parse avg60: %w",
			err,
		)
	}

	avg300, err := strconv.ParseFloat(values["avg300"], 64)
	if err != nil {
		return domain.PressureSample{}, fmt.Errorf(
			"parse avg300: %w",
			err,
		)
	}

	total, err := strconv.ParseUint(values["total"], 10, 64)
	if err != nil {
		return domain.PressureSample{}, fmt.Errorf(
			"parse total: %w",
			err,
		)
	}

	return domain.PressureSample{
		Avg10:   avg10,
		Avg60:   avg60,
		Avg300:  avg300,
		TotalUS: total,
	}, nil
}
