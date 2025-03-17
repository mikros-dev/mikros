package options

import (
	httpServiceAPI "github.com/mikros-dev/mikros/apis/services/http"
	"github.com/mikros-dev/mikros/components/definition"
)

// HttpServiceOptions gathers options to initialize a service as an HTTP service.
type HttpServiceOptions struct {
	ProtoHttpServer httpServiceAPI.HttpServer
}

func (h *HttpServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_HTTP
}
