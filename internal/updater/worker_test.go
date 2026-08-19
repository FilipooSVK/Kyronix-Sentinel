package updater

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestSystemdUpdateWorkerStart(
	t *testing.T,
) {

	worker := NewSystemdUpdateWorker(
		"sentinel-update.service",
	)

	var gotCommand string
	var gotArgs []string

	worker.run = func(
		ctx context.Context,
		command string,
		args ...string,
	) ([]byte, error) {

		gotCommand = command

		gotArgs = append(
			[]string(nil),
			args...,
		)

		return nil, nil
	}

	if err := worker.Start(
		context.Background(),
	); err != nil {

		t.Fatal(err)
	}

	if gotCommand != "systemctl" {

		t.Fatalf(
			"unexpected command: %q",
			gotCommand,
		)
	}

	wantArgs := []string{
		"start",
		"--no-block",
		"sentinel-update.service",
	}

	if !reflect.DeepEqual(
		gotArgs,
		wantArgs,
	) {

		t.Fatalf(
			"unexpected arguments: got %#v want %#v",
			gotArgs,
			wantArgs,
		)
	}
}

func TestSystemdUpdateWorkerTrimsServiceName(
	t *testing.T,
) {

	worker := NewSystemdUpdateWorker(
		"  sentinel-update.service  ",
	)

	if worker.Service !=
		"sentinel-update.service" {

		t.Fatalf(
			"unexpected service name: %q",
			worker.Service,
		)
	}
}

func TestSystemdUpdateWorkerRejectsEmptyService(
	t *testing.T,
) {

	worker := NewSystemdUpdateWorker(
		"",
	)

	err := worker.Start(
		context.Background(),
	)

	if err == nil {

		t.Fatal(
			"expected empty service error",
		)
	}
}

func TestSystemdUpdateWorkerRejectsNilContext(
	t *testing.T,
) {

	worker := NewSystemdUpdateWorker(
		"sentinel-update.service",
	)

	err := worker.Start(
		nil,
	)

	if err == nil {

		t.Fatal(
			"expected nil context error",
		)
	}
}

func TestSystemdUpdateWorkerRejectsNilRunner(
	t *testing.T,
) {

	worker := NewSystemdUpdateWorker(
		"sentinel-update.service",
	)

	worker.run = nil

	err := worker.Start(
		context.Background(),
	)

	if err == nil {

		t.Fatal(
			"expected nil runner error",
		)
	}
}

func TestSystemdUpdateWorkerReportsCommandFailure(
	t *testing.T,
) {

	worker := NewSystemdUpdateWorker(
		"sentinel-update.service",
	)

	expectedErr :=
		errors.New(
			"command failed",
		)

	worker.run = func(
		ctx context.Context,
		command string,
		args ...string,
	) ([]byte, error) {

		return []byte(
				"Failed to start unit",
			),
			expectedErr
	}

	err := worker.Start(
		context.Background(),
	)

	if err == nil {

		t.Fatal(
			"expected command failure",
		)
	}

	if !errors.Is(
		err,
		expectedErr,
	) {

		t.Fatalf(
			"expected wrapped command error, got %v",
			err,
		)
	}

	if !strings.Contains(
		err.Error(),
		"Failed to start unit",
	) {

		t.Fatalf(
			"expected systemctl output in error: %v",
			err,
		)
	}
}
