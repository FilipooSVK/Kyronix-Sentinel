package local

import (
	"encoding/json"
	"net"
	"time"
)

const clientTimeout = 5 * time.Second

// GetStatus returns current Sentinel status.
func GetStatus(
	socket string,
) (Status, error) {

	var status Status

	err := request(
		socket,
		"status",
		&status,
	)

	return status, err
}

// GetDiagnostics returns current Sentinel diagnostics.
func GetDiagnostics(
	socket string,
) (Diagnostics, error) {

	var diagnostics Diagnostics

	err := request(
		socket,
		"diagnose",
		&diagnostics,
	)

	return diagnostics, err
}

// GetPrediction returns current Sentinel prediction.
func GetPrediction(
	socket string,
) (Prediction, error) {

	var prediction Prediction

	err := request(
		socket,
		"prediction",
		&prediction,
	)

	return prediction, err
}

func request(
	socket string,
	command string,
	response interface{},
) error {

	conn, err := net.DialTimeout(
		"unix",
		socket,
		clientTimeout,
	)

	if err != nil {
		return err
	}

	defer conn.Close()

	if err := conn.SetDeadline(
		time.Now().Add(
			clientTimeout,
		),
	); err != nil {

		return err
	}

	if _, err := conn.Write(
		[]byte(command),
	); err != nil {

		return err
	}

	return json.NewDecoder(
		conn,
	).Decode(
		response,
	)
}
