package predictor

import (
	"time"

	"kyronix/sentinel/internal/history"
)

const (
	kernelEventWindow       = 15 * time.Minute
	persistenceSampleWindow = 8
)

// Predictor evaluates historical Sentinel state.
type Predictor struct{}

// New creates predictive engine.
func New() *Predictor {

	return &Predictor{}
}

// Evaluate evaluates historical snapshots and produces prediction.
func (p *Predictor) Evaluate(
	entries []history.SnapshotEntry,
) Prediction {

	memoryTrend := AnalyzeMemoryTrend(
		entries,
	)

	cpuPressureTrend := AnalyzeCPUPressureTrend(
		entries,
	)

	memoryPressureTrend := AnalyzeMemoryPressureTrend(
		entries,
	)

	ioPressureTrend := AnalyzeIOPressureTrend(
		entries,
	)

	healthTrend := AnalyzeHealthTrend(
		entries,
	)

	trends := []Trend{
		memoryTrend,
		cpuPressureTrend,
		memoryPressureTrend,
		ioPressureTrend,
		healthTrend,
	}

	// Base risk from current values and trends.
	assessment := CalculateRisk(
		trends,
	)

	// Recent kernel incidents.
	kernelEvents := AnalyzeKernelEventsWindow(
		entries,
		kernelEventWindow,
	)

	assessment = ApplyKernelRisk(
		assessment,
		kernelEvents,
	)

	// Persistent degradation.
	persistence := AnalyzePersistence(
		entries,
		persistenceSampleWindow,
	)

	assessment = ApplyPersistenceRisk(
		assessment,
		persistence,
	)

	// Historical confidence.
	confidence := 0.0

	switch {

	case len(entries) >= 10:
		confidence = 90

	case len(entries) >= 5:
		confidence = 70

	case len(entries) >= 3:
		confidence = 40
	}

	// Independent signal consensus.
	consensus := EvaluateSignalConsensus(
		trends,
		persistence,
		kernelEvents,
	)

	// Recovery safety gate.
	decision := EvaluateRecoveryDecision(
		assessment.Level,
		confidence,
		consensus,
	)

	reasons := append(
		[]string{},
		assessment.Reasons...,
	)

	if decision.Reason != "" {

		reasons = append(
			reasons,
			decision.Reason,
		)
	}

	return Prediction{
		Risk:           assessment.Level,
		Score:          assessment.Score,
		Confidence:     confidence,
		Reasons:        reasons,
		Trends:         trends,
		Recommendation: decision.Recommendation,

		ActiveSignals: consensus.ActiveSignals,

		PersistentSignals: consensus.PersistentSignals,

		KernelEvidence: consensus.KernelEvidence,

		Signals: consensus.Signals,
	}
}
