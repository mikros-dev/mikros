package errors

import (
	"context"
	"encoding/json"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	ferrors "github.com/mikros-dev/mikros/apis/features/errors"
	flogger "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/service"
)

// ServiceError is a structure that holds internal error details to improve
// error log description for the end-user, and it implements the errorApi.Error
// interface.
type ServiceError struct {
	err        *Error
	attributes []flogger.Attribute
	logger     func(ctx context.Context, msg string, attrs ...flogger.Attribute)
}

type serviceErrorOptions struct {
	Code        int32
	Kind        Kind
	ServiceName string
	Message     string
	Destination string
	Logger      func(ctx context.Context, msg string, attrs ...flogger.Attribute)
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

	// If we're dealing with a non mikros error, change it to an Internal
	// one so services can properly handle them.
	if st.Code() != codes.Unknown {
		retErr.Kind = KindInternal
		retErr.SubLevelError = msg
	}

	return &retErr
}

func (s *ServiceError) WithCode(code ferrors.Code) ferrors.Error {
	s.err.Code = code.ErrorCode()
	return s
}

func (s *ServiceError) WithAttributes(attrs ...flogger.Attribute) ferrors.Error {
	s.attributes = attrs
	return s
}

func (s *ServiceError) Submit(ctx context.Context) error {
	// Display the error message onto the output
	if s.logger != nil {
		logFields := []flogger.Attribute{withKind(s.err.Kind)}
		if s.err.SubLevelError != "" {
			logFields = append(logFields, logger.String("error.message", s.err.SubLevelError))
		}

		s.logger(ctx, s.err.Message, append(logFields, s.attributes...)...)
	}

	// And give back the proper error for the API
	return s.err
}

func (s *ServiceError) Kind() string {
	return s.err.Kind.String()
}

// withKind wraps a Kind into a structured log Attribute.
func withKind(kind Kind) flogger.Attribute {
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
