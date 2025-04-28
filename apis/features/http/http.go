package http

import (
	"context"
)

type HttpServerAPI interface {
	// AddResponseHeader adds a new header entry for the handler's response.
	AddResponseHeader(ctx context.Context, key, value string)

	// SetResponseCode sets a custom response code for the handler's response.
	SetResponseCode(ctx context.Context, code int)
}
