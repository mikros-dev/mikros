package behavior

import (
	"context"
)

// Recovery defines the contract for panic recovery plugins used in the HTTP
// service layer. Implementations of this interface are responsible for handling
// unexpected panics during HTTP request processing, allowing the service to
// recover gracefully and avoid crashes.
//
// This enables the HTTP service to encapsulate panic recovery behavior in a
// modular way, supporting custom recovery strategies such as logging, metrics,
// or fallback responses.
type Recovery interface {
	// Recover is invoked when a panic occurs during request processing.
	// It receives the current request context and is expected to handle
	// recovery logic.
	Recover(ctx context.Context)
}
