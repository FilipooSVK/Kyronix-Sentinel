package main

import (
	"fmt"
	"os"

	"kyronix/sentinel/internal/config"
	"kyronix/sentinel/internal/updater"
)

func runUpdateQuarantine(
	args []string,
) {

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

	store := updater.NewStateStore(
		cfg.Update.StatePath,
	)

	if len(args) == 0 {

		state, err := store.Load()

		if err != nil {

			if os.IsNotExist(err) {

				fmt.Println(
					"Kyronix Sentinel Update Quarantine",
				)

				fmt.Println()
				fmt.Println(
					"State: CLEAR",
				)

				return
			}

			fmt.Println(
				"Unable to load update state:",
				err,
			)

			return
		}

		printUpdateQuarantine(
			state,
		)

		return
	}

	if len(args) != 1 ||
		args[0] != "clear" {

		fmt.Println(
			"usage: sentinelctl update quarantine [clear]",
		)

		return
	}

	if os.Geteuid() != 0 {

		fmt.Println(
			"Status: ROOT PRIVILEGES REQUIRED",
		)

		fmt.Println()
		fmt.Println(
			"Run:",
			"sudo sentinelctl update quarantine clear",
		)

		return
	}

	state, err := store.Load()

	if err != nil {

		fmt.Println(
			"Status: QUARANTINE READ FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	if state.QuarantinedVersion == "" {

		fmt.Println(
			"Status: QUARANTINE ALREADY CLEAR",
		)

		return
	}

	blocked :=
		len(
			state.QuarantinedVersions,
		)

	if err := store.ClearQuarantine(); err != nil {

		fmt.Println(
			"Status: QUARANTINE CLEAR FAILED",
		)

		fmt.Println(
			"Error:",
			err,
		)

		return
	}

	fmt.Println(
		"Status: QUARANTINE CLEARED",
	)

	fmt.Println(
		"Released versions:",
		blocked,
	)
}

func printUpdateQuarantine(
	state updater.UpdateState,
) {

	fmt.Println(
		"Kyronix Sentinel Update Quarantine",
	)

	fmt.Println()

	if state.QuarantinedVersion == "" {

		fmt.Println(
			"State: CLEAR",
		)

		return
	}

	fmt.Println(
		"State: ACTIVE",
	)

	fmt.Println(
		"Blocked releases:",
		len(
			state.QuarantinedVersions,
		),
	)

	fmt.Println(
		"Latest failed:",
		state.QuarantinedVersion,
	)

	fmt.Println(
		"Quarantined:",
		state.QuarantinedAt.UTC().Format(
			"2006-01-02 15:04:05 UTC",
		),
	)

	fmt.Println(
		"Failures:",
		state.QuarantineFailureCount,
	)

	fmt.Println(
		"Reason:",
		state.QuarantineReason,
	)

	fmt.Println(
		"Error:",
		state.QuarantineLastError,
	)
}

func printUpdateQuarantineSummary(
	state updater.UpdateState,
) {

	fmt.Println()
	fmt.Println(
		"Quarantine:",
	)

	if state.QuarantinedVersion == "" {

		fmt.Println(
			"  State: clear",
		)

		return
	}

	fmt.Println(
		"  State: ACTIVE",
	)

	fmt.Println(
		"  Blocked releases:",
		len(
			state.QuarantinedVersions,
		),
	)

	fmt.Println(
		"  Latest failed:",
		state.QuarantinedVersion,
	)

	fmt.Println(
		"  Failures:",
		state.QuarantineFailureCount,
	)
}
