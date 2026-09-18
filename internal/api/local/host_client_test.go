package local

import (
	"encoding/json"
	"net"
	"path/filepath"
	"testing"
	"time"
)

func TestGetHostSendsHostCommand(
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

		n, err := conn.Read(buffer)
		if err != nil {
			return
		}

		done <- string(
			buffer[:n],
		)

		_ = json.NewEncoder(
			conn,
		).Encode(
			Host{
				Hostname:      "kyronix-test-host",
				Uptime:        100 * time.Second,
				Architecture:  "arm64",
				OSID:          "debian",
				OSName:        "Debian GNU/Linux 13 (trixie)",
				OSVersion:     "13",
				KernelVersion: "6.6.25-v8+",
				Environment:   "production",
				Role:          "ntp-appliance",
			},
		)
	}()

	host, err := GetHost(
		socket,
	)

	if err != nil {
		t.Fatal(err)
	}

	if host.Hostname != "kyronix-test-host" {
		t.Fatalf(
			"unexpected hostname: %s",
			host.Hostname,
		)
	}

	if host.Environment != "production" {
		t.Fatalf(
			"unexpected environment: %s",
			host.Environment,
		)
	}

	if host.Role != "ntp-appliance" {
		t.Fatalf(
			"unexpected role: %s",
			host.Role,
		)
	}

	select {
	case command := <-done:

		if command != "host" {
			t.Fatalf(
				"expected host command, got %q",
				command,
			)
		}

	case <-time.After(
		time.Second,
	):

		t.Fatal(
			"host command was not received",
		)
	}
}
