package metrics

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

const metricsPath = "/metrics"

// HostInfo contains host labels exported by Sentinel.
type HostInfo struct {
	Hostname       string
	Environment    string
	Role           string
	Architecture   string
	Virtualization string
	Platform       string
	OSID           string
	OSVersion      string
}

// Collector contains Prometheus data for one Sentinel collector.
type Collector struct {
	Name string

	Up bool

	CollectionDuration time.Duration
}

// State represents the current Sentinel runtime state exported as metrics.
type State struct {
	Up bool

	HealthScore int

	FreezeRisk string

	PredictionScore int

	PredictionConfidence float64

	ActiveSignals int

	PersistentSignals int

	KernelEvidence bool

	Host HostInfo

	Collectors []Collector
}

// Server exposes Sentinel metrics using the Prometheus text format.
type Server struct {
	address string

	mu sync.RWMutex

	state State

	handler http.Handler

	httpServer *http.Server
}

// NewServer creates a Prometheus-compatible metrics server.
func NewServer(
	address string,
) *Server {
	server := &Server{
		address: address,
	}

	mux := http.NewServeMux()

	mux.HandleFunc(
		metricsPath,
		server.handleMetrics,
	)

	server.handler = mux

	return server
}

// Update replaces the current metrics state.
func (s *Server) Update(
	state State,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.state = state
}

// Start starts the metrics HTTP server.
//
// The TCP listener is opened synchronously so bind errors are returned
// directly to the caller.
func (s *Server) Start() error {
	listener, err := net.Listen(
		"tcp",
		s.address,
	)

	if err != nil {
		return fmt.Errorf(
			"listen on %s: %w",
			s.address,
			err,
		)
	}

	s.httpServer = &http.Server{
		Handler: s.handler,
	}

	go func() {
		_ = s.httpServer.Serve(
			listener,
		)
	}()

	return nil
}

// Shutdown gracefully stops the metrics HTTP server.
func (s *Server) Shutdown(
	ctx context.Context,
) error {
	if s.httpServer == nil {
		return nil
	}

	return s.httpServer.Shutdown(
		ctx,
	)
}

func (s *Server) handleMetrics(
	w http.ResponseWriter,
	_ *http.Request,
) {
	s.mu.RLock()
	state := s.state
	s.mu.RUnlock()

	w.Header().Set(
		"Content-Type",
		"text/plain; version=0.0.4; charset=utf-8",
	)

	writeMetricHeader(
		w,
		"sentinel_up",
		"Whether the Sentinel daemon is running.",
	)

	fmt.Fprintf(
		w,
		"sentinel_up %d\n",
		boolValue(state.Up),
	)

	writeMetricHeader(
		w,
		"sentinel_health_score",
		"Current Sentinel host health score.",
	)

	fmt.Fprintf(
		w,
		"sentinel_health_score %d\n",
		state.HealthScore,
	)

	writeMetricHeader(
		w,
		"sentinel_freeze_risk_level",
		"Current freeze risk level: LOW=0, MEDIUM=1, HIGH=2, CRITICAL=3, unknown=-1.",
	)

	fmt.Fprintf(
		w,
		"sentinel_freeze_risk_level %d\n",
		riskLevel(state.FreezeRisk),
	)

	writeMetricHeader(
		w,
		"sentinel_prediction_score",
		"Current Sentinel prediction score.",
	)

	fmt.Fprintf(
		w,
		"sentinel_prediction_score %d\n",
		state.PredictionScore,
	)

	writeMetricHeader(
		w,
		"sentinel_prediction_confidence_percent",
		"Current Sentinel prediction confidence percentage.",
	)

	fmt.Fprintf(
		w,
		"sentinel_prediction_confidence_percent %v\n",
		state.PredictionConfidence,
	)

	writeMetricHeader(
		w,
		"sentinel_prediction_active_signals",
		"Number of currently active prediction signals.",
	)

	fmt.Fprintf(
		w,
		"sentinel_prediction_active_signals %d\n",
		state.ActiveSignals,
	)

	writeMetricHeader(
		w,
		"sentinel_prediction_persistent_signals",
		"Number of persistent prediction signals.",
	)

	fmt.Fprintf(
		w,
		"sentinel_prediction_persistent_signals %d\n",
		state.PersistentSignals,
	)

	writeMetricHeader(
		w,
		"sentinel_prediction_kernel_evidence",
		"Whether kernel evidence contributes to the current prediction.",
	)

	fmt.Fprintf(
		w,
		"sentinel_prediction_kernel_evidence %d\n",
		boolValue(
			state.KernelEvidence,
		),
	)

	writeMetricHeader(
		w,
		"sentinel_host_info",
		"Static information about the monitored Sentinel host.",
	)

	fmt.Fprintf(
		w,
		"sentinel_host_info{hostname=\"%s\",environment=\"%s\",role=\"%s\",architecture=\"%s\",virtualization=\"%s\",platform=\"%s\",os_id=\"%s\",os_version=\"%s\"} 1\n",
		escapeLabelValue(state.Host.Hostname),
		escapeLabelValue(state.Host.Environment),
		escapeLabelValue(state.Host.Role),
		escapeLabelValue(state.Host.Architecture),
		escapeLabelValue(state.Host.Virtualization),
		escapeLabelValue(state.Host.Platform),
		escapeLabelValue(state.Host.OSID),
		escapeLabelValue(state.Host.OSVersion),
	)

	collectors := append(
		[]Collector(nil),
		state.Collectors...,
	)

	sort.Slice(
		collectors,
		func(i, j int) bool {
			return collectors[i].Name <
				collectors[j].Name
		},
	)

	writeMetricHeader(
		w,
		"sentinel_collector_up",
		"Whether a Sentinel collector completed successfully.",
	)

	for _, collector := range collectors {
		fmt.Fprintf(
			w,
			"sentinel_collector_up{collector=\"%s\"} %d\n",
			escapeLabelValue(
				collector.Name,
			),
			boolValue(
				collector.Up,
			),
		)
	}

	writeMetricHeader(
		w,
		"sentinel_collector_collection_duration_seconds",
		"Duration of the latest Sentinel collector execution in seconds.",
	)

	for _, collector := range collectors {
		fmt.Fprintf(
			w,
			"sentinel_collector_collection_duration_seconds{collector=\"%s\"} %v\n",
			escapeLabelValue(
				collector.Name,
			),
			collector.CollectionDuration.Seconds(),
		)
	}
}

func writeMetricHeader(
	w http.ResponseWriter,
	name string,
	help string,
) {
	fmt.Fprintf(
		w,
		"# HELP %s %s\n",
		name,
		help,
	)

	fmt.Fprintf(
		w,
		"# TYPE %s gauge\n",
		name,
	)
}

func boolValue(
	value bool,
) int {
	if value {
		return 1
	}

	return 0
}

func riskLevel(
	risk string,
) int {
	switch strings.ToUpper(
		strings.TrimSpace(risk),
	) {
	case "LOW":
		return 0

	case "MEDIUM":
		return 1

	case "HIGH":
		return 2

	case "CRITICAL":
		return 3

	default:
		return -1
	}
}

func escapeLabelValue(
	value string,
) string {
	replacer := strings.NewReplacer(
		`\`,
		`\\`,
		"\n",
		`\n`,
		`"`,
		`\"`,
	)

	return replacer.Replace(
		value,
	)
}
