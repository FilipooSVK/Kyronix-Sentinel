package linux

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"kyronix/sentinel/internal/domain"
)

const (
	defaultProcRoot      = "/proc"
	defaultOSReleasePath = "/etc/os-release"
)

// HostCollector collects host information from Linux.
type HostCollector struct {
	procRoot      string
	osReleasePath string
	hostname      func() (string, error)
	architecture  func() string
}

// NewHostCollector creates a Linux host collector using the real system
// interfaces.
func NewHostCollector() *HostCollector {
	return &HostCollector{
		procRoot:      defaultProcRoot,
		osReleasePath: defaultOSReleasePath,
		hostname:      os.Hostname,
		architecture: func() string {
			return runtime.GOARCH
		},
	}
}

// CollectHost collects general Linux host information.
//
// Hostname and uptime are considered core host information and preserve
// the original collector error contract.
//
// Extended metadata is best-effort. Failure to read optional metadata
// must not make the complete host collector unavailable.
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

	stats := domain.HostStats{
		Hostname:     hostname,
		Uptime:       uptime,
		Architecture: runtime.GOARCH,
	}

	if c.architecture != nil {
		stats.Architecture = c.architecture()
	}

	if kernelVersion, err := c.readKernelVersion(); err == nil {
		stats.KernelVersion = kernelVersion
	}

	if osRelease, err := c.readOSRelease(); err == nil {
		stats.OSID = osRelease["ID"]
		stats.OSName = osRelease["PRETTY_NAME"]

		if stats.OSName == "" {
			stats.OSName = osRelease["NAME"]
		}

		stats.OSVersion = osRelease["VERSION_ID"]

		if stats.OSVersion == "" {
			stats.OSVersion = osRelease["VERSION"]
		}
	}

	return stats, nil
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
		return 0, fmt.Errorf(
			"parse %s uptime value %q: %w",
			path,
			fields[0],
			err,
		)
	}

	if seconds < 0 {
		return 0, fmt.Errorf("parse %s: negative uptime value", path)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

func (c *HostCollector) readKernelVersion() (string, error) {
	path := filepath.Join(
		c.procRoot,
		"sys",
		"kernel",
		"osrelease",
	)

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", path, err)
	}

	version := strings.TrimSpace(string(data))
	if version == "" {
		return "", fmt.Errorf("parse %s: kernel version missing", path)
	}

	return version, nil
}

func (c *HostCollector) readOSRelease() (map[string]string, error) {
	path := c.osReleasePath
	if path == "" {
		path = defaultOSReleasePath
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	values := make(map[string]string)

	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			continue
		}

		value = strings.Trim(value, `"'`)

		values[key] = value
	}

	return values, nil
}
