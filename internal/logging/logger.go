package logging

import (
	"encoding/json"
	"io"
	"log"
	"time"
)

// Logger provides Sentinel logging.
type Logger struct {
	logger *log.Logger
	level  string
}

// Entry represents structured log event.
type Entry struct {
	Time time.Time `json:"time"`

	Level string `json:"level"`

	Message string `json:"message"`

	Fields map[string]interface{} `json:"fields,omitempty"`
}

// New creates structured logger.
func New(
	writer io.Writer,
	level string,
) *Logger {

	return &Logger{
		logger: log.New(
			writer,
			"",
			0,
		),
		level: level,
	}
}

// Info writes informational event.
func (l *Logger) Info(
	message string,
	fields map[string]interface{},
) {

	l.write(
		"info",
		message,
		fields,
	)
}

// Error writes error event.
func (l *Logger) Error(
	message string,
	fields map[string]interface{},
) {

	l.write(
		"error",
		message,
		fields,
	)
}

func (l *Logger) write(
	level string,
	message string,
	fields map[string]interface{},
) {

	entry := Entry{

		Time: time.Now().UTC(),

		Level: level,

		Message: message,

		Fields: fields,
	}

	data, _ := json.Marshal(
		entry,
	)

	l.logger.Println(
		string(data),
	)
}
