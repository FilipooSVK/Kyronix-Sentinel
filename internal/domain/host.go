package domain

import "time"

// HostStats represents general host information.
//
// HostStats contains observable facts reported by the operating system.
// Administrator-defined identity such as environment and role is kept
// separate from these system-derived values.
type HostStats struct {
	Hostname      string        `json:"hostname"`
	Uptime        time.Duration `json:"uptime"`
	Architecture  string        `json:"architecture"`
	OSID          string        `json:"os_id"`
	OSName        string        `json:"os_name"`
	OSVersion     string        `json:"os_version"`
	KernelVersion string        `json:"kernel_version"`
}
