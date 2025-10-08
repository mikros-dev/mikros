package errors

import (
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/service"
)

// ServiceError is a structure that holds internal error details to improve
// error log description for the end-user, and it implements the errorApi.Error
// interface.
type ServiceError struct {
	err        *Error
	attributes []logger_api.Attribute
	logger     func(ctx context.Context, msg string, attrs ...logger_api.Attribute)
}

type serviceErrorOptions struct {
	Code        int32
	Kind        Kind
	ServiceName string
	Message     string
	Destination string
	Logger      func(ctx context.Context, msg string, attrs ...logger_api.Attribute)
	Error       error
}

func newServiceError(options *serviceErrorOptions) *ServiceError {
	err := &Error{
		Code:        options.Code,
		ServiceName: options.ServiceName,
		Message:     options.Message,
		Destination: options.Destination,
		Kind:        options.Kind,
	}

	if options.Error != nil {
		err.SubLevelError = options.Error.Error()
	}

	return &ServiceError{
		err:    err,
		logger: options.Logger,
	}
}

// FromGRPCStatus converts a gRPC status object into a standardized service
// error format for better interoperability.
func FromGRPCStatus(st *status.Status, from, to service.Name) error {
	var (
		msg    = st.Message()
		retErr Error
	)

	if err := json.Unmarshal([]byte(msg), &retErr); err != nil {
		return newServiceError(&serviceErrorOptions{
			Destination: to.String(),
			Kind:        KindInternal,
			ServiceName: from.String(),
			Message:     "got an internal error",
			Error:       errors.New(msg),
		}).Submit(context.TODO())
	}

	// If we're dealing with a non-mikros error, change it to an Internal
	// one so services can properly handle them.
	if st.Code() != codes.Unknown {
		retErr.Kind = KindInternal
		retErr.SubLevelError = msg
	}

	return &retErr
}

// WithCode attaches a numeric error code to the ServiceError.
func (s *ServiceError) WithCode(code errors_api.Code) errors_api.Error {
	s.err.Code = code.ErrorCode()
	return s
}

// WithAttributes adds custom log attributes to the ServiceError, augmenting
// the error context for detailed logging.
func (s *ServiceError) WithAttributes(attrs ...logger_api.Attribute) errors_api.Error {
	s.attributes = attrs
	return s
}

// Submit logs the error details using the configured logger and returns the
// underlying error for further handling.
func (s *ServiceError) Submit(ctx context.Context) error {
	// Display the error message onto the output
	if s.logger != nil {
		logFields := []logger_api.Attribute{withKind(s.err.Kind)}
		if s.err.SubLevelError != "" {
			logFields = append(logFields, logger.String("error.message", s.err.SubLevelError))
		}

		s.logger(ctx, s.err.Message, append(logFields, s.attributes...)...)
	}

	// And give back the proper error for the API
	return s.err
}

// Kind returns the string representation of the error's kind.
func (s *ServiceError) Kind() string {
	return s.err.Kind.String()
}

// withKind wraps a Kind into a structured log Attribute.
func withKind(kind Kind) logger_api.Attribute {
	return logger.String("error.kind", string(kind))
}

// Error is the framework error type that a service handler should return to
// keep a standard error between services.
type Error struct {
	Code          int32  `json:"code"`
	ServiceName   string `json:"service_name,omitempty"`
	Message       string `json:"message,omitempty"`
	Destination   string `json:"destination,omitempty"`
	Kind          Kind   `json:"kind"`
	SubLevelError string `json:"details,omitempty"`
}

func (e *Error) Error() string {
	return e.String()
}

func (e *Error) String() string {
	b, _ := json.Marshal(e)
	return string(b)
}
