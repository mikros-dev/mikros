package http

import (
	"context"

	"github.com/fasthttp/router"
)

// HttpServerAPI is the behavior that a service must implement to be accepted as
// a valid framework HTTP service.
type HttpServerAPI interface {
	// SetupServer is the place where a service can adjust and initialize
	// everything it requires to successfully initialize the HTTP server later.
	SetupServer(
		serviceName string,
		logger interface{},
		router *router.Router,
		apiHandlers interface{},
		authHandler func(ctx context.Context, handlers map[string]interface{}) error,
	) error
}
