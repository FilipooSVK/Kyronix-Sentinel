package rules

import (
	"kyronix/sentinel/internal/domain"
)

// KernelRule evaluates kernel instability signals.
type KernelRule struct{}

// NewKernelRule creates kernel health rule.
func NewKernelRule() KernelRule {
	return KernelRule{}
}

// Evaluate checks OOM and kernel errors.
func (KernelRule) Evaluate(
	snapshot domain.Snapshot,
) []domain.Finding {

	findings := make([]domain.Finding, 0)

	oom := snapshot.Kernel.OOM

	if oom.SystemKillCount != nil && *oom.SystemKillCount > 0 {
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "kernel",
			Message:  "System OOM killer activity detected",
			Penalty:  40,
		})
	}

	if oom.CgroupKillCount != nil && *oom.CgroupKillCount > 0 {
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "kernel",
			Message:  "Cgroup OOM killer activity detected",
			Penalty:  30,
		})
	}

	if snapshot.Kernel.FilesystemErrors > 0 {
		findings = append(findings, domain.Finding{
			Severity: domain.SeverityCritical,
			Source:   "kernel",
			Message:  "Filesystem errors detected",
			Penalty:  25,
		})
	}

	return findings
}
