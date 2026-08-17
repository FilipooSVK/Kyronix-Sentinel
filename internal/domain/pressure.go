package domain

// PressureSample represents one Linux PSI pressure measurement window.
//
// Avg10, Avg60 and Avg300 represent the percentage of wall-clock time
// during which tasks experienced resource pressure.
//
// TotalUS is the cumulative stall time reported by the kernel in microseconds.
type PressureSample struct {
	Avg10   float64 `json:"avg10"`
	Avg60   float64 `json:"avg60"`
	Avg300  float64 `json:"avg300"`
	TotalUS uint64  `json:"total_us"`
}

// ResourcePressureStats represents Linux PSI data for one resource.
//
// Linux exposes "some" pressure for CPU, memory and IO.
// A "full" line may also be present. For system-wide CPU PSI it is
// typically zero, but Sentinel preserves the value instead of discarding it.
type ResourcePressureStats struct {
	Some PressureSample `json:"some"`
	Full PressureSample `json:"full"`
}

// PressureStats represents Linux Pressure Stall Information for the host.
type PressureStats struct {
	CPU    ResourcePressureStats `json:"cpu"`
	Memory ResourcePressureStats `json:"memory"`
	IO     ResourcePressureStats `json:"io"`
}
