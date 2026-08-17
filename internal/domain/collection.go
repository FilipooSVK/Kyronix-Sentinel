package domain

import "time"

// CollectorState describes the result of a collector execution.
type CollectorState string

const (
	CollectorOK          CollectorState = "OK"
	CollectorUnavailable CollectorState = "UNAVAILABLE"
	CollectorError       CollectorState = "ERROR"
)

// CollectorStatus represents the execution state of one collector.
type CollectorStatus struct {
	State        CollectorState `json:"state"`
	Message      string         `json:"message,omitempty"`
	LastSuccess  *time.Time     `json:"last_success,omitempty"`
	CollectionMS int64          `json:"collection_ms"`
}

// CollectionStatus contains execution metadata for all collectors.
type CollectionStatus struct {
	CPU      CollectorStatus `json:"cpu"`
	Memory   CollectorStatus `json:"memory"`
	Pressure CollectorStatus `json:"pressure"`
	Disk     CollectorStatus `json:"disk"`
	Kernel   CollectorStatus `json:"kernel"`
	Host     CollectorStatus `json:"host"`
}
