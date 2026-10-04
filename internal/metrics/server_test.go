package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMetricsHandlerExportsRuntimeState(
	t *testing.T,
) {
	server := NewServer(
		"127.0.0.1:0",
	)

	server.Update(
		State{
			Up: true,

			HealthScore: 100,

			FreezeRisk: "LOW",

			PredictionScore: 10,

			PredictionConfidence: 90,

			ActiveSignals: 0,

			PersistentSignals: 0,

			KernelEvidence: false,

			Host: HostInfo{
				Hostname:       "kyronix-stratus-os",
				Environment:    "development",
				Role:           "stratus-os-development",
				Architecture:   "arm64",
				Virtualization: "lxc",
				Platform:       "",
				OSID:           "debian",
				OSVersion:      "13",
			},

			Collectors: []Collector{
				{
					Name:               "memory",
					Up:                 true,
					CollectionDuration: 2 * time.Millisecond,
				},
				{
					Name:               "host",
					Up:                 true,
					CollectionDuration: 7 * time.Millisecond,
				},
			},
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	response := httptest.NewRecorder()

	server.handler.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"unexpected status code: %d",
			response.Code,
		)
	}

	body := response.Body.String()

	expected := []string{
		"sentinel_up 1",
		"sentinel_health_score 100",
		"sentinel_freeze_risk_level 0",
		"sentinel_prediction_score 10",
		"sentinel_prediction_confidence_percent 90",
		"sentinel_prediction_active_signals 0",
		"sentinel_prediction_persistent_signals 0",
		"sentinel_prediction_kernel_evidence 0",
		`sentinel_host_info{hostname="kyronix-stratus-os",environment="development",role="stratus-os-development",architecture="arm64",virtualization="lxc",platform="",os_id="debian",os_version="13"} 1`,
		`sentinel_collector_up{collector="host"} 1`,
		`sentinel_collector_up{collector="memory"} 1`,
		`sentinel_collector_collection_duration_seconds{collector="host"} 0.007`,
		`sentinel_collector_collection_duration_seconds{collector="memory"} 0.002`,
	}

	for _, value := range expected {
		if !strings.Contains(
			body,
			value,
		) {
			t.Fatalf(
				"metrics output missing %q\n\n%s",
				value,
				body,
			)
		}
	}
}

func TestMetricsHandlerEscapesLabels(
	t *testing.T,
) {
	server := NewServer(
		"127.0.0.1:0",
	)

	server.Update(
		State{
			Host: HostInfo{
				Hostname: "host\"one\nlab",
			},
		},
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/metrics",
		nil,
	)

	response := httptest.NewRecorder()

	server.handler.ServeHTTP(
		response,
		request,
	)

	body := response.Body.String()

	if !strings.Contains(
		body,
		`hostname="host\"one\nlab"`,
	) {
		t.Fatalf(
			"label value was not escaped correctly:\n%s",
			body,
		)
	}
}

func TestRiskLevel(
	t *testing.T,
) {
	tests := map[string]int{
		"LOW":      0,
		"MEDIUM":   1,
		"HIGH":     2,
		"CRITICAL": 3,
		"unknown":  -1,
	}

	for risk, expected := range tests {
		if got := riskLevel(risk); got != expected {
			t.Fatalf(
				"risk %q: got %d, want %d",
				risk,
				got,
				expected,
			)
		}
	}
}
