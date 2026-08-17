package app

import "kyronix/sentinel/internal/api/local"

// UpdatePrediction updates local API prediction state.
func (d *Daemon) UpdatePrediction() {

	prediction := d.engine.LastPrediction()

	d.statusServer.UpdatePrediction(
		local.Prediction{
			Risk: string(
				prediction.Risk,
			),

			Score: prediction.Score,

			Confidence: prediction.Confidence,

			Recommendation: string(
				prediction.Recommendation,
			),

			Reasons: prediction.Reasons,

			ActiveSignals: prediction.ActiveSignals,

			PersistentSignals: prediction.PersistentSignals,

			KernelEvidence: prediction.KernelEvidence,

			Signals: prediction.Signals,
		},
	)
}
