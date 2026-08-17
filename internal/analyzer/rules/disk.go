package rules

import (
	"kyronix/sentinel/internal/domain"
)

// DiskRule evaluates filesystem capacity risks.
type DiskRule struct{}

// NewDiskRule creates a disk health rule.
func NewDiskRule() DiskRule {
	return DiskRule{}
}

// Evaluate analyzes filesystem utilization.
func (DiskRule) Evaluate(
	snapshot domain.Snapshot,
) []domain.Finding {

	findings := make([]domain.Finding, 0)

	for _, disk := range snapshot.Disks {

		switch {
		case disk.UsedPercent >= 95:
			findings = append(findings, domain.Finding{
				Severity: domain.SeverityCritical,
				Source:   "disk",
				Message:  "Critical filesystem utilization detected on " + disk.MountPoint,
				Penalty:  25,
			})

		case disk.UsedPercent >= 85:
			findings = append(findings, domain.Finding{
				Severity: domain.SeverityWarning,
				Source:   "disk",
				Message:  "High filesystem utilization detected on " + disk.MountPoint,
				Penalty:  10,
			})
		}

		if disk.AvailableBytes < 1024*1024*1024 {
			findings = append(findings, domain.Finding{
				Severity: domain.SeverityWarning,
				Source:   "disk",
				Message:  "Low available disk space on " + disk.MountPoint,
				Penalty:  10,
			})
		}
	}

	return findings
}
