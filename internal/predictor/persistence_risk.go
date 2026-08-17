package predictor

const (
	persistenceMinimumRatio   = 0.75
	persistenceMinimumSamples = 5
)

// ApplyPersistenceRisk adds risk for degradation that persists
// across multiple recent collection cycles.
func ApplyPersistenceRisk(
	assessment RiskAssessment,
	report PersistenceReport,
) RiskAssessment {

	score := assessment.Score

	reasons := append(
		[]string{},
		assessment.Reasons...,
	)

	if report.MemoryUtilization.Persistent(
		persistenceMinimumRatio,
		persistenceMinimumSamples,
	) {

		score += 15

		reasons = append(
			reasons,
			"persistent high memory utilization",
		)
	}

	if report.CPUPressure.Persistent(
		persistenceMinimumRatio,
		persistenceMinimumSamples,
	) {

		score += 15

		reasons = append(
			reasons,
			"persistent CPU pressure",
		)
	}

	if report.MemoryPressure.Persistent(
		persistenceMinimumRatio,
		persistenceMinimumSamples,
	) {

		score += 20

		reasons = append(
			reasons,
			"persistent memory pressure",
		)
	}

	if report.IOPressure.Persistent(
		persistenceMinimumRatio,
		persistenceMinimumSamples,
	) {

		score += 20

		reasons = append(
			reasons,
			"persistent I/O pressure",
		)
	}

	if report.HealthDegradation.Persistent(
		persistenceMinimumRatio,
		persistenceMinimumSamples,
	) {

		score += 10

		reasons = append(
			reasons,
			"persistent health degradation",
		)
	}

	if score > 100 {
		score = 100
	}

	level := RiskLow

	switch {

	case score >= 80:
		level = RiskCritical

	case score >= 60:
		level = RiskHigh

	case score >= 30:
		level = RiskMedium
	}

	return RiskAssessment{
		Score:   score,
		Level:   level,
		Reasons: reasons,
	}
}
