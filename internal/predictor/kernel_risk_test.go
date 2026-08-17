package predictor

import "testing"

func TestApplyKernelRiskSystemOOM(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 0,
		Level: RiskLow,
	}

	events := KernelEvents{
		SystemOOMKills: 1,
	}

	assessment := ApplyKernelRisk(
		base,
		events,
	)

	if assessment.Score != 40 {
		t.Fatalf(
			"expected score 40, got %d",
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

func TestApplyKernelRiskRepeatedSystemOOM(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 0,
		Level: RiskLow,
	}

	events := KernelEvents{
		SystemOOMKills: 3,
	}

	assessment := ApplyKernelRisk(
		base,
		events,
	)

	if assessment.Score != 60 {
		t.Fatalf(
			"expected score 60, got %d",
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

func TestApplyKernelRiskDoesNotDoubleCountCgroupOOM(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 0,
		Level: RiskLow,
	}

	events := KernelEvents{
		CgroupOOMKills:  1,
		CgroupOOMEvents: 4,
	}

	assessment := ApplyKernelRisk(
		base,
		events,
	)

	if assessment.Score != 30 {
		t.Fatalf(
			"expected score 30, got %d",
			assessment.Score,
		)
	}
}

func TestApplyKernelRiskFilesystemErrors(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 0,
		Level: RiskLow,
	}

	events := KernelEvents{
		FilesystemErrors: 4,
	}

	assessment := ApplyKernelRisk(
		base,
		events,
	)

	if assessment.Score != 30 {
		t.Fatalf(
			"expected score 30, got %d",
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

func TestApplyKernelRiskCombinedSignals(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 30,
		Level: RiskMedium,
		Reasons: []string{
			"memory growth detected",
		},
	}

	events := KernelEvents{
		SystemOOMKills:   1,
		FilesystemErrors: 4,
	}

	assessment := ApplyKernelRisk(
		base,
		events,
	)

	if assessment.Score != 100 {
		t.Fatalf(
			"expected score 100, got %d",
			assessment.Score,
		)
	}

	if assessment.Level != RiskCritical {
		t.Fatalf(
			"expected CRITICAL risk, got %s",
			assessment.Level,
		)
	}

	if len(assessment.Reasons) != 3 {
		t.Fatalf(
			"expected 3 reasons, got %d",
			len(assessment.Reasons),
		)
	}
}

func TestApplyKernelRiskNoEvents(
	t *testing.T,
) {

	base := RiskAssessment{
		Score: 20,
		Level: RiskLow,
	}

	assessment := ApplyKernelRisk(
		base,
		KernelEvents{},
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
