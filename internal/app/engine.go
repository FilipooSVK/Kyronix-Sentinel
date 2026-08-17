package app

import (
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
	"kyronix/sentinel/internal/predictor"
)

const (
	minHistoryCompactionInterval = 10
	maxHistoryCompactionInterval = 100
)

// SnapshotCollector collects one complete host snapshot.
type SnapshotCollector interface {
	Collect() domain.Snapshot
}

// Analyzer evaluates one host snapshot.
type Analyzer interface {
	Analyze(snapshot domain.Snapshot) domain.HealthResult
}

// Engine coordinates collection, analysis, history and prediction.
type Engine struct {
	collector SnapshotCollector

	analyzer Analyzer

	history *history.Store

	predictor *predictor.Predictor

	lastSnapshot domain.Snapshot

	lastPrediction predictor.Prediction

	historyPath string

	historyLimit int

	persistenceWrites int

	lastPersistenceError error
}

// NewEngine creates Sentinel runtime engine.
func NewEngine(
	collector SnapshotCollector,
	analyzer Analyzer,
	store *history.Store,
) *Engine {

	return &Engine{
		collector: collector,

		analyzer: analyzer,

		history: store,

		predictor: predictor.New(),
	}
}

// ConfigureHistoryPersistence enables persistent snapshot history.
//
// Existing persistent history is loaded into the in-memory store
// before new evaluations are performed.
//
// The persistent file is also compacted during startup so disk
// retention immediately matches the configured history limit.
func (e *Engine) ConfigureHistoryPersistence(
	path string,
	limit int,
) error {

	entries, err := history.LoadSnapshots(
		path,
		limit,
	)

	if err != nil {
		return err
	}

	e.history.RestoreSnapshots(
		entries,
	)

	if limit > 0 {

		if err := history.CompactSnapshots(
			path,
			entries,
			limit,
		); err != nil {

			return err
		}
	}

	e.historyPath = path

	e.historyLimit = limit

	e.persistenceWrites = 0

	e.lastPersistenceError = nil

	return nil
}

// RunOnce executes one complete collection and prediction cycle.
func (e *Engine) RunOnce() domain.HealthResult {

	snapshot := e.collector.Collect()

	e.lastSnapshot = snapshot

	result := e.analyzer.Analyze(
		snapshot,
	)

	entry := history.SnapshotEntry{
		Timestamp: time.Now(),

		Snapshot: snapshot,

		Health: result,
	}

	e.history.Add(
		result,
	)

	e.history.AddSnapshot(
		entry,
	)

	e.persistSnapshot(
		entry,
	)

	e.lastPrediction = e.predictor.Evaluate(
		e.history.GetSnapshots(),
	)

	return result
}

// persistSnapshot writes one snapshot to persistent history and
// periodically compacts the JSONL file to enforce retention.
func (e *Engine) persistSnapshot(
	entry history.SnapshotEntry,
) {

	if e.historyPath == "" {
		return
	}

	if err := history.AppendSnapshot(
		e.historyPath,
		entry,
	); err != nil {

		e.lastPersistenceError = err

		return
	}

	e.persistenceWrites++

	e.lastPersistenceError = nil

	if e.historyLimit <= 0 {
		return
	}

	if e.persistenceWrites <
		historyCompactionInterval(
			e.historyLimit,
		) {

		return
	}

	if err := history.CompactSnapshots(
		e.historyPath,
		e.history.GetSnapshots(),
		e.historyLimit,
	); err != nil {

		e.lastPersistenceError = err

		return
	}

	e.persistenceWrites = 0
}

// historyCompactionInterval determines how often persistent history
// should be compacted.
//
// Large histories are compacted at most every 100 writes.
// Small histories use a shorter interval while avoiding a rewrite
// on every collection cycle.
func historyCompactionInterval(
	limit int,
) int {

	if limit <= 0 {
		return maxHistoryCompactionInterval
	}

	interval := limit / 10

	if interval < minHistoryCompactionInterval {
		return minHistoryCompactionInterval
	}

	if interval > maxHistoryCompactionInterval {
		return maxHistoryCompactionInterval
	}

	return interval
}

// History returns in-memory health result history.
func (e *Engine) History() []domain.HealthResult {

	return e.history.GetAll()
}

// LastSnapshot returns the latest collected snapshot.
func (e *Engine) LastSnapshot() domain.Snapshot {

	return e.lastSnapshot
}

// LastPrediction returns the latest predictive assessment.
func (e *Engine) LastPrediction() predictor.Prediction {

	return e.lastPrediction
}

// LastPersistenceError returns the most recent persistent history
// error, if one occurred.
func (e *Engine) LastPersistenceError() error {

	return e.lastPersistenceError
}
