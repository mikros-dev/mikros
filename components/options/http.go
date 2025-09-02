package options

import (
	"net/http"
	"time"

	"github.com/mikros-dev/mikros/components/definition"
)

type HttpServiceOptions struct {
	// BasePath is a common prefix for all routes
	BasePath       string
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
	IdleTimeout    time.Duration
	MaxHeaderBytes int
	Middlewares    []func(handler http.Handler) http.Handler
}

func (h *HttpServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_HTTP
}
