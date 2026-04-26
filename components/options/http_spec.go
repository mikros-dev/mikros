package options

import (
	"github.com/mikros-dev/mikros/apis/runtimes/http_spec"
	"github.com/mikros-dev/mikros/components/definition"
)

// HTTPSpecServiceOptions gathers options to initialize a service as an HTTP service.
type HTTPSpecServiceOptions struct {
	ProtoHTTPServer http_spec.API
}

// Kind returns the type of service implemented by HTTPSpecServiceOptions as
// definition.RuntimeTypeHTTPSpec.
func (h *HTTPSpecServiceOptions) Kind() definition.RuntimeType {
	return definition.RuntimeTypeHTTPSpec
}
