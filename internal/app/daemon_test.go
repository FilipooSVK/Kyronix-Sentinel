package app

import (
	"context"
	"kyronix/sentinel/internal/api/local"
	"os"
	"testing"
	"time"

	"kyronix/sentinel/internal/domain"
	"kyronix/sentinel/internal/history"
	"kyronix/sentinel/internal/logging"
)

type daemonCollector struct{}

func (daemonCollector) Collect() domain.Snapshot {

	return domain.Snapshot{
		Memory: domain.MemoryStats{
			UsedPercent: 50,
		},
	}
}

type daemonAnalyzer struct{}

func (daemonAnalyzer) Analyze(
	domain.Snapshot,
) domain.HealthResult {

	return domain.HealthResult{
		HealthScore: 100,
		FreezeRisk:  domain.RiskLow,
	}
}

func TestDaemonStopsGracefully(t *testing.T) {

	engine := NewEngine(
		daemonCollector{},
		daemonAnalyzer{},
		history.NewStore(10),
	)

	logger := logging.New(
		os.Stdout,
		"info",
	)

	daemon := NewDaemon(
		engine,
		10*time.Millisecond,
		logger,
	)

	daemon.statusServer = local.NewServer(
		"/tmp/sentinel-test.sock",
		local.Status{},
	)

	daemon.statusServer = local.NewServer(
		"/tmp/sentinel-test.sock",
		local.Status{},
	)

	os.Remove("/tmp/sentinel-test.sock")

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	done := make(chan error)

	go func() {

		done <- daemon.Run(ctx)

	}()

	time.Sleep(
		50 * time.Millisecond,
	)

	cancel()

	select {

	case err := <-done:

		if err != nil {

			t.Fatalf(
				"daemon returned error: %v",
				err,
			)
		}

	case <-time.After(
		time.Second,
	):

		t.Fatal(
			"daemon did not stop",
		)
	}
}
