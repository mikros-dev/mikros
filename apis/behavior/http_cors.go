package behavior

import (
	"github.com/lab259/cors"
)

// Handler is a behavior that an HTTP cors plugin must implement if one
// wants CORS implemented in the HTTP server.
type Handler interface {
	// Cors is a method that must return the CORS options for an HTTP server.
	Cors() cors.Options
}
