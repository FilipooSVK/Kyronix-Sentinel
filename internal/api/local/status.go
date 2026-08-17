package local

import "time"

// CollectorStatus represents one collector state.
type CollectorStatus struct {
	Name string `json:"name"`

	State string `json:"state"`

	Message string `json:"message,omitempty"`

	CollectionMS int64 `json:"collection_ms"`

	LastSuccess *time.Time `json:"last_success,omitempty"`
}

// Diagnostics represents Sentinel diagnostics.
type Diagnostics struct {
	Running bool `json:"running"`

	HealthScore int `json:"health_score"`

	FreezeRisk string `json:"freeze_risk"`

	Version string `json:"version"`

	Collectors []CollectorStatus `json:"collectors"`
}

// Status represents Sentinel runtime status.
type Status struct {
	Running bool `json:"running"`

	HealthScore int `json:"health_score"`

	FreezeRisk string `json:"freeze_risk"`

	Version string `json:"version"`
}

// Prediction represents Sentinel predictive runtime state.
type Prediction struct {
	Risk string `json:"risk"`

	Score int `json:"score"`

	Confidence float64 `json:"confidence"`

	Recommendation string `json:"recommendation"`

	Reasons []string `json:"reasons"`

	ActiveSignals int `json:"active_signals"`

	PersistentSignals int `json:"persistent_signals"`

	KernelEvidence bool `json:"kernel_evidence"`

	Signals []string `json:"signals"`
}
