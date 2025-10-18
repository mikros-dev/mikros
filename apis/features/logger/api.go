package logger

import (
	"context"
)

// API provides a structured logging interface with multiple log levels.
//
// This interface is implemented by the mikros framework and is available to all
// services by default. It allows emitting log messages with contextual metadata
// using attributes and supports runtime log level configuration.
type API interface {
	// Debug logs a message at the debug level with optional attributes.
	Debug(ctx context.Context, msg string, attrs ...Attribute)

	// Internal logs a message at the internal level with optional attributes.
	Internal(ctx context.Context, msg string, attrs ...Attribute)

	// Info logs a message at the info level with optional attributes.
	Info(ctx context.Context, msg string, attrs ...Attribute)

	// Warn logs a message at the warning level with optional attributes.
	Warn(ctx context.Context, msg string, attrs ...Attribute)

	// Error logs a message at the error level with optional attributes.
	Error(ctx context.Context, msg string, attrs ...Attribute)

	// Fatal logs a message at the fatal level with optional attributes.
	// It causes the process to terminate.
	Fatal(ctx context.Context, msg string, attrs ...Attribute)

	// SetLogLevel changes the current log level to the specified value.
	// Returns the previous level or an error if the input is invalid.
	SetLogLevel(level string) (string, error)

	// Level returns the current log level as a string.
	Level() string
}

// Attribute represents a key-value pair attached to log messages.
//
// Attributes provide additional context to log entries and are used
// to structure logs for better querying and filtering.
type Attribute interface {
	// Key returns the attribute's name.
	Key() string

	// Value returns the attribute's value.
	Value() interface{}
}
