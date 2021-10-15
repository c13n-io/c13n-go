package slog

import (
	"io"
	"io/ioutil"
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

// Logger is a type alias for logrus.Entry.
type Logger = log.Entry

// Package variables for output control.
var (
	DefaultLogOutput io.Writer = os.Stderr
	LogLevel         log.Level = log.DebugLevel

	// The base for all loggers instantiated from this package.
	baseLogger *log.Logger

	// fieldOrder is the default field order for logged fields.
	fieldOrder = []string{
		"component",
		"system", "span.kind",
		"grpc.start_time", "grpc.time_ms",
		"grpc.service", "grpc.method",
		"peer.address",
		"grpc.request.content", "grpc.response.content",
		"grpc.code", "error",
	}
)

func init() {
	baseLogger = &log.Logger{
		Out:   DefaultLogOutput,
		Level: LogLevel,
		Formatter: &nested.Formatter{
			FieldsOrder:     fieldOrder,
			ShowFullLevel:   true,
			NoColors:        false,
			TimestampFormat: "2006-01-02 15:04:05.000000",
		},
	}
}

// SetLogOutput changes the logging output to the specified writer,
// and returns the old one.
func SetLogOutput(out io.Writer) io.Writer {
	oldLogOutput := DefaultLogOutput

	DefaultLogOutput = out
	baseLogger.SetOutput(out)

	return oldLogOutput
}

// SetLogLevel changes the logging level to correspond with the provided argument.
// Accepted argument values are
// "panic", "fatal", "error", "warn", "info", "debug", "trace".
func SetLogLevel(level string) error {
	parsedLevel, err := log.ParseLevel(level)
	if err != nil {
		return err
	}

	LogLevel = parsedLevel
	baseLogger.SetLevel(LogLevel)

	return nil
}

// Disable disables the logging output and returns the previous writer.
func Disable() io.Writer {
	return SetLogOutput(ioutil.Discard)
}

// NewLogger returns a new logger for a component.
func NewLogger(component string) *Logger {
	return baseLogger.WithField("component", component)
}
