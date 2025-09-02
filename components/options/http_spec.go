package options

import (
	"github.com/mikros-dev/mikros/apis/services/http_spec"
	"github.com/mikros-dev/mikros/components/definition"
)

// HttpSpecServiceOptions gathers options to initialize a service as an HTTP service.
type HttpSpecServiceOptions struct {
	ProtoHttpServer http_spec.HttpSpecServerAPI
}

func (h *HttpSpecServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_HTTPSpec
}
