package errors

import (
	"context"

	"github.com/mikros-dev/mikros/apis/features/logger"
)

// ErrorAPI is an API that a service must use to wrap its errors into a proper
// framework error.
type ErrorAPI interface {
	// RPC should be used when an error was received from an RPC call.
	RPC(err error, destination string) Error

	// InvalidArgument should be used when invalid arguments were received
	// inside a service handler.
	InvalidArgument(err error) Error

	// FailedPrecondition should be used when a specific condition is not met.
	FailedPrecondition(message string) Error

	// NotFound should be used when a resource was not found.
	NotFound() Error

	// Internal should be used when an unexpected behavior (or error) occurred
	// internally.
	Internal(err error) Error

	// PermissionDenied should be used when a client does not have access to
	// a specific resource.
	PermissionDenied() Error

	// Custom should be used by a service when none of the previous APIs are
	// able to handle the error that occurred. It will be forward as an internal
	// error.
	Custom(msg string) Error
}

// Error is the proper error that a service can return by its handlers. When
// submitted, it writes a log message describing what happened, and it gives
// the error using the language error type.
type Error interface {
	// WithCode sets a custom error code to be added inside the error.
	WithCode(code Code) Error

	// WithAttributes adds a set o custom log attributes to be inserted into
	// the log message.
	WithAttributes(attrs ...logger.Attribute) Error

	// Submit wraps the service error into a proper error type allowing the
	// service to return it.
	Submit(ctx context.Context) error

	// Kind returns the kind of error represented by the error.
	Kind() string
}

// Code is an interface that allows a service embed an integer value into an
// error.
type Code interface {
	ErrorCode() int32
}
