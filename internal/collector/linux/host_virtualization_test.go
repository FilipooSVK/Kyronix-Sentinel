package linux

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHostCollectorDetectsLXC(t *testing.T) {
	containerPath := filepath.Join(
		t.TempDir(),
		"container",
	)

	if err := os.WriteFile(
		containerPath,
		[]byte("lxc\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	collector := &HostCollector{
		systemdContainerPath: containerPath,
	}

	if got := collector.detectVirtualization(); got != "lxc" {
		t.Fatalf(
			"virtualization mismatch: got %q, want %q",
			got,
			"lxc",
		)
	}
}

func TestHostCollectorDetectsDocker(t *testing.T) {
	dockerEnvPath := filepath.Join(
		t.TempDir(),
		".dockerenv",
	)

	if err := os.WriteFile(
		dockerEnvPath,
		[]byte{},
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	collector := &HostCollector{
		dockerEnvPath: dockerEnvPath,
	}

	if got := collector.detectVirtualization(); got != "docker" {
		t.Fatalf(
			"virtualization mismatch: got %q, want %q",
			got,
			"docker",
		)
	}
}

func TestHostCollectorDetectsKVM(t *testing.T) {
	productNamePath := filepath.Join(
		t.TempDir(),
		"product_name",
	)

	if err := os.WriteFile(
		productNamePath,
		[]byte("KVM Virtual Machine\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	collector := &HostCollector{
		dmiProductNamePath: productNamePath,
	}

	if got := collector.detectVirtualization(); got != "kvm" {
		t.Fatalf(
			"virtualization mismatch: got %q, want %q",
			got,
			"kvm",
		)
	}
}

func TestHostCollectorFallsBackToBareMetal(t *testing.T) {
	root := t.TempDir()

	collector := &HostCollector{
		procRoot: filepath.Join(
			root,
			"proc",
		),

		systemdContainerPath: filepath.Join(
			root,
			"run",
			"systemd",
			"container",
		),

		dockerEnvPath: filepath.Join(
			root,
			".dockerenv",
		),

		dmiProductNamePath: filepath.Join(
			root,
			"product_name",
		),
	}

	if got := collector.detectVirtualization(); got != "bare-metal" {
		t.Fatalf(
			"virtualization mismatch: got %q, want %q",
			got,
			"bare-metal",
		)
	}
}

func TestHostCollectorDetectsProxmoxPlatform(t *testing.T) {
	membersPath := filepath.Join(
		t.TempDir(),
		".members",
	)

	if err := os.WriteFile(
		membersPath,
		[]byte("{}"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	collector := &HostCollector{
		proxmoxMembersPath: membersPath,
	}

	if got := collector.detectPlatform(); got != "proxmox" {
		t.Fatalf(
			"platform mismatch: got %q, want %q",
			got,
			"proxmox",
		)
	}
}

func TestHostCollectorDoesNotAssumeProxmoxForLXC(t *testing.T) {
	containerPath := filepath.Join(
		t.TempDir(),
		"container",
	)

	if err := os.WriteFile(
		containerPath,
		[]byte("lxc\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}

	collector := &HostCollector{
		systemdContainerPath: containerPath,
	}

	if got := collector.detectVirtualization(); got != "lxc" {
		t.Fatalf(
			"virtualization mismatch: got %q, want %q",
			got,
			"lxc",
		)
	}

	if got := collector.detectPlatform(); got != "" {
		t.Fatalf(
			"unexpected platform inference: %q",
			got,
		)
	}
}
