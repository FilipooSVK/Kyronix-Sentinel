package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	UpdateResultSuccess = "success"

	UpdateResultError = "error"
)

// UpdateState represents the most recent automatic update check.
type UpdateState struct {
	LastCheck time.Time `json:"last_check"`

	NextCheck time.Time `json:"next_check"`

	CurrentVersion string `json:"current_version"`

	LatestVersion string `json:"latest_version,omitempty"`

	UpdateAvailable bool `json:"update_available"`

	LastResult string `json:"last_result"`

	LastError string `json:"last_error,omitempty"`
}

// StateStore persists Sentinel update state.
type StateStore struct {
	path string
}

// NewStateStore creates an update state store.
func NewStateStore(
	path string,
) *StateStore {

	return &StateStore{
		path: strings.TrimSpace(
			path,
		),
	}
}

// Path returns the configured persistent state path.
func (s *StateStore) Path() string {

	if s == nil {
		return ""
	}

	return s.path
}

// Load loads the latest persisted update state.
func (s *StateStore) Load() (
	UpdateState,
	error,
) {

	if s == nil ||
		s.path == "" {

		return UpdateState{}, fmt.Errorf(
			"update state path is empty",
		)
	}

	data, err := os.ReadFile(
		s.path,
	)

	if err != nil {
		return UpdateState{}, err
	}

	decoder := json.NewDecoder(
		bytes.NewReader(data),
	)

	decoder.DisallowUnknownFields()

	var state UpdateState

	if err := decoder.Decode(
		&state,
	); err != nil {

		return UpdateState{}, fmt.Errorf(
			"decode update state: %w",
			err,
		)
	}

	var extra interface{}

	if err := decoder.Decode(
		&extra,
	); err != io.EOF {

		return UpdateState{}, fmt.Errorf(
			"update state contains trailing data",
		)
	}

	if err := validateUpdateState(
		state,
	); err != nil {

		return UpdateState{}, err
	}

	return state, nil
}

// Save atomically persists update state.
func (s *StateStore) Save(
	state UpdateState,
) error {

	if s == nil ||
		s.path == "" {

		return fmt.Errorf(
			"update state path is empty",
		)
	}

	if err := validateUpdateState(
		state,
	); err != nil {

		return err
	}

	data, err := json.MarshalIndent(
		state,
		"",
		"  ",
	)

	if err != nil {
		return err
	}

	data = append(
		data,
		'\n',
	)

	dir := filepath.Dir(
		s.path,
	)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {

		return fmt.Errorf(
			"create update state directory: %w",
			err,
		)
	}

	tempFile, err := os.CreateTemp(
		dir,
		".update-state-*.tmp",
	)

	if err != nil {
		return err
	}

	tempPath := tempFile.Name()

	cleanup := func() {

		tempFile.Close()

		os.Remove(
			tempPath,
		)
	}

	if err := tempFile.Chmod(
		0644,
	); err != nil {

		cleanup()

		return err
	}

	if _, err := tempFile.Write(
		data,
	); err != nil {

		cleanup()

		return err
	}

	if err := tempFile.Sync(); err != nil {

		cleanup()

		return err
	}

	if err := tempFile.Close(); err != nil {

		os.Remove(
			tempPath,
		)

		return err
	}

	if err := os.Rename(
		tempPath,
		s.path,
	); err != nil {

		os.Remove(
			tempPath,
		)

		return err
	}

	if err := syncDirectory(
		dir,
	); err != nil {

		return fmt.Errorf(
			"sync update state directory: %w",
			err,
		)
	}

	return nil
}

func validateUpdateState(
	state UpdateState,
) error {

	if state.LastCheck.IsZero() {

		return fmt.Errorf(
			"update state last_check is empty",
		)
	}

	if state.NextCheck.IsZero() {

		return fmt.Errorf(
			"update state next_check is empty",
		)
	}

	if state.NextCheck.Before(
		state.LastCheck,
	) {

		return fmt.Errorf(
			"update state next_check precedes last_check",
		)
	}

	if !IsValidVersion(
		state.CurrentVersion,
	) {

		return fmt.Errorf(
			"invalid update state current version: %s",
			state.CurrentVersion,
		)
	}

	if state.LatestVersion != "" &&
		!IsValidVersion(
			state.LatestVersion,
		) {

		return fmt.Errorf(
			"invalid update state latest version: %s",
			state.LatestVersion,
		)
	}

	switch state.LastResult {

	case UpdateResultSuccess:

		if state.LatestVersion == "" {

			return fmt.Errorf(
				"successful update state has no latest version",
			)
		}

	case UpdateResultError:

		if strings.TrimSpace(
			state.LastError,
		) == "" {

			return fmt.Errorf(
				"failed update state has no error",
			)
		}

	default:

		return fmt.Errorf(
			"invalid update state result: %s",
			state.LastResult,
		)
	}

	return nil
}
