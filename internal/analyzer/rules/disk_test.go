package rules

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestDiskRuleHealthy(t *testing.T) {

	rule := NewDiskRule()

	snapshot := domain.Snapshot{
		Disks: []domain.DiskStats{
			{
				MountPoint:     "/",
				UsedPercent:    50,
				AvailableBytes: 50 * 1024 * 1024 * 1024,
			},
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 0 {
		t.Fatalf(
			"expected no findings, got %d",
			len(findings),
		)
	}
}

func TestDiskRuleHighUsage(t *testing.T) {

	rule := NewDiskRule()

	snapshot := domain.Snapshot{
		Disks: []domain.DiskStats{
			{
				MountPoint:     "/",
				UsedPercent:    90,
				AvailableBytes: 10 * 1024 * 1024 * 1024,
			},
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 1 {
		t.Fatalf(
			"expected one finding, got %d",
			len(findings),
		)
	}

	if findings[0].Penalty != 10 {
		t.Errorf(
			"penalty mismatch: got %d",
			findings[0].Penalty,
		)
	}
}

func TestDiskRuleCriticalUsage(t *testing.T) {

	rule := NewDiskRule()

	snapshot := domain.Snapshot{
		Disks: []domain.DiskStats{
			{
				MountPoint:     "/",
				UsedPercent:    97,
				AvailableBytes: 5 * 1024 * 1024 * 1024,
			},
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 1 {
		t.Fatalf(
			"expected one finding, got %d",
			len(findings),
		)
	}

	if findings[0].Severity != domain.SeverityCritical {
		t.Errorf(
			"severity mismatch: got %s",
			findings[0].Severity,
		)
	}
}

func TestDiskRuleLowSpace(t *testing.T) {

	rule := NewDiskRule()

	snapshot := domain.Snapshot{
		Disks: []domain.DiskStats{
			{
				MountPoint:     "/",
				UsedPercent:    50,
				AvailableBytes: 500 * 1024 * 1024,
			},
		},
	}

	findings := rule.Evaluate(snapshot)

	if len(findings) != 1 {
		t.Fatalf(
			"expected one finding, got %d",
			len(findings),
		)
	}
}
