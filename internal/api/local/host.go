package local

import "time"

// Host represents current Sentinel host information exposed by the local API.
type Host struct {
	Hostname string `json:"hostname"`

	Uptime time.Duration `json:"uptime"`

	Architecture string `json:"architecture"`

	OSID string `json:"os_id"`

	OSName string `json:"os_name"`

	OSVersion string `json:"os_version"`

	KernelVersion string `json:"kernel_version"`

	Environment string `json:"environment"`

	Role string `json:"role"`
}

// UpdateHost changes runtime host information.
func (s *Server) UpdateHost(
	host Host,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.host = host
}
