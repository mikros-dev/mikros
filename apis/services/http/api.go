package http

import (
	"context"
	"net/http"
)

type HttpAPI interface {
	HTTPHandler(ctx context.Context) (http.Handler, error)
}
