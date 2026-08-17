package local

import (
	"encoding/json"
	"net"
	"os"
	"path/filepath"
	"sync"
)

const DefaultSocket = "/run/sentinel/sentinel.sock"

// Server exposes Sentinel status API.
type Server struct {
	socket string

	mu sync.RWMutex

	status Status

	diagnostics Diagnostics

	prediction Prediction
}

// NewServer creates local API server.
func NewServer(
	socket string,
	status Status,
) *Server {

	return &Server{
		socket: socket,

		status: status,

		diagnostics: Diagnostics{},

		prediction: Prediction{},
	}
}

// Update changes runtime status.
func (s *Server) Update(
	status Status,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.status = status
}

// UpdatePrediction changes runtime prediction.
func (s *Server) UpdatePrediction(
	prediction Prediction,
) {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.prediction = prediction
}

// Start starts unix socket server.
func (s *Server) Start() error {

	dir := filepath.Dir(s.socket)

	if err := os.MkdirAll(
		dir,
		0755,
	); err != nil {
		return err
	}

	_ = os.Remove(s.socket)

	listener, err := net.Listen(
		"unix",
		s.socket,
	)

	if err != nil {
		return err
	}

	if err := os.Chmod(
		s.socket,
		0666,
	); err != nil {
		listener.Close()
		return err
	}

	go func() {

		defer listener.Close()

		for {

			conn, err := listener.Accept()

			if err != nil {
				continue
			}

			command := make(
				[]byte,
				32,
			)

			n, _ := conn.Read(command)

			request := string(command[:n])

			s.mu.RLock()

			status := s.status

			diagnostics := s.diagnostics

			prediction := s.prediction

			s.mu.RUnlock()

			switch request {

			case "diagnose":

				_ = json.NewEncoder(conn).Encode(
					diagnostics,
				)

			case "prediction":

				_ = json.NewEncoder(conn).Encode(
					prediction,
				)

			default:

				_ = json.NewEncoder(conn).Encode(
					status,
				)
			}

			conn.Close()
		}
	}()

	return nil
}
