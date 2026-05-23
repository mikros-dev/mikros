package errors

import (
	"fmt"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	merrors "github.com/mikros-dev/mikros/components/errors"
)

// Ensure that the Builder implements the Errors interface.
var _ errors_api.Errors = &Builder{}

// Builder provides error creation utilities.
type Builder struct {
	serviceName string
}

// BuilderOptions represents configuration options for creating an error builder.
type BuilderOptions struct {
	// ServiceName specifies the name of the service associated with the builder.
	ServiceName string
}

// NewBuilder creates a new Builder object.
func NewBuilder(options BuilderOptions) *Builder {
	return &Builder{
		serviceName: options.ServiceName,
	}
}

// RPC sets that the current error is related to an RPC call with another gRPC
// service (destination).
func (b *Builder) RPC(err error, destination string) errors_api.Value {
	return &value{
		kind:        merrors.KindRPC,
		serviceName: b.serviceName,
		message:     "service RPC error",
		destination: destination,
		cause:       err,
	}
}

// InvalidArgument sets that the current error is related to an argument that
// didn't follow validation rules.
func (b *Builder) InvalidArgument(err error) errors_api.Value {
	return &value{
		kind:        merrors.KindInvalidArgument,
		serviceName: b.serviceName,
		message:     "request validation failed",
		cause:       err,
	}
}

// FailedPrecondition sets that the current error is related to an internal
// condition which wasn't satisfied.
func (b *Builder) FailedPrecondition(message string) errors_api.Value {
	return &value{
		kind:        merrors.KindPrecondition,
		serviceName: b.serviceName,
		message:     message,
	}
}

// NotFound sets that the current error is related to some data not being found,
// probably in the database.
func (b *Builder) NotFound() errors_api.Value {
	return &value{
		kind:        merrors.KindNotFound,
		serviceName: b.serviceName,
		message:     "not found",
	}
}

// Internal sets that the current error is related to an internal service
// error.
func (b *Builder) Internal(err error) errors_api.Value {
	return &value{
		kind:        merrors.KindInternal,
		serviceName: b.serviceName,
		message:     "got an internal error",
		cause:       err,
	}
}

// PermissionDenied sets that the current error is related to a client trying
// to access a resource without having permission to do so.
func (b *Builder) PermissionDenied() errors_api.Value {
	return &value{
		kind:        merrors.KindPermission,
		serviceName: b.serviceName,
		message:     fmt.Sprintf("no permission to access %s", b.serviceName),
	}
}
