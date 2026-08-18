package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/updater"
	"kyronix/sentinel/internal/version"
)

func runUpdatePolicy() {

	cfg, err := config.Load(
		sentinelSystemConfig,
	)

	if err != nil {

		fmt.Println(
			"Unable to load Sentinel configuration:",
			err,
		)

		return
	}

	fmt.Println(
		"Kyronix Sentinel Automatic Install Policy",
	)

	fmt.Println()

	fmt.Println(
		"Automatic install:",
		enabledDisabled(
			cfg.Update.AutoInstall,
		),
	)

	fmt.Println(
		"Minimum release age:",
		cfg.Update.AutoInstallPolicy.MinReleaseAge,
	)

	fmt.Println(
		"Patch only:",
		yesNo(
			cfg.Update.AutoInstallPolicy.PatchOnly,
		),
	)

	if !cfg.Update.Enabled {

		fmt.Println()
		fmt.Println(
			"Decision: DENY",
		)

		fmt.Println(
			"Reason: update system disabled",
		)

		return
	}

	if !updateRepositoryConfigured(
		cfg,
	) {

		fmt.Println()
		fmt.Println(
			"Decision: DENY",
		)

		fmt.Println(
			"Reason: update repository not configured",
		)

		return
	}

	client := updater.NewGitHubClient(
		cfg.Update.Owner,
		cfg.Update.Repository,
	)

	check, err := client.Check(
		context.Background(),
		version.Version,
	)

	if err != nil {

		fmt.Println()
		fmt.Println(
			"Decision: UNKNOWN",
		)

		fmt.Println(
			"Reason: update check failed",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println()

	fmt.Println(
		"Current:",
		check.CurrentVersion,
	)

	fmt.Println(
		"Latest:",
		check.LatestVersion,
	)

	if !check.Release.PublishedAt.IsZero() {

		fmt.Println(
			"Published:",
			check.Release.PublishedAt.UTC().Format(
				"2006-01-02 15:04:05 UTC",
			),
		)
	}

	store := updater.NewStateStore(
		cfg.Update.StatePath,
	)

	state, err := store.Load()

	if err != nil {

		if !os.IsNotExist(
			err,
		) {

			fmt.Println()
			fmt.Println(
				"Decision: UNKNOWN",
			)

			fmt.Println(
				"Reason: update state could not be read",
			)

			fmt.Println(
				"Error:",
				err,
			)

			return
		}

		state = updater.UpdateState{}
	}

	decision, err :=
		updater.EvaluateAutoInstallPolicy(
			updater.AutoInstallPolicy{
				Enabled: cfg.Update.AutoInstall,

				MinReleaseAge: cfg.Update.AutoInstallPolicy.MinReleaseAge,

				PatchOnly: cfg.Update.AutoInstallPolicy.PatchOnly,
			},
			updater.AutoInstallPolicyInput{
				Check: check,

				State: state,

				Now: time.Now().UTC(),
			},
		)

	if err != nil {

		fmt.Println()
		fmt.Println(
			"Decision: ERROR",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	if decision.ReleaseAge > 0 {

		fmt.Println(
			"Release age:",
			formatPolicyDuration(
				decision.ReleaseAge,
			),
		)
	}

	fmt.Println()

	if decision.Allowed {

		fmt.Println(
			"Decision: ALLOW",
		)

		fmt.Println(
			"Automatic installation is permitted by policy.",
		)

		return
	}

	fmt.Println(
		"Decision: DENY",
	)

	fmt.Println(
		"Reasons:",
	)

	for _, reason := range decision.Reasons {

		fmt.Printf(
			"  - %s: %s\n",
			reason,
			autoInstallReasonDescription(
				reason,
			),
		)
	}
}

func autoInstallReasonDescription(
	reason string,
) string {

	switch reason {

	case updater.AutoInstallReasonDisabled:

		return "automatic installation is disabled"

	case updater.AutoInstallReasonNoUpdate:

		return "no newer release is available"

	case updater.AutoInstallReasonInvalidCurrentVersion:

		return "current Sentinel version is invalid"

	case updater.AutoInstallReasonInvalidTargetVersion:

		return "target release version is invalid"

	case updater.AutoInstallReasonTargetNotNewer:

		return "target release is not newer than the installed version"

	case updater.AutoInstallReasonReleaseTagMismatch:

		return "GitHub release tag does not match the target version"

	case updater.AutoInstallReasonDraftRelease:

		return "draft releases cannot be installed automatically"

	case updater.AutoInstallReasonPrerelease:

		return "prereleases cannot be installed automatically"

	case updater.AutoInstallReasonReleaseTimeMissing:

		return "release publication time is missing"

	case updater.AutoInstallReasonReleaseTimeFuture:

		return "release publication time is in the future"

	case updater.AutoInstallReasonReleaseTooYoung:

		return "release has not reached the minimum required age"

	case updater.AutoInstallReasonNonPatchUpgrade:

		return "policy permits patch upgrades only"

	case updater.AutoInstallReasonInstallInProgress:

		return "another update installation is already in progress"

	case updater.AutoInstallReasonUnverifiedRollback:

		return "previous rollback was not verified"

	case updater.AutoInstallReasonQuarantined:

		return "release is quarantined after a previous failed activation"

	default:

		return "policy condition was not satisfied"
	}
}

func enabledDisabled(
	value bool,
) string {

	if value {
		return "enabled"
	}

	return "disabled"
}

func formatPolicyDuration(
	value time.Duration,
) string {

	if value < 0 {
		return value.String()
	}

	hours :=
		int64(
			value / time.Hour,
		)

	minutes :=
		int64(
			(value % time.Hour) /
				time.Minute,
		)

	if hours >= 24 {

		days := hours / 24

		remainingHours :=
			hours % 24

		if remainingHours == 0 &&
			minutes == 0 {

			return fmt.Sprintf(
				"%dd",
				days,
			)
		}

		return fmt.Sprintf(
			"%dd %dh %dm",
			days,
			remainingHours,
			minutes,
		)
	}

	return value.Round(
		time.Minute,
	).String()
}
