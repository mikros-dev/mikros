package behavior

import (
	"context"
)

// Recovery is a behavior that an HTTP panic recovery feature (plugin)
// must implement to be used inside the HTTP service implementation.
type Recovery interface {
	Recover(ctx context.Context)
}
