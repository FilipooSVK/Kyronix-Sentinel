package config

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {

	cfg := Default()

	if cfg.Daemon.Interval != 30*time.Second {

		t.Errorf(
			"interval mismatch: %v",
			cfg.Daemon.Interval,
		)
	}

	if cfg.History.Size != 1000 {

		t.Errorf(
			"history mismatch: %d",
			cfg.History.Size,
		)
	}

	if cfg.Host.Environment != "" {
		t.Errorf(
			"default host environment should be empty, got %q",
			cfg.Host.Environment,
		)
	}

	if cfg.Host.Role != "" {
		t.Errorf(
			"default host role should be empty, got %q",
			cfg.Host.Role,
		)
	}

	if cfg.Metrics.Enabled {
		t.Error(
			"metrics should be disabled by default",
		)
	}

	if cfg.Metrics.ListenAddress != "127.0.0.1:19130" {
		t.Errorf(
			"unexpected metrics listen address: %q",
			cfg.Metrics.ListenAddress,
		)
	}
}

func TestLoadMissingFileReturnsDefault(t *testing.T) {

	cfg, err := Load(
		"/tmp/nonexistent-sentinel.yaml",
	)

	if err != nil {
		t.Fatalf(
			"unexpected error: %v",
			err,
		)
	}

	if cfg.Daemon.Interval != 30*time.Second {

		t.Errorf(
			"expected default interval",
		)
	}
}
