package predictor

// ApplyKernelRisk adds kernel incident risk to an existing
// predictive risk assessment.
func ApplyKernelRisk(
	assessment RiskAssessment,
	events KernelEvents,
) RiskAssessment {

	score := assessment.Score

	reasons := append(
		[]string{},
		assessment.Reasons...,
	)

	// System-wide OOM kills are severe incidents.
	switch {

	case events.SystemOOMKills >= 3:

		score += 60

		reasons = append(
			reasons,
			"repeated system OOM kills detected",
		)

	case events.SystemOOMKills > 0:

		score += 40

		reasons = append(
			reasons,
			"system OOM kill detected",
		)
	}

	// A cgroup OOM kill already implies an OOM event.
	// Do not count CgroupOOMEvents again when kills exist.
	switch {

	case events.CgroupOOMKills >= 3:

		score += 45

		reasons = append(
			reasons,
			"repeated cgroup OOM kills detected",
		)

	case events.CgroupOOMKills > 0:

		score += 30

		reasons = append(
			reasons,
			"cgroup OOM kill detected",
		)

	case events.CgroupOOMEvents >= 3:

		score += 25

		reasons = append(
			reasons,
			"repeated cgroup OOM events detected",
		)

	case events.CgroupOOMEvents > 0:

		score += 15

		reasons = append(
			reasons,
			"cgroup OOM event detected",
		)
	}

	// Filesystem errors are treated as an independent failure signal.
	switch {

	case events.FilesystemErrors >= 10:

		score += 40

		reasons = append(
			reasons,
			"severe filesystem error activity detected",
		)

	case events.FilesystemErrors >= 3:

		score += 30

		reasons = append(
			reasons,
			"multiple filesystem errors detected",
		)

	case events.FilesystemErrors > 0:

		score += 15

		reasons = append(
			reasons,
			"filesystem error detected",
		)
	}

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
