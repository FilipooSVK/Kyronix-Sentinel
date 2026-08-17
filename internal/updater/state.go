package updater

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	UpdateResultSuccess = "success"
	UpdateResultError   = "error"

	InstallResultInProgress = "in_progress"
	InstallResultSuccess    = "success"
	InstallResultFailed     = "failed"
)

// UpdateState represents persistent Sentinel update state.
type UpdateState struct {
	LastCheck time.Time `json:"last_check"`

	NextCheck time.Time `json:"next_check"`

	CurrentVersion string `json:"current_version"`

	LatestVersion string `json:"latest_version,omitempty"`

	UpdateAvailable bool `json:"update_available"`

	LastResult string `json:"last_result"`

	LastError string `json:"last_error,omitempty"`

	LastInstallAttempt *time.Time `json:"last_install_attempt,omitempty"`

	LastInstallFromVersion string `json:"last_install_from_version,omitempty"`

	LastInstallTarget string `json:"last_install_target,omitempty"`

	LastInstallResult string `json:"last_install_result,omitempty"`

	LastInstallError string `json:"last_install_error,omitempty"`

	LastInstalledVersion string `json:"last_installed_version,omitempty"`

	LastRollback bool `json:"last_rollback,omitempty"`

	LastRollbackVerified bool `json:"last_rollback_verified,omitempty"`

	RecoveredVersion string `json:"recovered_version,omitempty"`
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

	if err := s.validatePath(); err != nil {
		return UpdateState{}, err
	}

	return s.loadUnlocked()
}

// Save atomically persists the complete update state.
func (s *StateStore) Save(
	state UpdateState,
) error {

	if err := s.validatePath(); err != nil {
		return err
	}

	if err := validateUpdateState(
		state,
	); err != nil {

		return err
	}

	return s.withExclusiveLock(
		func() error {

			return s.saveUnlocked(
				state,
			)
		},
	)
}

// SaveCheck persists update-check information while preserving
// the latest installation lifecycle state.
func (s *StateStore) SaveCheck(
	state UpdateState,
) error {

	if err := s.validatePath(); err != nil {
		return err
	}

	if err := validateCheckState(
		state,
	); err != nil {

		return err
	}

	return s.withExclusiveLock(
		func() error {

			existing, err := s.loadUnlocked()

			if err != nil {

				if !os.IsNotExist(
					err,
				) {

					return err
				}

			} else {

				copyInstallLifecycle(
					&state,
					existing,
				)
			}

			return s.saveUnlocked(
				state,
			)
		},
	)
}

// RecordInstallStarted records the beginning of a manual
// Sentinel update installation.
func (s *StateStore) RecordInstallStarted(
	at time.Time,
	fromVersion string,
	targetVersion string,
) error {

	if err := s.validatePath(); err != nil {
		return err
	}

	if at.IsZero() {

		return fmt.Errorf(
			"install attempt time is empty",
		)
	}

	if !IsValidVersion(
		fromVersion,
	) {

		return fmt.Errorf(
			"invalid install source version: %s",
			fromVersion,
		)
	}

	if !IsValidVersion(
		targetVersion,
	) {

		return fmt.Errorf(
			"invalid install target version: %s",
			targetVersion,
		)
	}

	if !IsNewerVersion(
		fromVersion,
		targetVersion,
	) {

		return fmt.Errorf(
			"install target version %s is not newer than %s",
			NormalizeVersion(
				targetVersion,
			),
			NormalizeVersion(
				fromVersion,
			),
		)
	}

	return s.updateExisting(
		func(state *UpdateState) error {

			attempt := at.UTC()

			state.LastInstallAttempt = &attempt

			state.LastInstallFromVersion =
				NormalizeVersion(
					fromVersion,
				)

			state.LastInstallTarget =
				NormalizeVersion(
					targetVersion,
				)

			state.LastInstallResult =
				InstallResultInProgress

			state.LastInstallError = ""

			state.LastInstalledVersion = ""

			state.LastRollback = false

			state.LastRollbackVerified = false

			state.RecoveredVersion = ""

			return nil
		},
	)
}

