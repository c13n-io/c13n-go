package slog

import (
	"github.com/ThreeDotsLabs/watermill"
	log "github.com/sirupsen/logrus"
)

// WatermillLogger is the logger used for watermill.
type WatermillLogger struct {
	*Logger
}

// NewWLogger initializes a new WatermillLogger.
func NewWLogger(component string) *WatermillLogger {
	embeddedLogger := NewLogger(component)

	return &WatermillLogger{embeddedLogger}
}

// These methods are needed so that `WatermillLogger` satisfies the `LoggerAdapter` interface,
// which is necessary if it is to be provided as the logger to `watermill`.
var _ watermill.LoggerAdapter = new(WatermillLogger)

func (l *WatermillLogger) Error(msg string, err error, fields watermill.LogFields) {
	logFields := watermillFieldsToLogrus(fields)

	l.WithFields(logFields).WithError(err).Error(msg)
}

// Info creates a new info log entry.
func (l *WatermillLogger) Info(msg string, fields watermill.LogFields) {
	logFields := watermillFieldsToLogrus(fields)

	l.WithFields(logFields).Info(msg)
}

// Debug creates a new debug log entry.
func (l *WatermillLogger) Debug(msg string, fields watermill.LogFields) {
	logFields := watermillFieldsToLogrus(fields)

	l.WithFields(logFields).Debug(msg)
}

// Trace creates a new trace log entry.
func (l *WatermillLogger) Trace(msg string, fields watermill.LogFields) {
	logFields := watermillFieldsToLogrus(fields)

	l.WithFields(logFields).Trace(msg)
}

// With adds a map of fields to the log entry.
func (l *WatermillLogger) With(fields watermill.LogFields) watermill.LoggerAdapter {
	logFields := watermillFieldsToLogrus(fields)

	return &WatermillLogger{
		l.WithFields(logFields),
	}
}

func watermillFieldsToLogrus(fields watermill.LogFields) log.Fields {
	logFields := make(log.Fields, len(fields))
	for k, v := range fields {
		logFields[k] = v
	}

	return logFields
}
