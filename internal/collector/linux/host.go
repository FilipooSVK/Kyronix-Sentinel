package linux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"kyronix/sentinel/internal/domain"
)

const defaultProcRoot = "/proc"

// HostCollector collects host information from Linux.
type HostCollector struct {
	procRoot string
	hostname func() (string, error)
}

// NewHostCollector creates a Linux host collector using the real system
// interfaces.
func NewHostCollector() *HostCollector {
	return &HostCollector{
		procRoot: defaultProcRoot,
		hostname: os.Hostname,
	}
}

// CollectHost collects hostname and host uptime.
func (c *HostCollector) CollectHost(ctx context.Context) (domain.HostStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.HostStats{}, err
	}

	hostname, err := c.hostname()
	if err != nil {
		return domain.HostStats{}, fmt.Errorf("read hostname: %w", err)
	}

	uptime, err := c.readUptime()
	if err != nil {
		return domain.HostStats{}, err
	}

	return domain.HostStats{
		Hostname: hostname,
		Uptime:   uptime,
	}, nil
}

func (c *HostCollector) readUptime() (time.Duration, error) {
	path := filepath.Join(c.procRoot, "uptime")

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", path, err)
	}

	fields := strings.Fields(string(data))
	if len(fields) < 1 {
		return 0, fmt.Errorf("parse %s: uptime value missing", path)
	}

	seconds, err := strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s uptime value %q: %w", path, fields[0], err)
	}

	if seconds < 0 {
		return 0, fmt.Errorf("parse %s: negative uptime value", path)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}