// RecordInstallSuccess records a successfully activated release.
func (s *StateStore) RecordInstallSuccess() error {

	if err := s.validatePath(); err != nil {
		return err
	}

	return s.updateExisting(
		func(state *UpdateState) error {

			if state.LastInstallResult !=
				InstallResultInProgress {

				return fmt.Errorf(
					"no update installation is in progress",
				)
			}

			state.LastInstallResult =
				InstallResultSuccess

			state.LastInstallError = ""

			state.LastInstalledVersion =
				state.LastInstallTarget

			state.LastRollback = false

			state.LastRollbackVerified = false

			state.RecoveredVersion = ""

			return nil
		},
	)
}

// RecordInstallFailure records a failed installation.
//
// If rollbackVerified is true, the recovered version is recorded
// as the version from which the installation started.
func (s *StateStore) RecordInstallFailure(
	installError string,
	rolledBack bool,
	rollbackVerified bool,
) error {

	if err := s.validatePath(); err != nil {
		return err
	}

	installError = strings.TrimSpace(
		installError,
	)

	if installError == "" {

		return fmt.Errorf(
			"install failure error is empty",
		)
	}

	if rollbackVerified &&
		!rolledBack {

		return fmt.Errorf(
			"rollback cannot be verified when rollback was not attempted",
		)
	}

	return s.updateExisting(
		func(state *UpdateState) error {

			if state.LastInstallResult !=
				InstallResultInProgress {

				return fmt.Errorf(
					"no update installation is in progress",
				)
			}

			state.LastInstallResult =
				InstallResultFailed

			state.LastInstallError =
				installError

			state.LastInstalledVersion = ""

			state.LastRollback =
				rolledBack

			state.LastRollbackVerified =
				rollbackVerified

			state.RecoveredVersion = ""

			if rollbackVerified {

				state.RecoveredVersion =
					state.LastInstallFromVersion
			}

			return nil
		},
	)
}

func (s *StateStore) updateExisting(
	update func(
		*UpdateState,
	) error,
) error {

	return s.withExclusiveLock(
		func() error {

			state, err := s.loadUnlocked()

			if err != nil {
				return err
			}

			if err := update(
				&state,
			); err != nil {

				return err
			}

			return s.saveUnlocked(
				state,
			)
		},
	)
}

