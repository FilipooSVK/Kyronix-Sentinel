package app

import (
	"testing"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
)

type fakeCollector struct{}

func (fakeCollector) Collect() domain.Snapshot {

	return domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 50,
		},
	}
}

type fakeAnalyzer struct{}

func (fakeAnalyzer) Analyze(
	domain.Snapshot,
) domain.HealthResult {

	return domain.HealthResult{
		HealthScore: 90,
		FreezeRisk:  domain.RiskLow,
	}
}

func TestEngineRunOnce(t *testing.T) {

	store := history.NewStore(10)

	engine := NewEngine(
		fakeCollector{},
		fakeAnalyzer{},
		store,
	)

	result := engine.RunOnce()

	if result.HealthScore != 90 {
		t.Errorf(
			"health score mismatch: got %d",
			result.HealthScore,
		)
	}

	results := engine.History()

	if len(results) != 1 {
		t.Errorf(
			"history length mismatch: got %d",
			len(results),
		)
	}
}

func TestEngineMultipleRuns(t *testing.T) {

	store := history.NewStore(10)

	engine := NewEngine(
		fakeCollector{},
		fakeAnalyzer{},
		store,
	)

	engine.RunOnce()
	engine.RunOnce()
	engine.RunOnce()

	results := engine.History()

	if len(results) != 3 {
		t.Errorf(
			"history length mismatch: got %d",
			len(results),
		)
	}
}
