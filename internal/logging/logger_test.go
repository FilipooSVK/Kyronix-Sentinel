package logging

import (
	"bytes"
	"strings"
	"testing"
)

func TestLoggerWritesJSON(t *testing.T) {

	var buffer bytes.Buffer

	logger := New(
		&buffer,
		"info",
	)

	logger.Info(
		"test message",
		map[string]interface{}{
			"value": 100,
		},
	)

	result := buffer.String()

	if !strings.Contains(
		result,
		"test message",
	) {

		t.Fatal(
			"message missing",
		)
	}
}
