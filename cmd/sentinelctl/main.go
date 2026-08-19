package main

import (
	"fmt"
	"os"

	"kyronix/sentinel/internal/api/local"
	"kyronix/sentinel/internal/version"
)

func main() {

	if len(os.Args) < 2 {

		fmt.Println("usage: sentinelctl <command>")
		fmt.Println()
		fmt.Println("commands:")
		fmt.Println("  status")
		fmt.Println("  diagnose")
		fmt.Println("  prediction")
		fmt.Println("  update")
		fmt.Println("  version")

		return
	}

	switch os.Args[1] {

	case "version":

		info := version.Current()

		fmt.Println("Kyronix Sentinel")
		fmt.Println("Version:", info.Version)

	case "update":

		os.Exit(
			runUpdateCommand(
				os.Args[2:],
			),
		)

	case "diagnose":

		diagnostics, err := local.GetDiagnostics(
			local.DefaultSocket,
		)

		if err != nil {

			fmt.Println(
				"Sentinel unavailable:",
				err,
			)

			return
		}

		fmt.Println("Kyronix Sentinel Diagnostics")
		fmt.Println()

		fmt.Println(
			"Running:",
			diagnostics.Running,
		)

		fmt.Println(
			"Health:",
			diagnostics.HealthScore,
		)

		fmt.Println(
			"Risk:",
			diagnostics.FreezeRisk,
		)

		fmt.Println()
		fmt.Println("Collectors:")

		for _, collector := range diagnostics.Collectors {

			fmt.Println()
			fmt.Println(collector.Name)

			fmt.Println(
				"  State:",
				collector.State,
			)

			fmt.Println(
				"  Collection:",
				collector.CollectionMS,
				"ms",
			)

			if collector.LastSuccess != nil {

				fmt.Println(
					"  Last success:",
					collector.LastSuccess.Format(
						"15:04:05",
					),
				)
			}

			if collector.Message != "" {

				fmt.Println(
					"  Message:",
					collector.Message,
				)
			}
		}

	case "prediction":

		prediction, err := local.GetPrediction(
			local.DefaultSocket,
		)

		if err != nil {

			fmt.Println(
				"Sentinel unavailable:",
				err,
			)

			return
		}

		fmt.Println("Kyronix Sentinel Prediction")
		fmt.Println()

		fmt.Println(
			"Risk:",
			prediction.Risk,
		)

		fmt.Println(
			"Score:",
			prediction.Score,
		)

		fmt.Printf(
			"Confidence: %.0f%%\n",
			prediction.Confidence,
		)

		fmt.Println(
			"Recommendation:",
			prediction.Recommendation,
		)

		fmt.Println()
		fmt.Println("Consensus:")

		fmt.Println(
			"  Active signals:",
			prediction.ActiveSignals,
		)

		fmt.Println(
			"  Persistent signals:",
			prediction.PersistentSignals,
		)

		fmt.Println(
			"  Kernel evidence:",
			yesNo(
				prediction.KernelEvidence,
			),
		)

		if len(prediction.Signals) > 0 {

			fmt.Println()
			fmt.Println("Signals:")

			for _, signal := range prediction.Signals {

				fmt.Println(
					"  -",
					signal,
				)
			}
		}

		if len(prediction.Reasons) > 0 {

			fmt.Println()
			fmt.Println("Reasons:")

			for _, reason := range prediction.Reasons {

				fmt.Println(
					"  -",
					reason,
				)
			}
		}

	case "status":

		status, err := local.GetStatus(
			local.DefaultSocket,
		)

		if err != nil {

			fmt.Println(
				"Sentinel unavailable:",
				err,
			)

			return
		}

		fmt.Println("Kyronix Sentinel")

		fmt.Println(
			"Running:",
			status.Running,
		)

		fmt.Println(
			"Health:",
			status.HealthScore,
		)

		fmt.Println(
			"Risk:",
			status.FreezeRisk,
		)

	default:

		fmt.Println(
			"unknown command:",
			os.Args[1],
		)
	}
}

func yesNo(
	value bool,
) string {

	if value {
		return "yes"
	}

	return "no"
}
