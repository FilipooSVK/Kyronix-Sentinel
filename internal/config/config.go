package config

import "time"

// Config represents Sentinel configuration.
type Config struct {
	Daemon DaemonConfig `yaml:"daemon"`

	History HistoryConfig `yaml:"history"`

	Logging LoggingConfig `yaml:"logging"`

	Update UpdateConfig `yaml:"update"`
}

// DaemonConfig controls runtime behaviour.
type DaemonConfig struct {
	Interval time.Duration `yaml:"interval"`
}

// HistoryConfig controls history storage.
type HistoryConfig struct {
	Size int `yaml:"size"`

	Persistence bool `yaml:"persistence"`

	Path string `yaml:"path"`
}

// LoggingConfig controls logging.
type LoggingConfig struct {
	Level string `yaml:"level"`
}

// UpdateConfig controls Sentinel software update behaviour.
type UpdateConfig struct {
	Enabled bool `yaml:"enabled"`

	Owner string `yaml:"owner"`

	Repository string `yaml:"repository"`

	AutoCheck bool `yaml:"auto_check"`

	AutoInstall bool `yaml:"auto_install"`

	CheckInterval time.Duration `yaml:"check_interval"`
}

// Default returns default Sentinel configuration.
func Default() Config {

	return Config{
		Daemon: DaemonConfig{
			Interval: 30 * time.Second,
		},

		History: HistoryConfig{
			Size:        1000,
			Persistence: true,
			Path:        "/var/lib/sentinel/history.jsonl",
		},

		Logging: LoggingConfig{
			Level: "info",
		},

		Update: UpdateConfig{
			Enabled:       true,
			Owner:         "",
			Repository:    "",
			AutoCheck:     true,
			AutoInstall:   false,
			CheckInterval: 24 * time.Hour,
		},
	}
}
