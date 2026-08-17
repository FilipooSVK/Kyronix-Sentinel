package local

import (
	"encoding/json"
	"net"
)

// GetStatus returns current Sentinel status.
func GetStatus(socket string) (Status, error) {

	conn, err := net.Dial(
		"unix",
		socket,
	)

	if err != nil {
		return Status{}, err
	}

	defer conn.Close()

	var status Status

	err = json.NewDecoder(conn).Decode(
		&status,
	)

	return status, err
}

// GetDiagnostics returns current Sentinel diagnostics.
func GetDiagnostics(socket string) (Diagnostics, error) {

	conn, err := net.Dial(
		"unix",
		socket,
	)

	if err != nil {
		return Diagnostics{}, err
	}

	defer conn.Close()

	_, err = conn.Write(
		[]byte("diagnose"),
	)

	if err != nil {
		return Diagnostics{}, err
	}

	var diagnostics Diagnostics

	err = json.NewDecoder(conn).Decode(
		&diagnostics,
	)

	return diagnostics, err
}

// GetPrediction returns current Sentinel prediction.
func GetPrediction(socket string) (Prediction, error) {

	conn, err := net.Dial(
		"unix",
		socket,
	)

	if err != nil {
		return Prediction{}, err
	}

	defer conn.Close()

	_, err = conn.Write(
		[]byte("prediction"),
	)

	if err != nil {
		return Prediction{}, err
	}

	var prediction Prediction

	err = json.NewDecoder(conn).Decode(
		&prediction,
	)

	return prediction, err
}
