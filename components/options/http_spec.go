package options

import (
	"github.com/mikros-dev/mikros/apis/services/http_spec"
	"github.com/mikros-dev/mikros/components/definition"
)

// HTTPSpecServiceOptions gathers options to initialize a service as an HTTP service.
type HTTPSpecServiceOptions struct {
	ProtoHTTPServer http_spec.API
}

// Kind returns the type of service implemented by HTTPSpecServiceOptions as
// definition.ServiceTypeHTTPSpec.
func (h *HTTPSpecServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceTypeHTTPSpec
}
