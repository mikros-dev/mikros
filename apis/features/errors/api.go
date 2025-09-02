package errors

import (
	"context"

	"github.com/mikros-dev/mikros/apis/features/logger"
)

// ErrorAPI provides a structured way for services to create and handle errors.
//
// This interface is implemented by the mikros framework and made available to
// services that opt into the feature. It enables classification of service
// errors into standard types (e.g., internal, not found, invalid argument)
// for consistent error reporting and logging.
type ErrorAPI interface {
	// RPC should be used when an error is received from an RPC call to
	// another service. The destination identifies the remote service.
	RPC(err error, destination string) Error

	// InvalidArgument should be used when a handler receives invalid input
	// parameters.
	InvalidArgument(err error) Error

	// FailedPrecondition should be used when a required condition is not met
	// for an operation to proceed.
	FailedPrecondition(message string) Error

	// NotFound should be used when a requested resource could not be located.
	NotFound() Error

	// Internal should be used when an unexpected internal behavior or failure
	// occurs in the service.
	Internal(err error) Error

	// PermissionDenied should be used when a client is not authorized to
	// access the requested resource.
	PermissionDenied() Error

	// Custom should be used for error cases that do not match any of the
	// predefined types. These are treated as internal errors by default.
	Custom(msg string) Error
}

// Error represents a structured service error returned by handlers.
//
// The mikros framework provides this interface to allow services to enrich
// errors with metadata such as error codes and log attributes. When submitted,
// the error is logged and returned in a format suitable for clients.
type Error interface {
	// WithCode attaches a custom numeric code to the error.
	WithCode(code Code) Error

	// WithAttributes adds custom log attributes to be included in the log
	// entry generated for this error.
	WithAttributes(attrs ...logger.Attribute) Error

	// Submit finalizes the error, logs it, and converts it into a standard
	// Go error for return from a handler.
	Submit(ctx context.Context) error

	// Kind returns the classification of the error (e.g., "not_found",
	// "internal", "invalid_argument").
	Kind() string
}

// Code allows embedding a numeric error code into a service error.
//
// This can be used to define domain-specific codes for client interpretation
// or structured logging.
type Code interface {
	// ErrorCode returns the numeric code associated with the error.
	ErrorCode() int32
}
