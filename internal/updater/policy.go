package updater

import (
	"fmt"
	"time"
)

const (
	AutoInstallReasonDisabled = "auto_install_disabled"

	AutoInstallReasonNoUpdate = "no_update_available"

	AutoInstallReasonInvalidCurrentVersion = "invalid_current_version"

	AutoInstallReasonInvalidTargetVersion = "invalid_target_version"

	AutoInstallReasonTargetNotNewer = "target_not_newer"

	AutoInstallReasonReleaseTagMismatch = "release_tag_mismatch"

	AutoInstallReasonDraftRelease = "draft_release"

	AutoInstallReasonPrerelease = "prerelease_release"

	AutoInstallReasonReleaseTimeMissing = "release_time_missing"

	AutoInstallReasonReleaseTimeFuture = "release_time_in_future"

	AutoInstallReasonReleaseTooYoung = "release_too_young"

	AutoInstallReasonNonPatchUpgrade = "non_patch_upgrade"

	AutoInstallReasonInstallInProgress = "install_in_progress"

	AutoInstallReasonUnverifiedRollback = "unverified_rollback"

	AutoInstallReasonQuarantined = "release_quarantined"
)

// AutoInstallPolicy defines Sentinel unattended update rules.
type AutoInstallPolicy struct {
	Enabled bool

	MinReleaseAge time.Duration

	PatchOnly bool
}

// AutoInstallPolicyInput contains all state required to evaluate
// one automatic installation decision.
type AutoInstallPolicyInput struct {
	Check CheckResult

	State UpdateState

	Now time.Time
}

// AutoInstallDecision explains whether unattended installation
// is currently permitted.
type AutoInstallDecision struct {
	Allowed bool

	Reasons []string

	ReleaseAge time.Duration
}

// EvaluateAutoInstallPolicy evaluates whether Sentinel may
// automatically install a checked release.
//
// This function has no side effects. It does not download,
// install, restart, modify state, or clear quarantine.
func EvaluateAutoInstallPolicy(
	policy AutoInstallPolicy,
	input AutoInstallPolicyInput,
) (
	AutoInstallDecision,
	error,
) {

	if policy.MinReleaseAge < 0 {

		return AutoInstallDecision{}, fmt.Errorf(
			"minimum release age cannot be negative",
		)
	}

	if input.Now.IsZero() {

		return AutoInstallDecision{}, fmt.Errorf(
			"policy evaluation time is empty",
		)
	}

	decision := AutoInstallDecision{}

	addReason := func(
		reason string,
	) {

		for _, existing := range decision.Reasons {

			if existing == reason {
				return
			}
		}

		decision.Reasons =
			append(
				decision.Reasons,
				reason,
			)
	}

	if !policy.Enabled {

		addReason(
			AutoInstallReasonDisabled,
		)
	}

	if !input.Check.UpdateAvailable {

		addReason(
			AutoInstallReasonNoUpdate,
		)
	}

	currentVersion, currentOK :=
		parseSemanticVersion(
			input.Check.CurrentVersion,
		)

	if !currentOK {

		addReason(
			AutoInstallReasonInvalidCurrentVersion,
		)
	}

	targetVersion, targetOK :=
		parseSemanticVersion(
			input.Check.LatestVersion,
		)

	if !targetOK {

		addReason(
			AutoInstallReasonInvalidTargetVersion,
		)
	}

	if currentOK &&
		targetOK &&
		CompareVersions(
			input.Check.LatestVersion,
			input.Check.CurrentVersion,
		) <= 0 {

		addReason(
			AutoInstallReasonTargetNotNewer,
		)
	}

	if targetOK {

		if !IsValidVersion(
			input.Check.Release.TagName,
		) ||
			NormalizeVersion(
				input.Check.Release.TagName,
			) !=
				NormalizeVersion(
					input.Check.LatestVersion,
				) {

			addReason(
				AutoInstallReasonReleaseTagMismatch,
			)
		}
	}

	if input.Check.Release.Draft {

		addReason(
			AutoInstallReasonDraftRelease,
		)
	}

	if input.Check.Release.Prerelease {

		addReason(
			AutoInstallReasonPrerelease,
		)
	}

	if targetOK &&
		len(
			targetVersion.prerelease,
		) > 0 {

		addReason(
			AutoInstallReasonPrerelease,
		)
	}

	publishedAt :=
		input.Check.Release.PublishedAt

	switch {

	case publishedAt.IsZero():

		addReason(
			AutoInstallReasonReleaseTimeMissing,
		)

	case publishedAt.After(
		input.Now,
	):

		addReason(
			AutoInstallReasonReleaseTimeFuture,
		)

	default:

		decision.ReleaseAge =
			input.Now.Sub(
				publishedAt,
			)

		if decision.ReleaseAge <
			policy.MinReleaseAge {

			addReason(
				AutoInstallReasonReleaseTooYoung,
			)
		}
	}

	if policy.PatchOnly &&
		currentOK &&
		targetOK {

		if currentVersion.major !=
			targetVersion.major ||
			currentVersion.minor !=
				targetVersion.minor {

			addReason(
				AutoInstallReasonNonPatchUpgrade,
			)
		}
	}

	if input.State.LastInstallResult ==
		InstallResultInProgress {

		addReason(
			AutoInstallReasonInstallInProgress,
		)
	}

	if input.State.LastRollback &&
		!input.State.LastRollbackVerified {

		addReason(
			AutoInstallReasonUnverifiedRollback,
		)
	}

	if updateStateQuarantinesVersion(
		input.State,
		input.Check.LatestVersion,
	) {

		addReason(
			AutoInstallReasonQuarantined,
		)
	}

	decision.Allowed =
		len(
			decision.Reasons,
		) == 0

	return decision, nil
}

func updateStateQuarantinesVersion(
	state UpdateState,
	version string,
) bool {

	if !IsValidVersion(
		version,
	) {

		return false
	}

	target :=
		NormalizeVersion(
			version,
		)

	for _, blocked := range state.QuarantinedVersions {

		if NormalizeVersion(
			blocked,
		) == target {

			return true
		}
	}

	// Compatibility fallback for update state written by
	// early quarantine implementations.
	if state.QuarantinedVersion != "" &&
		NormalizeVersion(
			state.QuarantinedVersion,
		) == target {

		return true
	}

	return false
}
