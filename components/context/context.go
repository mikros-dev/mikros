package context

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"google.golang.org/grpc/metadata"

	"github.com/mikros-dev/mikros/components/service"
)

const (
	contextKeyName = "service-context-"
)

// ServiceContext is an object stored inside a service RPC/HTTP handler
// context.Context to provide information at all source levels, such as, the
// logger.
type ServiceContext struct {
	values map[string]string
}

// Options encapsulates the required configuration for initializing a service
// context.
type Options struct {
	Name service.Name `validate:"required"`
}

// New initializes a new ServiceContext.
func New(options *Options) (*ServiceContext, error) {
	validate := validator.New()
	if err := validate.Struct(options); err != nil {
		return nil, err
	}

	serviceContext := newServiceContext()

	// Adds constant values into the service context.
	serviceContext.Add("caller", options.Name.String())

	return serviceContext, nil
}

// newServiceContext creates a new ServiceContext object.
func newServiceContext() *ServiceContext {
	return &ServiceContext{
		values: make(map[string]string),
	}
}

// AppendServiceContext adds all ServiceContext key-values inside the current
// context.
func AppendServiceContext(ctx context.Context, svcCtx *ServiceContext) context.Context {
	if svcCtx == nil {
		return ctx
	}

	var mdValues []string

	// Stores all ServiceContext values inside the current context appending
	// a custom string prefix to identify them later.
	for k, v := range svcCtx.values {
		mdValues = append(mdValues, fmt.Sprintf("%s%s", contextKeyName, k), v)
	}

	return metadata.AppendToOutgoingContext(ctx, mdValues...)
}

// AppendValue adds a new key-value pair inside the current context.
func AppendValue(ctx context.Context, key, value string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, fmt.Sprintf("%s%s", contextKeyName, key), value)
}

// FromContext retrieves a ServiceContext from the current context.
func FromContext(ctx context.Context) (*ServiceContext, bool) {
	// Notice that we are reading the IncomingContext here, because we want to
	// retrieve the ServiceContext that someone is sending to us.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		svcCtx := newServiceContext()

		for k, v := range md {
			if strings.HasPrefix(k, contextKeyName) {
				svcCtx.Add(strings.TrimPrefix(k, contextKeyName), v[0])
			}
		}

		return svcCtx, true
	}

	return nil, false
}

// Add adds a new key-value value pair inside the ServiceContext object.
func (s *ServiceContext) Add(key, value string) {
	s.values[key] = value
}

// Get retrieves a value stored inside the current ServiceContext.
func (s *ServiceContext) Get(key string) (string, bool) {
	v, ok := s.values[key]
	return v, ok
}

// Values gives a copy of the internal key-value ServiceContext container.
func (s *ServiceContext) Values() map[string]string {
	values := make(map[string]string)
	for k, v := range s.values {
		values[k] = v
	}

	return values
}
