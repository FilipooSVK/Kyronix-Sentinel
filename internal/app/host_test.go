package app

import (
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
)

func TestBuildHost(
	t *testing.T,
) {
	snapshot := domain.Snapshot{
		Host: domain.HostStats{
			Hostname:      "kyronix-stratus-os",
			Uptime:        7 * time.Hour,
			Architecture:  "arm64",
			OSID:          "debian",
			OSName:        "Debian GNU/Linux 13 (trixie)",
			OSVersion:     "13",
			KernelVersion: "6.6.25-v8+",
		},

		Identity: domain.HostIdentity{
			Environment: "production",
			Role:        "ntp-appliance",
		},
	}

	host := BuildHost(snapshot)

	if host.Hostname != "kyronix-stratus-os" {
		t.Fatalf(
			"unexpected hostname: %s",
			host.Hostname,
		)
	}

	if host.Architecture != "arm64" {
		t.Fatalf(
			"unexpected architecture: %s",
			host.Architecture,
		)
	}

	if host.Environment != "production" {
		t.Fatalf(
			"unexpected environment: %s",
			host.Environment,
		)
	}

	if host.Role != "ntp-appliance" {
		t.Fatalf(
			"unexpected role: %s",
			host.Role,
		)
	}
}
