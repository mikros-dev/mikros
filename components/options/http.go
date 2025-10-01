package options

import (
	"net/http"
	"time"

	"github.com/mikros-dev/mikros/components/definition"
)

// HttpServiceOptions defines runtime options for an HTTP service.
type HttpServiceOptions struct {
	// CORSStrict controls how invalid CORS configurations are handled if a
	// CORS middleware implementation is supplied. When true, invalid CORS
	// settings cause service initialization to fail. Otherwise, a warning
	// is emitted and the middleware is disabled.
	CORSStrict bool

	// BasePath is a common URL prefix under which all routes of this service
	// are mounted. For example, if BasePath = "/api", a handler registered at
	// "/items" will be served at "/api/items". An empty string mounts the
	// service at root.
	BasePath string

	// ReadTimeout is the maximum duration allowed for reading the entire
	// request, including the body. A zero value uses the Mikros default (15 s).
	ReadTimeout time.Duration

	// WriteTimeout is the maximum duration before timing out writes of the
	// response. A zero value uses the Mikros default (15 s).
	WriteTimeout time.Duration

	// IdleTimeout is the maximum time to wait for the next request when keep-alive
	// is enabled. A zero value uses the Mikros default (60 s).
	IdleTimeout time.Duration

	// MaxHeaderBytes controls the maximum number of bytes the server will
	// read parsing request headers. A zero value uses the Go standard
	// library default (1 MiB).
	MaxHeaderBytes int

	// Middlewares is a slice of user-supplied HTTP middlewares in the form
	// func(http.Handler) http.Handler. They are composed after core middlewares
	// (such as CORS and authentication). The first element in the slice becomes
	// the outermost wrapper.
	Middlewares []func(handler http.Handler) http.Handler
}

func (h *HttpServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_HTTP
}
