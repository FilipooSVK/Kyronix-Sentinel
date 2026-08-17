package predictor

import "testing"

func TestApplyPersistenceRiskCPU(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 20,
		Level: RiskLow,
	}

	report := PersistenceReport{
		CPUPressure: PersistenceSignal{
			Metric:  "cpu_pressure",
			Hits:    6,
			Samples: 8,
			Ratio:   0.75,
		},
	}

	assessment := ApplyPersistenceRisk(
		base,
		report,
	)

	if assessment.Score != 35 {
		t.Fatalf(
			"expected score 35, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskMedium {
		t.Fatalf(
			"expected MEDIUM risk, got %s",
			assessment.Level,
		)
	}
}

func TestApplyPersistenceRiskMultipleSignals(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 25,
		Level: RiskLow,
	}

	report := PersistenceReport{
		MemoryUtilization: PersistenceSignal{
			Hits:    8,
			Samples: 8,
			Ratio:   1,
		},

		MemoryPressure: PersistenceSignal{
			Hits:    7,
			Samples: 8,
			Ratio:   0.875,
		},

		HealthDegradation: PersistenceSignal{
			Hits:    6,
			Samples: 8,
			Ratio:   0.75,
		},
	}

	assessment := ApplyPersistenceRisk(
		base,
		report,
	)

	// Base 25
	// Memory utilization +15
	// Memory pressure +20
	// Health degradation +10
	// Total 70
	if assessment.Score != 70 {
		t.Fatalf(
			"expected score 70, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskHigh {
		t.Fatalf(
			"expected HIGH risk, got %s",
			assessment.Level,
		)
	}
}

func TestApplyPersistenceRiskIgnoresShortHistory(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 20,
		Level: RiskLow,
	}

	report := PersistenceReport{
		IOPressure: PersistenceSignal{
			Hits:    4,
			Samples: 4,
			Ratio:   1,
		},
	}

	assessment := ApplyPersistenceRisk(
		base,
		report,
	)

	if assessment.Score != 20 {
		t.Fatalf(
			"expected unchanged score 20, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskLow {
		t.Fatalf(
			"expected LOW risk, got %s",
			assessment.Level,
		)
	}
}

func TestApplyPersistenceRiskRequiresRatio(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 10,
		Level: RiskLow,
	}

	report := PersistenceReport{
		IOPressure: PersistenceSignal{
			Hits:    5,
			Samples: 8,
			Ratio:   0.625,
		},
	}

	assessment := ApplyPersistenceRisk(
		base,
		report,
	)

	if assessment.Score != 10 {
		t.Fatalf(
			"expected unchanged score 10, got %d",
			assessment.Score,
		)
	}
}

func TestApplyPersistenceRiskCapsScore(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 90,
		Level: RiskCritical,
	}

	report := PersistenceReport{
		MemoryPressure: PersistenceSignal{
			Hits:    8,
			Samples: 8,
			Ratio:   1,
		},

		IOPressure: PersistenceSignal{
			Hits:    8,
			Samples: 8,
			Ratio:   1,
		},
	}

	assessment := ApplyPersistenceRisk(
		base,
		report,
	)

	if assessment.Score != 100 {
		t.Fatalf(
			"expected capped score 100, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskCritical {
		t.Fatalf(
			"expected CRITICAL risk, got %s",
			assessment.Level,
		)
	}
}
