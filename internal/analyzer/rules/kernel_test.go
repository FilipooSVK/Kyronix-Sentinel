package rules

import (
	"testing"

	"kyronix/sentinel/internal/domain"
)

func TestKernelRuleHealthy(t *testing.T) {

	rule := NewKernelRule()

	findings := rule.Evaluate(
		domain.Snapshot{},
	)

	if len(findings) != 0 {
		t.Fatalf(
			"expected no findings, got %d",
			len(findings),
		)
	}
}

func TestKernelRuleSystemOOM(t *testing.T) {

	value := uint64(1)

	rule := NewKernelRule()

	snapshot := domain.Snapshot{
		Kernel: domain.KernelStats{
			OOM: domain.OOMStats{
				SystemKillCount: &value,
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

	if findings[0].Penalty != 40 {
		t.Errorf(
			"penalty mismatch: got %d",
			findings[0].Penalty,
		)
	}
}

func TestKernelRuleCgroupOOM(t *testing.T) {

	value := uint64(2)

	rule := NewKernelRule()

	snapshot := domain.Snapshot{
		Kernel: domain.KernelStats{
			OOM: domain.OOMStats{
				CgroupKillCount: &value,
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

func TestKernelRuleFilesystemErrors(t *testing.T) {

	rule := NewKernelRule()

	snapshot := domain.Snapshot{
		Kernel: domain.KernelStats{
			FilesystemErrors: 5,
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
