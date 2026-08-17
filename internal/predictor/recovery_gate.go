package predictor

// RecoveryDecision represents the recommendation produced after
// applying confidence and signal-consensus safety gates.
type RecoveryDecision struct {
	Recommendation Recommendation

	Reason string
}

// EvaluateRecoveryDecision determines the safe recovery recommendation.
//
// Risk score alone is not sufficient for reboot recommendations.
// Multiple independent degradation signals and sufficient historical
// confidence are required.
func EvaluateRecoveryDecision(
	level RiskLevel,
	confidence float64,
	consensus SignalConsensus,
) RecoveryDecision {

	switch level {

	case RiskLow:

		return RecoveryDecision{
			Recommendation: ActionMonitor,
		}

	case RiskMedium:

		return RecoveryDecision{
			Recommendation: ActionInvestigate,
		}

	case RiskHigh:

		if confidence >= 70 &&
			consensus.ActiveSignals >= 2 {

			return RecoveryDecision{
				Recommendation: ActionRebootAdvised,

				Reason: "multiple independent degradation signals confirm reboot risk",
			}
		}

		return RecoveryDecision{
			Recommendation: ActionInvestigate,

			Reason: "high risk detected but recovery consensus is insufficient",
		}

	case RiskCritical:

		// AUTO_RECOVERY remains only a recommendation.
		// No reboot execution exists at this stage.
		if confidence >= 90 &&
			consensus.ActiveSignals >= 3 &&
			consensus.PersistentSignals >= 2 &&
			consensus.KernelEvidence {

			return RecoveryDecision{
				Recommendation: ActionAutoRecovery,

				Reason: "critical multi-signal degradation satisfies automatic recovery gate",
			}
		}

		if confidence >= 70 &&
			consensus.ActiveSignals >= 2 {

			return RecoveryDecision{
				Recommendation: ActionRebootAdvised,

				Reason: "critical risk confirmed by multiple independent signals",
			}
		}

		return RecoveryDecision{
			Recommendation: ActionInvestigate,

			Reason: "critical score detected but recovery safety gate is not satisfied",
		}
	}

	return RecoveryDecision{
		Recommendation: ActionMonitor,
	}
}
