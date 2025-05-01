package behavior

import (
	"context"
)

// Tracker is an interface that a plugin should implement to provide a way to
// add a unique tracing value in each call between services.
type Tracker interface {
	// Generate is responsible for creating a new unique tracker ID.
	Generate() string

	// Add adds a tracker ID into the context and return a new one, updated.
	Add(ctx context.Context, id string) context.Context

	// Retrieve tries to retrieve the tracker ID from the current context and
	// return it.
	Retrieve(ctx context.Context) (string, bool)
}
