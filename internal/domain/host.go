package domain

import "time"

// HostStats represents general host information.
type HostStats struct {
	Hostname string        `json:"hostname"`
	Uptime   time.Duration `json:"uptime"`
}
