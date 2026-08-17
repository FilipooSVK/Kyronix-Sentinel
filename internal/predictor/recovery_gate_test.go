package predictor

import "testing"

func TestRecoveryGateHighRiskSingleSignal(
	t *testing.T,
) {

	consensus := SignalConsensus{
		ActiveSignals: 1,
	}

	decision := EvaluateRecoveryDecision(
		RiskHigh,
		90,
		consensus,
	)

	if decision.Recommendation != ActionInvestigate {

		t.Fatalf(
			"expected INVESTIGATE, got %s",
			decision.Recommendation,
		)
	}
}

func TestRecoveryGateHighRiskConsensus(
	t *testing.T,
) {

	consensus := SignalConsensus{
		ActiveSignals: 2,
	}

	decision := EvaluateRecoveryDecision(
		RiskHigh,
		70,
		consensus,
	)

	if decision.Recommendation != ActionRebootAdvised {

		t.Fatalf(
			"expected REBOOT_ADVISED, got %s",
			decision.Recommendation,
		)
	}
}

func TestRecoveryGateHighRiskLowConfidence(
	t *testing.T,
) {

	consensus := SignalConsensus{
		ActiveSignals: 3,
	}

	decision := EvaluateRecoveryDecision(
		RiskHigh,
		40,
		consensus,
	)

	if decision.Recommendation != ActionInvestigate {

		t.Fatalf(
			"expected INVESTIGATE, got %s",
			decision.Recommendation,
		)
	}
}

func TestRecoveryGateCriticalRebootAdvised(
	t *testing.T,
) {

	consensus := SignalConsensus{
		ActiveSignals:     3,
		PersistentSignals: 2,
		KernelEvidence:    false,
	}

	decision := EvaluateRecoveryDecision(
		RiskCritical,
		90,
		consensus,
	)

	if decision.Recommendation != ActionRebootAdvised {

		t.Fatalf(
			"expected REBOOT_ADVISED, got %s",
			decision.Recommendation,
		)
	}
}

func TestRecoveryGateCriticalAutoRecoveryEligible(
	t *testing.T,
) {

	consensus := SignalConsensus{
		ActiveSignals:     4,
		PersistentSignals: 2,
		KernelEvidence:    true,
	}

	decision := EvaluateRecoveryDecision(
		RiskCritical,
		90,
		consensus,
	)

	if decision.Recommendation != ActionAutoRecovery {

		t.Fatalf(
			"expected AUTO_RECOVERY, got %s",
			decision.Recommendation,
		)
	}
}

func TestEvaluateSignalConsensus(
	t *testing.T,
) {

	trends := []Trend{
		{
			Metric:  "memory",
			Current: 95,
		},
		{
			Metric:  "cpu_pressure_some_avg10",
			Current: 75,
		},
		{
			Metric:  "io_pressure_some_avg10",
			Current: 10,
		},
	}

	consensus := EvaluateSignalConsensus(
		trends,
		PersistenceReport{},
		KernelEvents{
			SystemOOMKills: 1,
		},
	)

	if consensus.ActiveSignals != 3 {

		t.Fatalf(
			"expected 3 active signals, got %d",
			consensus.ActiveSignals,
		)
	}

	if !consensus.KernelEvidence {

		t.Fatal(
			"expected kernel evidence",
		)
	}
}
