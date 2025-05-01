package behavior

import (
	"context"
)

// Authenticator is a behavior that an HTTP authentication feature (plugin)
// must implement to be used inside the HTTP service implementation.
type Authenticator interface {
	AuthHandlers() (func(ctx context.Context, handlers map[string]interface{}) error, error)
}
