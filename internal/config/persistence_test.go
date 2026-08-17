package config

import (
	"os"
	"testing"
	"time"
)

func TestDefaultHistoryPersistence(
	t *testing.T,
) {

	cfg := Default()

	if !cfg.History.Persistence {
		t.Fatal(
			"expected history persistence to be enabled by default",
		)
	}

	if cfg.History.Path != "/var/lib/sentinel/history.jsonl" {
		t.Fatalf(
			"unexpected history path: %s",
			cfg.History.Path,
		)
	}

	if cfg.History.Size != 1000 {
		t.Fatalf(
			"expected history size 1000, got %d",
			cfg.History.Size,
		)
	}
}

func TestLoadHistoryPersistenceConfiguration(
	t *testing.T,
) {

	path := t.TempDir() + "/sentinel.yaml"

	data := []byte(`
history:
  size: 250
  persistence: false
  path: /tmp/sentinel-history.jsonl
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

	if cfg.History.Size != 250 {
		t.Fatalf(
			"expected history size 250, got %d",
			cfg.History.Size,
		)
	}

	if cfg.History.Persistence {
		t.Fatal(
			"expected history persistence to be disabled",
		)
	}

	if cfg.History.Path != "/tmp/sentinel-history.jsonl" {
		t.Fatalf(
			"unexpected history path: %s",
			cfg.History.Path,
		)
	}

	// Values not specified in YAML must retain their defaults.
	if cfg.Daemon.Interval != 30*time.Second {
		t.Fatalf(
			"expected default daemon interval 30s, got %s",
			cfg.Daemon.Interval,
		)
	}

	if cfg.Logging.Level != "info" {
		t.Fatalf(
			"expected default logging level info, got %s",
			cfg.Logging.Level,
		)
	}
}
