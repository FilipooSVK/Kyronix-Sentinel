package domain

import "time"

// FreezeRisk represents predicted freeze probability.
type FreezeRisk string

const (
	RiskLow      FreezeRisk = "LOW"
	RiskMedium   FreezeRisk = "MEDIUM"
	RiskHigh     FreezeRisk = "HIGH"
	RiskCritical FreezeRisk = "CRITICAL"
)

// AssessmentQuality describes completeness of health assessment.
type AssessmentQuality string

const (
	AssessmentComplete AssessmentQuality = "COMPLETE"
	AssessmentPartial  AssessmentQuality = "PARTIAL"
	AssessmentFailed   AssessmentQuality = "FAILED"
)

// HealthResult represents analyzer output.
//
// HealthScore represents current host health.
// FreezeRiskScore represents predicted instability risk.
type HealthResult struct {
	Timestamp time.Time `json:"timestamp"`

	HealthScore     int `json:"health_score"`
	FreezeRiskScore int `json:"freeze_risk_score"`

	FreezeRisk FreezeRisk `json:"freeze_risk"`

	Quality AssessmentQuality `json:"quality"`

	Findings []Finding `json:"findings"`
}
