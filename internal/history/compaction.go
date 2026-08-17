package history

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
)

// CompactSnapshots rewrites persistent history so only the newest
// entries up to limit remain on disk.
//
// The rewrite is performed through a temporary file followed by an
// atomic rename so an interrupted compaction does not corrupt the
// active history file.
func CompactSnapshots(
	path string,
	entries []SnapshotEntry,
	limit int,
) error {

	if limit > 0 &&
		len(entries) > limit {

		entries = entries[len(entries)-limit:]
	}

	dir := filepath.Dir(
		path,
	)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {

		return err
	}

	tempFile, err := os.CreateTemp(
		dir,
		".history-*.tmp",
	)

	if err != nil {
		return err
	}

	tempPath := tempFile.Name()

	cleanup := func() {
		tempFile.Close()
		os.Remove(tempPath)
	}

	writer := bufio.NewWriter(
		tempFile,
	)

	encoder := json.NewEncoder(
		writer,
	)

	for _, entry := range entries {

		if err := encoder.Encode(
			entry,
		); err != nil {

			cleanup()
			return err
		}
	}

	if err := writer.Flush(); err != nil {

		cleanup()
		return err
	}

	if err := tempFile.Sync(); err != nil {

		cleanup()
		return err
	}

	if err := tempFile.Chmod(
		0640,
	); err != nil {

		cleanup()
		return err
	}

	if err := tempFile.Close(); err != nil {

		os.Remove(tempPath)
		return err
	}

	if err := os.Rename(
		tempPath,
		path,
	); err != nil {

		os.Remove(tempPath)
		return err
	}

	return nil
}
