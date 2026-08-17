package linux

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPressureCollectorCollectPressure(t *testing.T) {
	procRoot := t.TempDir()
	pressureRoot := filepath.Join(procRoot, "pressure")

	if err := os.MkdirAll(pressureRoot, 0o755); err != nil {
		t.Fatalf("failed to create pressure directory: %v", err)
	}

	fixtures := map[string]string{
		"cpu": `some avg10=40.84 avg60=40.48 avg300=16.22 total=60147130
full avg10=0.00 avg60=0.00 avg300=0.00 total=0
`,
		"memory": `some avg10=0.00 avg60=0.02 avg300=0.00 total=84011
full avg10=0.00 avg60=0.02 avg300=0.00 total=61584
`,
		"io": `some avg10=88.34 avg60=69.18 avg300=26.43 total=96315198
full avg10=27.22 avg60=21.90 avg300=9.46 total=36185662
`,
	}

	for name, data := range fixtures {
		path := filepath.Join(pressureRoot, name)

		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatalf("failed to create %s fixture: %v", name, err)
		}
	}

	collector := &PressureCollector{
		procRoot: procRoot,
	}

	stats, err := collector.CollectPressure(context.Background())
	if err != nil {
		t.Fatalf("CollectPressure returned error: %v", err)
	}

	if stats.CPU.Some.Avg10 != 40.84 {
		t.Errorf(
			"CPU some avg10 mismatch: got %.2f, want 40.84",
			stats.CPU.Some.Avg10,
		)
	}

	if stats.CPU.Some.TotalUS != 60147130 {
		t.Errorf(
			"CPU some total mismatch: got %d, want 60147130",
			stats.CPU.Some.TotalUS,
		)
	}

	if stats.CPU.Full.TotalUS != 0 {
		t.Errorf(
			"CPU full total mismatch: got %d, want 0",
			stats.CPU.Full.TotalUS,
		)
	}

	if stats.Memory.Full.Avg60 != 0.02 {
		t.Errorf(
			"memory full avg60 mismatch: got %.2f, want 0.02",
			stats.Memory.Full.Avg60,
		)
	}

	if stats.IO.Some.Avg10 != 88.34 {
		t.Errorf(
			"IO some avg10 mismatch: got %.2f, want 88.34",
			stats.IO.Some.Avg10,
		)
	}

	if stats.IO.Full.Avg10 != 27.22 {
		t.Errorf(
			"IO full avg10 mismatch: got %.2f, want 27.22",
			stats.IO.Full.Avg10,
		)
	}

	if stats.IO.Full.TotalUS != 36185662 {
		t.Errorf(
			"IO full total mismatch: got %d, want 36185662",
			stats.IO.Full.TotalUS,
		)
	}
}

func TestPressureCollectorUnavailable(t *testing.T) {
	collector := &PressureCollector{
		procRoot: t.TempDir(),
	}

	_, err := collector.CollectPressure(context.Background())
	if err == nil {
		t.Fatal("expected CollectPressure to return an error")
	}

	if !errors.Is(err, ErrPressureUnavailable) {
		t.Fatalf(
			"expected ErrPressureUnavailable, got %v",
			err,
		)
	}
}

func TestPressureCollectorRejectsInvalidData(t *testing.T) {
	procRoot := t.TempDir()
	pressureRoot := filepath.Join(procRoot, "pressure")

	if err := os.MkdirAll(pressureRoot, 0o755); err != nil {
		t.Fatalf("failed to create pressure directory: %v", err)
	}

	if err := os.WriteFile(
		filepath.Join(pressureRoot, "cpu"),
		[]byte("some avg10=invalid avg60=1.00 avg300=2.00 total=100\n"),
		0o600,
	); err != nil {
		t.Fatalf("failed to create CPU fixture: %v", err)
	}

	collector := &PressureCollector{
		procRoot: procRoot,
	}

	_, err := collector.CollectPressure(context.Background())
	if err == nil {
		t.Fatal("expected CollectPressure to return an error")
	}
}

func TestParsePressureWithoutFullRecord(t *testing.T) {
	data := `some avg10=1.25 avg60=2.50 avg300=3.75 total=123456
`

	stats, err := parsePressure(
		strings.NewReader(data),
	)
	if err != nil {
		t.Fatalf("parsePressure returned error: %v", err)
	}

	if stats.Some.Avg10 != 1.25 {
		t.Errorf(
			"some avg10 mismatch: got %.2f, want 1.25",
			stats.Some.Avg10,
		)
	}

	if stats.Full.TotalUS != 0 {
		t.Errorf(
			"expected zero-value full sample, got total %d",
			stats.Full.TotalUS,
		)
	}
}
