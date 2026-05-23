package errors

import (
	"github.com/mikros-dev/mikros/apis/features/logger"
)

// Errors provides a structured way for services to create and handle errors.
//
// This interface is implemented by the mikros framework and made available to
// services that opt into the feature. It enables classification of service
// errors into standard types (e.g., internal, not found, invalid argument)
// for consistent error reporting.
type Errors interface {
	// RPC should be used when an error is received from an RPC call to
	// another service. The destination identifies the remote service.
	RPC(err error, destination string) Value

	// InvalidArgument should be used when a handler receives invalid input
	// parameters.
	InvalidArgument(err error) Value

	// FailedPrecondition should be used when a required condition is not met
	// for an operation to proceed.
	FailedPrecondition(message string) Value

	// NotFound should be used when a requested resource could not be located.
	NotFound() Value

	// Internal should be used when an unexpected internal behavior or failure
	// occurs in the service.
	Internal(err error) Value

	// PermissionDenied should be used when a client is not authorized to
	// access the requested resource.
	PermissionDenied() Value
}

// Value represents a structured service error returned by handlers.
//
// The mikros framework provides this interface to allow services to enrich
// errors with metadata such as error codes and log attributes.
type Value interface {
	error

	// WithCode attaches a custom numeric code to the error.
	WithCode(code Code) Value

	// WithAttributes adds custom log attributes to be included in the log
	// entry generated for this error.
	WithAttributes(attrs ...logger.Attribute) Value
}

// Code allows embedding a numeric error code into a service error.
//
// This can be used to define domain-specific codes for client interpretation
// or structured logging.
type Code interface {
	// ErrorCode returns the numeric code associated with the error.
	ErrorCode() int32
}
