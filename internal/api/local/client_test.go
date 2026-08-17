package local

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestGetStatusSendsStatusCommand(
	t *testing.T,
) {

	socket := filepath.Join(
		t.TempDir(),
		"sentinel.sock",
	)

	listener, err := net.Listen(
		"unix",
		socket,
	)

	if err != nil {
		t.Fatal(err)
	}

	defer listener.Close()

	done := make(
		chan string,
		1,
	)

	go func() {

		conn, err := listener.Accept()

		if err != nil {
			return
		}

		defer conn.Close()

		buffer := make(
			[]byte,
			32,
		)

		n, err := conn.Read(
			buffer,
		)

		if err != nil {
			return
		}

		done <- string(
			buffer[:n],
		)

		_ = json.NewEncoder(
			conn,
		).Encode(
			Status{
				Running: true,

				HealthScore: 100,

				FreezeRisk: "LOW",

				Version: "0.1.0",
			},
		)
	}()

	status, err := GetStatus(
		socket,
	)

	if err != nil {
		t.Fatal(err)
	}

	if !status.Running {

		t.Fatal(
			"expected Sentinel running",
		)
	}

	if status.Version != "0.1.0" {

		t.Fatalf(
			"unexpected version: %s",
			status.Version,
		)
	}

	select {

	case command := <-done:

		if command != "status" {

			t.Fatalf(
				"expected status command, got %q",
				command,
			)
		}

	case <-time.After(
		time.Second,
	):

		t.Fatal(
			"status command was not received",
		)
	}
}
