package errors

import (
	"fmt"

	ferrors "github.com/mikros-dev/mikros/apis/features/errors"
	flogger "github.com/mikros-dev/mikros/apis/features/logger"
)

type Factory struct {
	serviceName string
	logger      flogger.LoggerAPI
}

type FactoryOptions struct {
	ServiceName string
	Logger      flogger.LoggerAPI
}

// NewFactory creates a new Factory object.
func NewFactory(options FactoryOptions) *Factory {
	return &Factory{
		serviceName: options.ServiceName,
		logger:      options.Logger,
	}
}

// RPC sets that the current error is related to an RPC call with another gRPC
// service (destination).
func (f *Factory) RPC(err error, destination string) ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindRPC,
		ServiceName: f.serviceName,
		Message:     "service RPC error",
		Destination: destination,
		Logger:      f.logger.Warn,
		Error:       err,
	})
}

// InvalidArgument sets that the current error is related to an argument that
// didn't follow validation rules.
func (f *Factory) InvalidArgument(err error) ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindValidation,
		ServiceName: f.serviceName,
		Message:     "request validation failed",
		Logger:      f.logger.Warn,
		Error:       err,
	})
}

// FailedPrecondition sets that the current error is related to an internal
// condition which wasn't satisfied.
func (f *Factory) FailedPrecondition(message string) ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindPrecondition,
		ServiceName: f.serviceName,
		Message:     message,
		Logger:      f.logger.Warn,
	})
}

// NotFound sets that the current error is related to some data not being found,
// probably in the database.
func (f *Factory) NotFound() ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindNotFound,
		ServiceName: f.serviceName,
		Message:     "not found",
		Logger:      f.logger.Warn,
	})
}

// Internal sets that the current error is related to an internal service
// error.
func (f *Factory) Internal(err error) ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindInternal,
		ServiceName: f.serviceName,
		Message:     "got an internal error",
		Logger:      f.logger.Error,
		Error:       err,
	})
}

// PermissionDenied sets that the current error is related to a client trying
// to access a resource without having permission to do so.
func (f *Factory) PermissionDenied() ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindPermission,
		ServiceName: f.serviceName,
		Message:     fmt.Sprintf("no permission to access %s", f.serviceName),
		Logger:      f.logger.Info,
	})
}

// Custom lets a service set a custom error kind for its errors. Internally, it
// will be treated as an Internal error.
func (f *Factory) Custom(msg string) ferrors.Error {
	return newServiceError(&serviceErrorOptions{
		Kind:        KindCustom,
		ServiceName: f.serviceName,
		Message:     msg,
		Logger:      f.logger.Info,
	})
}
