package behavior

import (
	"github.com/lab259/cors"
)

// CorsHandler defines the contract for CORS plugins used in the HTTP service layer.
// Implementations of this interface are responsible for providing CORS configuration
// that will be applied to the HTTP server.
//
// This enables the HTTP service to support Cross-Origin Resource Sharing (CORS)
// in a modular and configurable way, allowing for dynamic control over origin,
// method, and header access policies.
type CorsHandler interface {
	// Cors returns the CORS options that should be applied to the HTTP server.
	// These options control how cross-origin requests are handled.
	Cors() cors.Options
}
