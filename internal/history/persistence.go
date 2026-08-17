package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

const maxHistoryLineSize = 4 * 1024 * 1024

// AppendSnapshot appends one snapshot entry to persistent JSONL history.
func AppendSnapshot(
	path string,
	entry SnapshotEntry,
) error {

	dir := filepath.Dir(path)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {
		return err
	}

	file, err := os.OpenFile(
		path,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0640,
	)

	if err != nil {
		return err
	}

	defer file.Close()

	encoder := json.NewEncoder(
		file,
	)

	return encoder.Encode(
		entry,
	)
}

// LoadSnapshots loads persistent JSONL history.
//
// Invalid lines are ignored so an incomplete final write caused by
// an unexpected shutdown does not prevent Sentinel from starting.
func LoadSnapshots(
	path string,
	limit int,
) ([]SnapshotEntry, error) {

	file, err := os.Open(
		path,
	)

	if err != nil {

		if os.IsNotExist(err) {
			return []SnapshotEntry{}, nil
		}

		return nil, err
	}

	defer file.Close()

	entries := []SnapshotEntry{}

	scanner := bufio.NewScanner(
		file,
	)

	scanner.Buffer(
		make([]byte, 64*1024),
		maxHistoryLineSize,
	)

	for scanner.Scan() {

		line := scanner.Bytes()

		if len(line) == 0 {
			continue
		}

		var entry SnapshotEntry

		if err := json.Unmarshal(
			line,
			&entry,
		); err != nil {

			// Ignore malformed or partially written records.
			continue
		}

		entries = append(
			entries,
			entry,
		)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if limit > 0 &&
		len(entries) > limit {

		entries = entries[len(entries)-limit:]
	}

	return entries, nil
}
