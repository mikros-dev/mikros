package errors

import (
	"errors"

	errorsApi "github.com/mikros-dev/mikros/apis/errors"
	merrors "github.com/mikros-dev/mikros/internal/components/errors"
)

// IsInternalError checks if an error is a framework Internal error.
func IsInternalError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindInternal
	}

	return false
}

// IsNotFoundError checks if an error is a framework NotFound error.
func IsNotFoundError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindNotFound
	}

	return false
}

// IsInvalidArgumentError checks if an error is a framework InvalidArgument error.
func IsInvalidArgumentError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindValidation
	}

	return false
}

// IsPreconditionError checks if an error is a framework FailedPrecondition error.
func IsPreconditionError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindPrecondition
	}

	return false
}

// IsPermissionDeniedError checks if an error is a framework PermissionDenied error.
func IsPermissionDeniedError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindPermission
	}

	return false
}

// IsCustomError checks if an error is a framework Custom error.
func IsCustomError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindCustom
	}

	return false
}

// IsRPCError checks if an error is a framework RPC error.
func IsRPCError(err error) bool {
	if e, ok := IsKnownError(err); ok {
		return e.Kind == errorsApi.KindRPC
	}

	return false
}

func IsKnownError(err error) (*merrors.Error, bool) {
	var e *merrors.Error
	ok := errors.As(err, &e)
	return e, ok
}
