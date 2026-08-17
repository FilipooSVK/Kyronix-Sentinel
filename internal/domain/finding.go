package domain

// Severity represents importance of a finding.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Finding represents one analyzer observation.
type Finding struct {
	Severity Severity `json:"severity"`

	Source string `json:"source"`

	Message string `json:"message"`

	Penalty int `json:"penalty"`
}
