package updater

import (
	"testing"
	"time"
)

func policyTestInput() AutoInstallPolicyInput {

	now := time.Date(
		2026,
		time.August,
		18,
		8,
		0,
		0,
		0,
		time.UTC,
	)

	return AutoInstallPolicyInput{
		Check: CheckResult{
			CurrentVersion: "v0.1.1",

			LatestVersion: "v0.1.2",

			UpdateAvailable: true,

			Release: Release{
				TagName: "v0.1.2",

				PublishedAt: now.Add(
					-48 * time.Hour,
				),
			},
		},

		Now: now,
	}
}

func enabledPatchPolicy() AutoInstallPolicy {

	return AutoInstallPolicy{
		Enabled: true,

		MinReleaseAge: 24 * time.Hour,

		PatchOnly: true,
	}
}

func decisionHasReason(
	decision AutoInstallDecision,
	reason string,
) bool {

	for _, existing := range decision.Reasons {

		if existing == reason {
			return true
		}
	}

	return false
}

func TestAutoInstallPolicyAllowsStablePatch(
	t *testing.T,
) {

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			policyTestInput(),
		)

	if err != nil {
		t.Fatal(err)
	}

	if !decision.Allowed {

		t.Fatalf(
			"expected automatic install to be allowed, reasons: %v",
			decision.Reasons,
		)
	}

	if decision.ReleaseAge !=
		48*time.Hour {

		t.Fatalf(
			"unexpected release age: %s",
			decision.ReleaseAge,
		)
	}
}

func TestAutoInstallPolicyDisabled(
	t *testing.T,
) {

	input := policyTestInput()

	policy := enabledPatchPolicy()

	policy.Enabled = false

	decision, err :=
		EvaluateAutoInstallPolicy(
			policy,
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"disabled automatic install was allowed",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonDisabled,
	) {

		t.Fatalf(
			"missing disabled reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksYoungRelease(
	t *testing.T,
) {

	input := policyTestInput()

	input.Check.Release.PublishedAt =
		input.Now.Add(
			-2 * time.Hour,
		)

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"young release was allowed",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonReleaseTooYoung,
	) {

		t.Fatalf(
			"missing release-too-young reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksMinorUpgrade(
	t *testing.T,
) {

	input := policyTestInput()

	input.Check.LatestVersion = "v0.2.0"

	input.Check.Release.TagName = "v0.2.0"

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"minor upgrade was allowed by patch-only policy",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonNonPatchUpgrade,
	) {

		t.Fatalf(
			"missing non-patch reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyCanAllowMinorUpgrade(
	t *testing.T,
) {

	input := policyTestInput()

	input.Check.LatestVersion = "v0.2.0"

	input.Check.Release.TagName = "v0.2.0"

	policy := enabledPatchPolicy()

	policy.PatchOnly = false

	decision, err :=
		EvaluateAutoInstallPolicy(
			policy,
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if !decision.Allowed {

		t.Fatalf(
			"minor upgrade should be allowed when patch-only is disabled: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksPrerelease(
	t *testing.T,
) {

	input := policyTestInput()

	input.Check.LatestVersion =
		"v0.1.2-rc.1"

	input.Check.Release.TagName =
		"v0.1.2-rc.1"

	input.Check.Release.Prerelease =
		true

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"prerelease was allowed",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonPrerelease,
	) {

		t.Fatalf(
			"missing prerelease reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksQuarantinedRelease(
	t *testing.T,
) {

	input := policyTestInput()

	input.State.QuarantinedVersion =
		"v0.1.2"

	input.State.QuarantinedVersions =
		[]string{
			"v0.1.2",
		}

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"quarantined release was allowed",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonQuarantined,
	) {

		t.Fatalf(
			"missing quarantine reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksInstallInProgress(
	t *testing.T,
) {

	input := policyTestInput()

	input.State.LastInstallResult =
		InstallResultInProgress

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"concurrent installation was allowed",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonInstallInProgress,
	) {

		t.Fatalf(
			"missing install-in-progress reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksUnverifiedRollback(
	t *testing.T,
) {

	input := policyTestInput()

	input.State.LastRollback = true

	input.State.LastRollbackVerified = false

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"automatic install was allowed after unverified rollback",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonUnverifiedRollback,
	) {

		t.Fatalf(
			"missing rollback reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyBlocksMissingReleaseTime(
	t *testing.T,
) {

	input := policyTestInput()

	input.Check.Release.PublishedAt =
		time.Time{}

	decision, err :=
		EvaluateAutoInstallPolicy(
			enabledPatchPolicy(),
			input,
		)

	if err != nil {
		t.Fatal(err)
	}

	if decision.Allowed {

		t.Fatal(
			"release without publication time was allowed",
		)
	}

	if !decisionHasReason(
		decision,
		AutoInstallReasonReleaseTimeMissing,
	) {

		t.Fatalf(
			"missing publication-time reason: %v",
			decision.Reasons,
		)
	}
}

func TestAutoInstallPolicyRejectsNegativeReleaseAgePolicy(
	t *testing.T,
) {

	policy := enabledPatchPolicy()

	policy.MinReleaseAge = -time.Hour

	_, err :=
		EvaluateAutoInstallPolicy(
			policy,
			policyTestInput(),
		)

	if err == nil {

		t.Fatal(
			"expected negative minimum release age error",
		)
	}
}
