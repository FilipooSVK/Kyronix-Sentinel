package predictor

// SignalConsensus describes independent evidence groups that
// currently indicate host degradation.
type SignalConsensus struct {
	ActiveSignals int

	PersistentSignals int

	KernelEvidence bool

	Signals []string
}

// EvaluateSignalConsensus evaluates independent degradation groups.
//
// A signal becomes active when it is either persistent across recent
// samples or currently severe enough to represent immediate pressure.
func EvaluateSignalConsensus(
	trends []Trend,
	persistence PersistenceReport,
	kernelEvents KernelEvents,
) SignalConsensus {

	consensus := SignalConsensus{
		Signals: []string{},
	}

	memoryTrend := findTrend(
		trends,
		"memory",
	)

	memoryPressureTrend := findTrend(
		trends,
		"memory_pressure_some_avg10",
	)

	cpuPressureTrend := findTrend(
		trends,
		"cpu_pressure_some_avg10",
	)

	ioPressureTrend := findTrend(
		trends,
		"io_pressure_some_avg10",
	)

	healthTrend := findTrend(
		trends,
		"health_score",
	)

	memoryPersistent :=
		persistence.MemoryUtilization.Persistent(
			persistenceMinimumRatio,
			persistenceMinimumSamples,
		) ||
			persistence.MemoryPressure.Persistent(
				persistenceMinimumRatio,
				persistenceMinimumSamples,
			)

	memorySevere :=
		memoryTrend.Current >= 90 ||
			memoryPressureTrend.Current >= 30

	if memoryPersistent || memorySevere {

		consensus.ActiveSignals++

		consensus.Signals = append(
			consensus.Signals,
			"memory",
		)
	}

	if memoryPersistent {
		consensus.PersistentSignals++
	}

	cpuPersistent :=
		persistence.CPUPressure.Persistent(
			persistenceMinimumRatio,
			persistenceMinimumSamples,
		)

	cpuSevere :=
		cpuPressureTrend.Current >= 70

	if cpuPersistent || cpuSevere {

		consensus.ActiveSignals++

		consensus.Signals = append(
			consensus.Signals,
			"cpu",
		)
	}

	if cpuPersistent {
		consensus.PersistentSignals++
	}

	ioPersistent :=
		persistence.IOPressure.Persistent(
			persistenceMinimumRatio,
			persistenceMinimumSamples,
		)

	ioSevere :=
		ioPressureTrend.Current >= 70

	if ioPersistent || ioSevere {

		consensus.ActiveSignals++

		consensus.Signals = append(
			consensus.Signals,
			"io",
		)
	}

	if ioPersistent {
		consensus.PersistentSignals++
	}

	healthPersistent :=
		persistence.HealthDegradation.Persistent(
			persistenceMinimumRatio,
			persistenceMinimumSamples,
		)

	healthSevere :=
		healthTrend.Direction == TrendDecreasing &&
			healthTrend.Current <= 50

	if healthPersistent || healthSevere {

		consensus.ActiveSignals++

		consensus.Signals = append(
			consensus.Signals,
			"health",
		)
	}

	if healthPersistent {
		consensus.PersistentSignals++
	}

	if kernelEvents.HasEvents() {

		consensus.ActiveSignals++

		consensus.KernelEvidence = true

		consensus.Signals = append(
			consensus.Signals,
			"kernel",
		)
	}

	return consensus
}

func findTrend(
	trends []Trend,
	metric string,
) Trend {

	for _, trend := range trends {

		if trend.Metric == metric {
			return trend
		}
	}

	return Trend{
		Metric:    metric,
		Direction: TrendStable,
	}
}
