package local

import (
	"testing"
)

func TestStatusStruct(t *testing.T) {

	status := Status{

		Running: true,

		HealthScore: 95,

		FreezeRisk: "LOW",

		Version: "0.1.0",
	}

	if !status.Running {

		t.Fatal(
			"expected running",
		)
	}
}
