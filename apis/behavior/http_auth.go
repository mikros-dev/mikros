package behavior

import (
	"context"
)

// Authenticator defines the contract for authentication plugins used in the HTTP
// service layer. Implementations of this interface are responsible for registering
// authentication-related handlers (such as middleware or endpoint-specific logic)
// into the HTTP service's handler chain.
//
// This allows the HTTP service to integrate authentication behavior (e.g., token
// validation, identity extraction, etc.) in a pluggable and modular way.
type Authenticator interface {
	// AuthHandlers returns a function that injects authentication-related handlers
	// into the HTTP service, along with any initialization error.
	AuthHandlers() (func(ctx context.Context, handlers map[string]interface{}) error, error)
}
