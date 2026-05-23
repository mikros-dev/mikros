package errors

import (
	"encoding/json"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	merrors "github.com/mikros-dev/mikros/components/errors"
	"github.com/mikros-dev/mikros/components/service"
)

// Ensure that the value implements the value interface
var _ merrors.View = &value{}

// value is the framework error type that a service handler should return to
// keep a standard error between services.
type value struct {
	code        int32
	serviceName string
	message     string
	destination string
	kind        merrors.Kind
	cause       error
	attributes  []logger_api.Attribute
}

func (v *value) Code() int32 {
	return v.code
}

func (v *value) Error() string {
	return v.String()
}

func (v *value) String() string {
	b, _ := json.Marshal(v.grpcMessage())
	return string(b)
}

func (v *value) WithCode(code errors_api.Code) errors_api.Value {
	v.code = code.ErrorCode()
	return v
}

func (v *value) WithAttributes(attrs ...logger_api.Attribute) errors_api.Value {
	v.attributes = append(v.attributes, attrs...)
	return v
}

func (v *value) Message() string {
	return v.message
}

func (v *value) Attributes() []logger_api.Attribute {
	return append([]logger_api.Attribute(nil), v.attributes...)
}

func (v *value) Cause() error {
	return v.cause
}

func (v *value) Kind() merrors.Kind {
	return v.kind
}

func (v *value) Unwrap() error {
	return v.cause
}

type grpcErrorMessage struct {
	Kind        merrors.Kind `json:"kind"`
	Message     string       `json:"message,omitempty"`
	Cause       string       `json:"cause,omitempty"`
	Code        int32        `json:"code,omitempty"`
	ServiceName string       `json:"service_name,omitempty"`
	Destination string       `json:"destination,omitempty"`
}

func (v *value) grpcMessage() grpcErrorMessage {
	msg := grpcErrorMessage{
		Kind:        v.kind,
		Message:     v.message,
		Code:        v.code,
		ServiceName: v.serviceName,
		Destination: v.destination,
	}
	if v.cause != nil {
		msg.Cause = v.cause.Error()
	}

	return msg
}

func ToGRPCStatus(err error) (*status.Status, bool, error) {
	var v *value
	if !errors.As(err, &v) {
		return nil, false, nil
	}

	b, err := json.Marshal(v.grpcMessage())
	if err != nil {
		return nil, false, err
	}

	return status.New(grpcCode(v.kind), string(b)), true, nil
}

func grpcCode(kind merrors.Kind) codes.Code {
	switch kind {
	case merrors.KindInternal:
		return codes.Internal
	case merrors.KindNotFound:
		return codes.NotFound
	case merrors.KindInvalidArgument:
		return codes.InvalidArgument
	case merrors.KindPrecondition:
		return codes.FailedPrecondition
	case merrors.KindPermission:
		return codes.PermissionDenied
	case merrors.KindRPC:
		return codes.Unavailable
	default:
		return codes.Unknown
	}
}

// FromGRPCStatus converts a gRPC status object into a standardized service
// error format for better interoperability.
func FromGRPCStatus(st *status.Status, from, to service.Name) error {
	if st == nil {
		return internalRemoteError(from, to, "nil gRPC status")
	}

	var msg grpcErrorMessage
	if err := json.Unmarshal([]byte(st.Message()), &msg); err != nil {
		return internalRemoteError(from, to, st.Message())
	}

	var cause error
	if msg.Cause != "" {
		cause = errors.New(msg.Cause)
	}

	return &value{
		code:        msg.Code,
		serviceName: from.String(),
		destination: to.String(),
		message:     msg.Message,
		kind:        msg.Kind,
		cause:       cause,
	}
}

func internalRemoteError(from, to service.Name, msg string) error {
	return &value{
		serviceName: from.String(),
		destination: to.String(),
		kind:        merrors.KindInternal,
		cause:       errors.New(msg),
		message:     "got an internal error",
	}
}
