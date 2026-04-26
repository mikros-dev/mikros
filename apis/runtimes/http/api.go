package http

import (
	"context"
	"net/http"
)

// API represents an interface defining the contract for an HTTP service.
type API interface {
	HTTPHandler(ctx context.Context) (http.Handler, error)
}
