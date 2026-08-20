package updater

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var ErrReleaseQuarantined = errors.New("release is quarantined")

type ExecutorStage string

const (
	ExecutorStageIdle         ExecutorStage = "idle"
	ExecutorStageLock         ExecutorStage = "lock"
	ExecutorStageCheck        ExecutorStage = "check"
	ExecutorStagePersistCheck ExecutorStage = "persist_check"
	ExecutorStageQuarantine   ExecutorStage = "quarantine"
	ExecutorStageLifecycle    ExecutorStage = "lifecycle"
	ExecutorStageAssets       ExecutorStage = "assets"
	ExecutorStageWorkDir      ExecutorStage = "work_dir"
	ExecutorStageDownload     ExecutorStage = "download"
	ExecutorStageValidation   ExecutorStage = "validation"
	ExecutorStageActivation   ExecutorStage = "activation"
	ExecutorStageComplete     ExecutorStage = "complete"
)

type ExecutorEventType string

const (
	ExecutorEventChecking        ExecutorEventType = "checking"
	ExecutorEventChecked         ExecutorEventType = "checked"
	ExecutorEventUpdateAvailable ExecutorEventType = "update_available"
	ExecutorEventAssetsSelected  ExecutorEventType = "assets_selected"
	ExecutorEventDownloading     ExecutorEventType = "downloading"
	ExecutorEventDownloaded      ExecutorEventType = "downloaded"
	ExecutorEventValidating      ExecutorEventType = "validating"
	ExecutorEventValidated       ExecutorEventType = "validated"
	ExecutorEventInstalling      ExecutorEventType = "installing"
	ExecutorEventCompleted       ExecutorEventType = "completed"
)

type ExecutorEvent struct {
	Type ExecutorEventType

	Check CheckResult

	Assets ReleaseAssets

	Artifact VerifiedArtifact

	Activation ActivationResult
}

type ExecutorObserver func(
	event ExecutorEvent,
)

type InstallExecutorConfig struct {
	CurrentVersion string

	CheckInterval time.Duration

	StatePath string

	InstallDir string

	WorkerUnitTarget string
}

type InstallExecutorResult struct {
	Stage ExecutorStage

	Check CheckResult

	Assets ReleaseAssets

	Artifact VerifiedArtifact

	Activation ActivationResult

	UpToDate bool

	QuarantineFailures uint64

	StateRecordError error

	CleanupError error
}

type InstallExecutor struct {
	checker ReleaseChecker

	config InstallExecutorConfig

	service ServiceController

	health VersionHealthChecker

	reloader SystemdReloader

	selectAssets func(
		Release,
	) (ReleaseAssets, error)

	download func(
		context.Context,
		ReleaseAssets,
		string,
	) (VerifiedArtifact, error)

	extract func(
		VerifiedArtifact,
		string,
		string,
		string,
		string,
	) (ExtractedRelease, error)

	activate func(
		context.Context,
		ExtractedRelease,
		string,
		string,
		string,
		ServiceController,
		VersionHealthChecker,
		...ActivationOptions,
	) (ActivationResult, error)

	now func() time.Time

	mkdirTemp func(
		string,
		string,
	) (string, error)
}

func NewInstallExecutor(
	checker ReleaseChecker,
	config InstallExecutorConfig,
	service ServiceController,
	reloader SystemdReloader,
	health VersionHealthChecker,
) *InstallExecutor {

	return &InstallExecutor{
		checker: checker,

		config: config,

		service: service,

		reloader: reloader,

		health: health,

		selectAssets: SelectCurrentPlatformAssets,

		download: DownloadAndVerify,

		extract: ExtractAndValidateRelease,

		activate: ActivateRelease,

		now: func() time.Time {
			return time.Now().UTC()
		},

		mkdirTemp: os.MkdirTemp,
	}
}

