package domain

import "time"

// Snapshot represents the complete observable state of a host
// at a specific point in time.
//
// Collectors populate the snapshot with raw system facts.
// The analyzer consumes the snapshot but must not modify it.
type Snapshot struct {
	Timestamp time.Time `json:"timestamp"`

	Host     HostStats     `json:"host"`
	CPU      CPUStats      `json:"cpu"`
	Memory   MemoryStats   `json:"memory"`
	Pressure PressureStats `json:"pressure"`
	Disks    []DiskStats   `json:"disks"`
	Kernel   KernelStats   `json:"kernel"`

	Collection CollectionStatus `json:"collection"`
}
