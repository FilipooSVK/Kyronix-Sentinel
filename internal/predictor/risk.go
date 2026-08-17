package predictor

// RiskAssessment represents combined predictive risk.
type RiskAssessment struct {
	Score   int
	Level   RiskLevel
	Reasons []string
}

// CalculateRisk calculates predictive risk from multiple trends.
func CalculateRisk(
	trends []Trend,
) RiskAssessment {

	score := 0
	reasons := []string{}

	for _, trend := range trends {

		switch trend.Metric {

		case "memory":

			if trend.Direction == TrendIncreasing {

				switch {

				case trend.Rate >= 10:

					score += 50

					reasons = append(
						reasons,
						"high memory growth rate",
					)

				case trend.Rate >= 5:

					score += 30

					reasons = append(
						reasons,
						"memory growth detected",
					)

				case trend.Rate > 0:

					score += 10

					reasons = append(
						reasons,
						"slow memory growth",
					)
				}
			}

			if trend.Current >= 90 {

				score += 30

				reasons = append(
					reasons,
					"memory utilization critically high",
				)

			} else if trend.Current >= 80 {

				score += 15

				reasons = append(
					reasons,
					"memory utilization high",
				)
			}

		case "cpu_pressure_some_avg10":

			switch {

			case trend.Current >= 70:

				score += 30

				reasons = append(
					reasons,
					"critical CPU pressure",
				)

			case trend.Current >= 40:

				score += 20

				reasons = append(
					reasons,
					"high CPU pressure",
				)

			case trend.Current >= 20:

				score += 10

				reasons = append(
					reasons,
					"elevated CPU pressure",
				)
			}

		case "memory_pressure_some_avg10":

			switch {

			case trend.Current >= 30:

				score += 40

				reasons = append(
					reasons,
					"critical memory pressure",
				)

			case trend.Current >= 15:

				score += 25

				reasons = append(
					reasons,
					"high memory pressure",
				)

			case trend.Current >= 5:

				score += 10

				reasons = append(
					reasons,
					"elevated memory pressure",
				)
			}

		case "io_pressure_some_avg10":

			switch {

			case trend.Current >= 70:

				score += 40

				reasons = append(
					reasons,
					"critical I/O pressure",
				)

			case trend.Current >= 40:

				score += 25

				reasons = append(
					reasons,
					"high I/O pressure",
				)

			case trend.Current >= 20:

				score += 10

				reasons = append(
					reasons,
					"elevated I/O pressure",
				)
			}

		case "health_score":

			if trend.Direction == TrendDecreasing {

				switch {

				case trend.Current <= 30:

					score += 30

					reasons = append(
						reasons,
						"health score critically low and degrading",
					)

				case trend.Current <= 50:

					score += 20

					reasons = append(
						reasons,
						"health score significantly degrading",
					)

				case trend.Current <= 70:

					score += 10

					reasons = append(
						reasons,
						"health score degrading",
					)
				}
			}
		}
	}

	// Keep public risk score within 0-100.
	if score > 100 {
		score = 100
	}

	level := RiskLow

	switch {

	case score >= 80:

		level = RiskCritical

	case score >= 60:

		level = RiskHigh

	case score >= 30:

		level = RiskMedium
	}

	return RiskAssessment{
		Score:   score,
		Level:   level,
		Reasons: reasons,
	}
}
