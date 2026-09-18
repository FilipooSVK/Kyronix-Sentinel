package app

import (
	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/domain"
)

// BuildHost creates the local API host response from the latest snapshot.
func BuildHost(
	snapshot domain.Snapshot,
) local.Host {
	return local.Host{
		Hostname: snapshot.Host.Hostname,

		Uptime: snapshot.Host.Uptime,

		Architecture: snapshot.Host.Architecture,

		OSID: snapshot.Host.OSID,

		OSName: snapshot.Host.OSName,

		OSVersion: snapshot.Host.OSVersion,

		KernelVersion: snapshot.Host.KernelVersion,

		Environment: snapshot.Identity.Environment,

		Role: snapshot.Identity.Role,
	}
}
