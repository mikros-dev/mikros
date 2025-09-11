package errors

import (
	"fmt"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

type Factory struct {
	serviceName string
	logger      logger_api.LoggerAPI
}

type FactoryOptions struct {
	ServiceName string
	Logger      logger_api.LoggerAPI
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
func (f *Factory) RPC(err error, destination string) errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindRPC,
		ServiceName: f.serviceName,
		Message:     "service RPC error",
		Destination: destination,
		Error:       err,
	}
	if f.logger != nil {
		options.Logger = f.logger.Warn
	}

	return newServiceError(options)
}

// InvalidArgument sets that the current error is related to an argument that
// didn't follow validation rules.
func (f *Factory) InvalidArgument(err error) errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindValidation,
		ServiceName: f.serviceName,
		Message:     "request validation failed",
		Error:       err,
	}
	if f.logger != nil {
		options.Logger = f.logger.Warn
	}

	return newServiceError(options)
}

// FailedPrecondition sets that the current error is related to an internal
// condition which wasn't satisfied.
func (f *Factory) FailedPrecondition(message string) errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindPrecondition,
		ServiceName: f.serviceName,
		Message:     message,
	}
	if f.logger != nil {
		options.Logger = f.logger.Warn
	}

	return newServiceError(options)
}

// NotFound sets that the current error is related to some data not being found,
// probably in the database.
func (f *Factory) NotFound() errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindNotFound,
		ServiceName: f.serviceName,
		Message:     "not found",
	}
	if f.logger != nil {
		options.Logger = f.logger.Warn
	}

	return newServiceError(options)
}

// Internal sets that the current error is related to an internal service
// error.
func (f *Factory) Internal(err error) errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindInternal,
		ServiceName: f.serviceName,
		Message:     "got an internal error",
		Error:       err,
	}
	if f.logger != nil {
		options.Logger = f.logger.Error
	}

	return newServiceError(options)
}

// PermissionDenied sets that the current error is related to a client trying
// to access a resource without having permission to do so.
func (f *Factory) PermissionDenied() errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindPermission,
		ServiceName: f.serviceName,
		Message:     fmt.Sprintf("no permission to access %s", f.serviceName),
	}
	if f.logger != nil {
		options.Logger = f.logger.Info
	}

	return newServiceError(options)
}

// Custom lets a service set a custom error kind for its errors. Internally, it
// will be treated as an Internal error.
func (f *Factory) Custom(msg string) errors_api.Error {
	options := &serviceErrorOptions{
		Kind:        KindCustom,
		ServiceName: f.serviceName,
		Message:     msg,
	}
	if f.logger != nil {
		options.Logger = f.logger.Info
	}

	return newServiceError(options)
}
