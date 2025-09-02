package http

import (
	"context"
)

// HttpServerAPI provides methods to interact with the current HTTP response
// within a request handler.
//
// This interface is implemented by the mikros framework and made available to
// services that opt into the feature. It allows handler logic to customize the
// HTTP response dynamically, such as setting headers and status codes.
type HttpServerAPI interface {
	// AddResponseHeader adds a new header entry to the response associated
	// with the given context. If the same key is added multiple times, the
	// values are appended.
	AddResponseHeader(ctx context.Context, key, value string)

	// SetResponseCode sets a custom HTTP status code for the response
	// associated with the given context. This overrides the default 200 OK
	// status.
	SetResponseCode(ctx context.Context, code int)
}
