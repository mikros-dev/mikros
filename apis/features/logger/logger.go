package logger

import (
	"context"
)

// LoggerAPI is the log interface that is available for all services to show
// messages using different levels.
type LoggerAPI interface {
	// Debug outputs messages using debug level.
	Debug(ctx context.Context, msg string, attrs ...Attribute)

	// Internal outputs messages using the internal level.
	Internal(ctx context.Context, msg string, attrs ...Attribute)

	// Info outputs messages using the info level.
	Info(ctx context.Context, msg string, attrs ...Attribute)

	// Warn outputs messages using warning level.
	Warn(ctx context.Context, msg string, attrs ...Attribute)

	// Error outputs messages using error level.
	Error(ctx context.Context, msg string, attrs ...Attribute)

	// Fatal outputs message using fatal level.
	Fatal(ctx context.Context, msg string, attrs ...Attribute)

	// SetLogLevel changes the current messages log level.
	SetLogLevel(level string) (string, error)

	// Level gets the current log level.
	Level() string
}

// Attribute is an interface that a property that can be written into a log
// message must have in order to do so.
type Attribute interface {
	Key() string
	Value() interface{}
}
