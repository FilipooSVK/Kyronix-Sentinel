package domain

import (
	"encoding/json"
	"testing"
	"time"
)

func TestHealthResultJSON(t *testing.T) {
	result := HealthResult{
		Timestamp:       time.Date(2026, 8, 16, 8, 0, 0, 0, time.UTC),
		HealthScore:     92,
		FreezeRiskScore: 12,
		FreezeRisk:      RiskLow,
		Quality:         AssessmentComplete,
		Findings:        []Finding{},
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal HealthResult: %v", err)
	}

	var decoded HealthResult

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal HealthResult: %v", err)
	}

	if decoded.HealthScore != result.HealthScore {
		t.Errorf(
			"HealthScore mismatch: got %d, want %d",
			decoded.HealthScore,
			result.HealthScore,
		)
	}

	if decoded.FreezeRisk != RiskLow {
		t.Errorf(
			"FreezeRisk mismatch: got %q, want %q",
			decoded.FreezeRisk,
			RiskLow,
		)
	}

	if decoded.Quality != AssessmentComplete {
		t.Errorf(
			"Quality mismatch: got %q, want %q",
			decoded.Quality,
			AssessmentComplete,
		)
	}
}
