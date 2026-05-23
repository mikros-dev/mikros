package errors

import (
	"errors"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

// View is the public read-only representation of a framework error.
//
// These methods are available to inspect errors returned by service handlers
// through the errors feature.
type View interface {
	error

	// Code returns the error code, which was previously adjusted with
	// Value.WithCode.
	Code() int32

	// Message returns the error message.
	Message() string

	// Cause returns the underlying error that caused this error.
	Cause() error

	// Kind returns the error kind.
	Kind() Kind

	// Attributes returns the error attributes, which were previously added with
	// Value.WithAttributes.
	Attributes() []logger_api.Attribute
}

// From returns the framework error if the given error wraps one.
func From(err error) (View, bool) {
	var e View
	ok := errors.As(err, &e)
	return e, ok
}

// IsInternal checks if an error is a framework Internal error.
func IsInternal(err error) bool {
	return IsKind(err, KindInternal)
}

// IsNotFound checks if an error is a framework NotFound error.
func IsNotFound(err error) bool {
	return IsKind(err, KindNotFound)
}

// IsInvalidArgument checks if an error is a framework InvalidArgument error.
func IsInvalidArgument(err error) bool {
	return IsKind(err, KindInvalidArgument)
}

// IsFailedPrecondition checks if an error is a framework FailedPrecondition error.
func IsFailedPrecondition(err error) bool {
	return IsKind(err, KindPrecondition)
}

// IsPermissionDenied checks if an error is a framework PermissionDenied error.
func IsPermissionDenied(err error) bool {
	return IsKind(err, KindPermission)
}

// IsRPC checks if an error is a framework RPC error.
func IsRPC(err error) bool {
	return IsKind(err, KindRPC)
}

// IsKind reports whether the given error is a known framework error.
func IsKind(err error, kind Kind) bool {
	e, ok := From(err)
	return ok && e.Kind() == kind
}
