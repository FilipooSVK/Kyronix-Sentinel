package updater

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const updateOperationLockSuffix = ".operation.lock"

var ErrUpdateOperationLocked = errors.New("update operation already in progress")

// OperationLock serializes complete Sentinel update transactions
// across processes.
//
// Unlike the StateStore lock, this lock is intended to remain held
// for the entire update operation: check, download, validation,
// installation, activation and post-install health verification.
type OperationLock struct {
	path string

	file *os.File
}

// OperationLockPath derives the global operation lock path from
// the persistent update state path.
func OperationLockPath(
	statePath string,
) string {

	statePath = strings.TrimSpace(
		statePath,
	)

	if statePath == "" {
		return ""
	}

	return statePath +
		updateOperationLockSuffix
}

// NewOperationLock creates a global Sentinel update operation lock.
func NewOperationLock(
	path string,
) *OperationLock {

	return &OperationLock{
		path: strings.TrimSpace(
			path,
		),
	}
}

// Path returns the configured lock file path.
func (l *OperationLock) Path() string {

	if l == nil {
		return ""
	}

	return l.path
}

// Acquire attempts to acquire the operation lock.
//
// Acquisition is deliberately non-blocking. If another updater
// already owns the lock, ErrUpdateOperationLocked is returned
// immediately.
func (l *OperationLock) Acquire() error {

	if l == nil ||
		l.path == "" {

		return fmt.Errorf(
			"update operation lock path is empty",
		)
	}

	if l.file != nil {

		return fmt.Errorf(
			"update operation lock is already acquired by this instance",
		)
	}

	dir := filepath.Dir(
		l.path,
	)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {

		return fmt.Errorf(
			"create update operation lock directory: %w",
			err,
		)
	}

	lockFile, err := os.OpenFile(
		l.path,
		os.O_CREATE|os.O_RDWR,
		0644,
	)

	if err != nil {

		return fmt.Errorf(
			"open update operation lock: %w",
			err,
		)
	}

	err = syscall.Flock(
		int(
			lockFile.Fd(),
		),
		syscall.LOCK_EX|
			syscall.LOCK_NB,
	)

	if err != nil {

		lockFile.Close()

		if errors.Is(
			err,
			syscall.EWOULDBLOCK,
		) ||
			errors.Is(
				err,
				syscall.EAGAIN,
			) {

			return fmt.Errorf(
				"%w: %s",
				ErrUpdateOperationLocked,
				l.path,
			)
		}

		return fmt.Errorf(
			"lock update operation: %w",
			err,
		)
	}

	l.file = lockFile

	return nil
}

// Release releases the operation lock.
//
// Release is idempotent so deferred cleanup is safe.
func (l *OperationLock) Release() error {

	if l == nil ||
		l.file == nil {

		return nil
	}

	lockFile := l.file
	l.file = nil

	unlockErr := syscall.Flock(
		int(
			lockFile.Fd(),
		),
		syscall.LOCK_UN,
	)

	closeErr := lockFile.Close()

	if unlockErr != nil {

		return fmt.Errorf(
			"unlock update operation: %w",
			unlockErr,
		)
	}

	if closeErr != nil {

		return fmt.Errorf(
			"close update operation lock: %w",
			closeErr,
		)
	}

	return nil
}
