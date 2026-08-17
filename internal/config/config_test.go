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
