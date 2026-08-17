package domain

// CPUStats represents host CPU utilization and load information.
type CPUStats struct {
	// UsagePercent is nil when CPU utilization cannot yet be calculated,
	// for example during the first collector sample.
	UsagePercent *float64 `json:"usage_percent"`

	CoreCount int `json:"core_count"`

	Load1  float64 `json:"load_1"`
	Load5  float64 `json:"load_5"`
	Load15 float64 `json:"load_15"`
}
