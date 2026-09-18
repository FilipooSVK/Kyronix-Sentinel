package domain

// HostIdentity represents administrator-defined host identity.
//
// Unlike HostStats, these values are not discovered from the operating
// system. They describe the operational purpose of the host.
type HostIdentity struct {
	Environment string `json:"environment"`
	Role        string `json:"role"`
}