func (e *InstallExecutor) Execute(
	ctx context.Context,
	observer ExecutorObserver,
) (
	result InstallExecutorResult,
	err error,
) {

	result.Stage = ExecutorStageIdle

	if err := e.validate(); err != nil {
		return result, err
	}

	stateStore := NewStateStore(
		e.config.StatePath,
	)

	result.Stage =
		ExecutorStageLock

	operationLock := NewOperationLock(
		OperationLockPath(
			e.config.StatePath,
		),
	)

	if err := operationLock.Acquire(); err != nil {
		return result, err
	}

	defer func() {

		if releaseErr :=
			operationLock.Release(); releaseErr != nil {

			result.CleanupError =
				releaseErr
		}
	}()

	result.Stage =
		ExecutorStageCheck

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventChecking,
		},
	)

	check, err := e.checker.Check(
		ctx,
		e.config.CurrentVersion,
	)

	if err != nil {
		return result, err
	}

	result.Check = check

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type:  ExecutorEventChecked,
			Check: check,
		},
	)

	result.Stage =
		ExecutorStagePersistCheck

	checkedAt := e.now().UTC()

	checkState := UpdateState{
		LastCheck: checkedAt,

		NextCheck: checkedAt.Add(
			e.config.CheckInterval,
		),

		CurrentVersion: check.CurrentVersion,

		LatestVersion: check.LatestVersion,

		UpdateAvailable: check.UpdateAvailable,

		LastResult: UpdateResultSuccess,
	}

	if err := stateStore.SaveCheck(
		checkState,
	); err != nil {

		return result, err
	}

	if !check.UpdateAvailable {

		result.UpToDate = true

		result.Stage =
			ExecutorStageComplete

		return result, nil
	}

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventUpdateAvailable,

			Check: check,
		},
	)

	result.Stage =
		ExecutorStageQuarantine

	quarantined, err :=
		stateStore.IsVersionQuarantined(
			check.LatestVersion,
		)

	if err != nil {
		return result, err
	}

	if quarantined {

		state, loadErr :=
			stateStore.Load()

		if loadErr != nil {
			return result, loadErr
		}

		result.QuarantineFailures =
			state.QuarantineFailureCount

		return result, fmt.Errorf(
			"%w: %s",
			ErrReleaseQuarantined,
			check.LatestVersion,
		)
	}

	result.Stage =
		ExecutorStageLifecycle

	if err := stateStore.RecordInstallStarted(
		e.now().UTC(),
		e.config.CurrentVersion,
		check.LatestVersion,
	); err != nil {

		return result, err
	}

	recordFailure := func(
		installErr error,
		rolledBack bool,
		rollbackVerified bool,
	) {

		if installErr == nil {
			return
		}

		recordErr :=
			stateStore.RecordInstallFailure(
				installErr.Error(),
				rolledBack,
				rollbackVerified,
			)

		if recordErr != nil {

			result.StateRecordError =
				recordErr
		}
	}

	result.Stage =
		ExecutorStageAssets

	assets, err :=
		e.selectAssets(
			check.Release,
		)

	if err != nil {

		recordFailure(
			err,
			false,
			false,
		)

		return result, err
	}

	result.Assets = assets

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventAssetsSelected,

			Check: check,

			Assets: assets,
		},
	)

	result.Stage =
		ExecutorStageWorkDir

	workDir, err :=
		e.mkdirTemp(
			"",
			"sentinel-update-*",
		)

	if err != nil {

		recordFailure(
			err,
			false,
			false,
		)

		return result, err
	}

	defer os.RemoveAll(
		workDir,
	)

	result.Stage =
		ExecutorStageDownload

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventDownloading,

			Check: check,

			Assets: assets,
		},
	)

	artifact, err :=
		e.download(
			ctx,
			assets,
			filepath.Join(
				workDir,
				"download",
			),
		)

	if err != nil {

		recordFailure(
			err,
			false,
			false,
		)

		return result, err
	}

	result.Artifact = artifact

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventDownloaded,

			Check: check,

			Assets: assets,

			Artifact: artifact,
		},
	)

	result.Stage =
		ExecutorStageValidation

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventValidating,

			Check: check,

			Assets: assets,

			Artifact: artifact,
		},
	)

	release, err :=
		e.extract(
			artifact,
			filepath.Join(
				workDir,
				"extract",
			),
			check.LatestVersion,
			assets.OS,
			assets.Arch,
		)

	if err != nil {

		recordFailure(
			err,
			false,
			false,
		)

		return result, err
	}

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventValidated,

			Check: check,

			Assets: assets,

			Artifact: artifact,
		},
	)

	result.Stage =
		ExecutorStageActivation

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventInstalling,

			Check: check,

			Assets: assets,

			Artifact: artifact,
		},
	)

	activation, err :=
		e.activate(
			ctx,
			release,
			e.config.InstallDir,
			check.LatestVersion,
			e.config.CurrentVersion,
			e.service,
			e.health,
			ActivationOptions{
				WorkerUnitTarget: e.config.WorkerUnitTarget,

				Reloader: e.reloader,
			},
		)

	result.Activation =
		activation

	if err != nil {

		recordFailure(
			err,
			activation.RolledBack,
			activation.RollbackVerified,
		)

		return result, err
	}

	if recordErr :=
		stateStore.RecordInstallSuccess(); recordErr != nil {

		result.StateRecordError =
			recordErr
	}

	result.Stage =
		ExecutorStageComplete

	emitExecutorEvent(
		observer,
		ExecutorEvent{
			Type: ExecutorEventCompleted,

			Check: check,

			Assets: assets,

			Artifact: artifact,

			Activation: activation,
		},
	)

	return result, nil
}

func (e *InstallExecutor) validate() error {

	if e == nil {
		return fmt.Errorf(
			"install executor is nil",
		)
	}

	if e.checker == nil {
		return fmt.Errorf(
			"update checker is nil",
		)
	}

	if e.service == nil {
		return fmt.Errorf(
			"service controller is nil",
		)
	}

	if e.health == nil {
		return fmt.Errorf(
			"health checker is nil",
		)
	}

	if e.reloader == nil {
		return fmt.Errorf(
			"systemd reloader is nil",
		)
	}

	if !IsValidVersion(
		e.config.CurrentVersion,
	) {

		return fmt.Errorf(
			"invalid current version: %s",
			e.config.CurrentVersion,
		)
	}

	if e.config.CheckInterval <= 0 {

		return fmt.Errorf(
			"update check interval must be greater than zero",
		)
	}

	if strings.TrimSpace(
		e.config.StatePath,
	) == "" {

		return fmt.Errorf(
			"update state path is empty",
		)
	}

	if strings.TrimSpace(
		e.config.InstallDir,
	) == "" {

		return fmt.Errorf(
			"install directory is empty",
		)
	}

	if strings.TrimSpace(
		e.config.WorkerUnitTarget,
	) == "" {

		return fmt.Errorf(
			"worker unit target path is empty",
		)
	}

	if e.selectAssets == nil ||
		e.download == nil ||
		e.extract == nil ||
		e.activate == nil ||
		e.now == nil ||
		e.mkdirTemp == nil {

		return fmt.Errorf(
			"install executor dependencies are incomplete",
		)
	}

	return nil
}

func emitExecutorEvent(
	observer ExecutorObserver,
	event ExecutorEvent,
) {

	if observer == nil {
		return
	}

	observer(
		event,
	)
}
