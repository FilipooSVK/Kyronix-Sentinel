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
	defaultProcRoot             = "/proc"
	defaultOSReleasePath        = "/etc/os-release"
	defaultSystemdContainerPath = "/run/systemd/container"
	defaultDockerEnvPath        = "/.dockerenv"
	defaultDMIProductNamePath   = "/sys/class/dmi/id/product_name"
	defaultProxmoxMembersPath   = "/etc/pve/.members"
)

// HostCollector collects host information from Linux.
type HostCollector struct {
	procRoot string

	osReleasePath string

	systemdContainerPath string

	dockerEnvPath string

	dmiProductNamePath string

	proxmoxMembersPath string

	hostname func() (string, error)

	architecture func() string
}

// NewHostCollector creates a Linux host collector using the real system
// interfaces.
func NewHostCollector() *HostCollector {
	return &HostCollector{
		procRoot:             defaultProcRoot,
		osReleasePath:        defaultOSReleasePath,
		systemdContainerPath: defaultSystemdContainerPath,
		dockerEnvPath:        defaultDockerEnvPath,
		dmiProductNamePath:   defaultDMIProductNamePath,
		proxmoxMembersPath:   defaultProxmoxMembersPath,

		hostname: os.Hostname,

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
func (c *HostCollector) CollectHost(
	ctx context.Context,
) (domain.HostStats, error) {
	if err := ctx.Err(); err != nil {
		return domain.HostStats{}, err
	}

	hostname, err := c.hostname()
	if err != nil {
		return domain.HostStats{}, fmt.Errorf(
			"read hostname: %w",
			err,
		)
	}

	uptime, err := c.readUptime()
	if err != nil {
		return domain.HostStats{}, err
	}

	stats := domain.HostStats{
		Hostname:       hostname,
		Uptime:         uptime,
		Architecture:   runtime.GOARCH,
		Virtualization: c.detectVirtualization(),
		Platform:       c.detectPlatform(),
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
	path := filepath.Join(
		c.procRoot,
		"uptime",
	)

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf(
			"read %s: %w",
			path,
			err,
		)
	}

	fields := strings.Fields(
		string(data),
	)

	if len(fields) < 1 {
		return 0, fmt.Errorf(
			"parse %s: uptime value missing",
			path,
		)
	}

	seconds, err := strconv.ParseFloat(
		fields[0],
		64,
	)

	if err != nil {
		return 0, fmt.Errorf(
			"parse %s uptime value %q: %w",
			path,
			fields[0],
			err,
		)
	}

	if seconds < 0 {
		return 0, fmt.Errorf(
			"parse %s: negative uptime value",
			path,
		)
	}

	return time.Duration(
		seconds * float64(time.Second),
	), nil
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
		return "", fmt.Errorf(
			"read %s: %w",
			path,
			err,
		)
	}

	version := strings.TrimSpace(
		string(data),
	)

	if version == "" {
		return "", fmt.Errorf(
			"parse %s: kernel version missing",
			path,
		)
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
		return nil, fmt.Errorf(
			"read %s: %w",
			path,
			err,
		)
	}

	values := make(
		map[string]string,
	)

	for _, line := range strings.Split(
		string(data),
		"\n",
	) {
		line = strings.TrimSpace(line)

		if line == "" ||
			strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(
			line,
			"=",
		)

		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if key == "" {
			continue
		}

		value = strings.Trim(
			value,
			`"'`,
		)

		values[key] = value
	}

	return values, nil
}

// detectVirtualization detects the environment Sentinel is running in.
//
// Detection is read-only and intentionally avoids external commands.
// The result represents the environment visible from the monitored host,
// not necessarily the virtualization technology of its physical parent.
func (c *HostCollector) detectVirtualization() string {
	if container := strings.ToLower(
		readOptionalText(
			c.systemdContainerPath,
		),
	); container != "" {
		switch container {
		case "lxc":
			return "lxc"

		case "docker":
			return "docker"
		}
	}

	if fileExists(
		c.dockerEnvPath,
	) {
		return "docker"
	}

	environPath := ""

	if c.procRoot != "" {
		environPath = filepath.Join(
			c.procRoot,
			"1",
			"environ",
		)
	}

	if data := readOptionalBytes(
		environPath,
	); len(data) > 0 {
		environ := strings.ToLower(
			string(data),
		)

		if strings.Contains(
			environ,
			"container=lxc",
		) {
			return "lxc"
		}

		if strings.Contains(
			environ,
			"container=docker",
		) {
			return "docker"
		}
	}

	cgroupPath := ""

	if c.procRoot != "" {
		cgroupPath = filepath.Join(
			c.procRoot,
			"1",
			"cgroup",
		)
	}

	if cgroup := strings.ToLower(
		readOptionalText(
			cgroupPath,
		),
	); cgroup != "" {
		if strings.Contains(
			cgroup,
			"docker",
		) ||
			strings.Contains(
				cgroup,
				"containerd",
			) {
			return "docker"
		}

		if strings.Contains(
			cgroup,
			"lxc",
		) {
			return "lxc"
		}
	}

	productName := strings.ToLower(
		readOptionalText(
			c.dmiProductNamePath,
		),
	)

	if strings.Contains(
		productName,
		"kvm",
	) ||
		strings.Contains(
			productName,
			"qemu",
		) {
		return "kvm"
	}

	return "bare-metal"
}

// detectPlatform detects known infrastructure platforms.
//
// Platform is intentionally separate from virtualization. For example,
// an LXC container is not automatically assumed to run on Proxmox.
func (c *HostCollector) detectPlatform() string {
	if fileExists(
		c.proxmoxMembersPath,
	) {
		return "proxmox"
	}

	return ""
}

func readOptionalText(
	path string,
) string {
	data := readOptionalBytes(
		path,
	)

	if len(data) == 0 {
		return ""
	}

	return strings.TrimSpace(
		string(data),
	)
}

func readOptionalBytes(
	path string,
) []byte {
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(
		path,
	)

	if err != nil {
		return nil
	}

	return data
}

func fileExists(
	path string,
) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(
		path,
	)

	return err == nil
}