func (s *StateStore) loadUnlocked() (
	UpdateState,
	error,
) {

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

func (s *StateStore) saveUnlocked(
	state UpdateState,
) error {

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

func (s *StateStore) withExclusiveLock(
	fn func() error,
) error {

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

	lockFile, err := os.OpenFile(
		s.path+".lock",
		os.O_CREATE|os.O_RDWR,
		0644,
	)

	if err != nil {

		return fmt.Errorf(
			"open update state lock: %w",
			err,
		)
	}

	defer lockFile.Close()

	if err := syscall.Flock(
		int(
			lockFile.Fd(),
		),
		syscall.LOCK_EX,
	); err != nil {

		return fmt.Errorf(
			"lock update state: %w",
			err,
		)
	}

	defer syscall.Flock(
		int(
			lockFile.Fd(),
		),
		syscall.LOCK_UN,
	)

	return fn()
}

func (s *StateStore) validatePath() error {

	if s == nil ||
		s.path == "" {

		return fmt.Errorf(
			"update state path is empty",
		)
	}

	return nil
}

func validateUpdateState(
	state UpdateState,
) error {

	if err := validateCheckState(
		state,
	); err != nil {

		return err
	}

	return validateInstallState(
		state,
	)
}

func validateCheckState(
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

func validateInstallState(
	state UpdateState,
) error {

	if state.LastInstallAttempt == nil {

		if state.LastInstallFromVersion != "" ||
			state.LastInstallTarget != "" ||
			state.LastInstallResult != "" ||
			state.LastInstallError != "" ||
			state.LastInstalledVersion != "" ||
			state.LastRollback ||
			state.LastRollbackVerified ||
			state.RecoveredVersion != "" {

			return fmt.Errorf(
				"install lifecycle exists without install attempt",
			)
		}

		return nil
	}

	if state.LastInstallAttempt.IsZero() {

		return fmt.Errorf(
			"install attempt time is empty",
		)
	}

	if !IsValidVersion(
		state.LastInstallFromVersion,
	) {

		return fmt.Errorf(
			"invalid install source version: %s",
			state.LastInstallFromVersion,
		)
	}

	if !IsValidVersion(
		state.LastInstallTarget,
	) {

		return fmt.Errorf(
			"invalid install target version: %s",
			state.LastInstallTarget,
		)
	}

	switch state.LastInstallResult {

	case InstallResultInProgress:

		if state.LastInstallError != "" ||
			state.LastInstalledVersion != "" ||
			state.LastRollback ||
			state.LastRollbackVerified ||
			state.RecoveredVersion != "" {

			return fmt.Errorf(
				"in-progress installation contains completed lifecycle state",
			)
		}

	case InstallResultSuccess:

		if state.LastInstallError != "" {

			return fmt.Errorf(
				"successful installation contains an error",
			)
		}

		if !IsValidVersion(
			state.LastInstalledVersion,
		) {

			return fmt.Errorf(
				"successful installation has invalid installed version: %s",
				state.LastInstalledVersion,
			)
		}

		if NormalizeVersion(
			state.LastInstalledVersion,
		) != NormalizeVersion(
			state.LastInstallTarget,
		) {

			return fmt.Errorf(
				"installed version does not match install target",
			)
		}

		if state.LastRollback ||
			state.LastRollbackVerified ||
			state.RecoveredVersion != "" {

			return fmt.Errorf(
				"successful installation contains rollback state",
			)
		}

	case InstallResultFailed:

		if strings.TrimSpace(
			state.LastInstallError,
		) == "" {

			return fmt.Errorf(
				"failed installation has no error",
			)
		}

		if state.LastInstalledVersion != "" {

			return fmt.Errorf(
				"failed installation contains installed version",
			)
		}

		if state.LastRollbackVerified &&
			!state.LastRollback {

			return fmt.Errorf(
				"verified rollback without rollback attempt",
			)
		}

		if state.LastRollbackVerified {

			if !IsValidVersion(
				state.RecoveredVersion,
			) {

				return fmt.Errorf(
					"verified rollback has invalid recovered version: %s",
					state.RecoveredVersion,
				)
			}

			if NormalizeVersion(
				state.RecoveredVersion,
			) != NormalizeVersion(
				state.LastInstallFromVersion,
			) {

				return fmt.Errorf(
					"recovered version does not match install source version",
				)
			}

		} else if state.RecoveredVersion != "" {

			return fmt.Errorf(
				"unverified rollback contains recovered version",
			)
		}

	default:

		return fmt.Errorf(
			"invalid install result: %s",
			state.LastInstallResult,
		)
	}

	return nil
}

func copyInstallLifecycle(
	destination *UpdateState,
	source UpdateState,
) {

	destination.LastInstallAttempt =
		source.LastInstallAttempt

	destination.LastInstallFromVersion =
		source.LastInstallFromVersion

	destination.LastInstallTarget =
		source.LastInstallTarget

	destination.LastInstallResult =
		source.LastInstallResult

	destination.LastInstallError =
		source.LastInstallError

	destination.LastInstalledVersion =
		source.LastInstalledVersion

	destination.LastRollback =
		source.LastRollback

	destination.LastRollbackVerified =
		source.LastRollbackVerified

	destination.RecoveredVersion =
		source.RecoveredVersion
}
