package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultUpdateConfiguration(
	t *testing.T,
) {

	cfg := Default()

	if !cfg.Update.Enabled {
		t.Fatal(
			"expected updater to be enabled by default",
		)
	}

	if !cfg.Update.AutoCheck {
		t.Fatal(
			"expected automatic update checks to be enabled",
		)
	}

	if cfg.Update.AutoInstall {
		t.Fatal(
			"automatic installation must be disabled by default",
		)
	}

	if cfg.Update.AutoInstallPolicy.MinReleaseAge !=
		24*time.Hour {

		t.Fatalf(
			"expected 24h minimum release age, got %s",
			cfg.Update.AutoInstallPolicy.MinReleaseAge,
		)
	}

	if !cfg.Update.AutoInstallPolicy.PatchOnly {

		t.Fatal(
			"automatic installation must default to patch-only",
		)
	}

	if cfg.Update.CheckInterval != 24*time.Hour {
		t.Fatalf(
			"expected 24h check interval, got %s",
			cfg.Update.CheckInterval,
		)
	}

	if cfg.Update.StatePath != "/var/lib/sentinel/update-state.json" {
		t.Fatalf(
			"unexpected update state path: %s",
			cfg.Update.StatePath,
		)
	}
}

func TestLoadUpdateConfiguration(
	t *testing.T,
) {

	path := t.TempDir() + "/sentinel.yaml"

	data := []byte(`
update:
  enabled: true
  owner: example
  repository: sentinel
  auto_check: false
  auto_install: false
  auto_install_policy:
    min_release_age: 12h
    patch_only: false
  check_interval: 6h
  state_path: /tmp/sentinel-update-state.json
`)

	if err := os.WriteFile(
		path,
		data,
		0644,
	); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(
		path,
	)

	if err != nil {
		t.Fatal(err)
	}

	if cfg.Update.Owner != "example" {
		t.Fatalf(
			"unexpected owner: %s",
			cfg.Update.Owner,
		)
	}

	if cfg.Update.Repository != "sentinel" {
		t.Fatalf(
			"unexpected repository: %s",
			cfg.Update.Repository,
		)
	}

	if cfg.Update.AutoCheck {
		t.Fatal(
			"expected auto_check false",
		)
	}

	if cfg.Update.AutoInstall {
		t.Fatal(
			"expected auto_install false",
		)
	}

	if cfg.Update.AutoInstallPolicy.MinReleaseAge !=
		12*time.Hour {

		t.Fatalf(
			"expected 12h minimum release age, got %s",
			cfg.Update.AutoInstallPolicy.MinReleaseAge,
		)
	}

	if cfg.Update.AutoInstallPolicy.PatchOnly {

		t.Fatal(
			"expected patch_only false",
		)
	}

	if cfg.Update.CheckInterval != 6*time.Hour {
		t.Fatalf(
			"expected 6h interval, got %s",
			cfg.Update.CheckInterval,
		)
	}

	if cfg.Update.StatePath != "/tmp/sentinel-update-state.json" {
		t.Fatalf(
			"unexpected state path: %s",
			cfg.Update.StatePath,
		)
	}
}
