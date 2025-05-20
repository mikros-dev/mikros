package behavior

import (
	"context"
)

// Tracker defines the contract for plugins that manage request tracking across
// service boundaries in mikros-based systems.
//
// This interface is used to generate and propagate a unique tracking ID for each
// service call or task execution. The tracker ID enables correlation of logs,
// traces, and metrics across multiple services, making it easier to follow a
// request's journey through the system.
//
// Typical implementations may store the ID in the context, headers, or metadata
// depending on the transport layer.
//
// Example:
//
//	func (t *MyTracker) Generate() string {
//	    return uuid.New().String()
//	}
//
//	func (t *MyTracker) Add(ctx context.Context, id string) context.Context {
//	    return context.WithValue(ctx, trackerKey, id)
//	}
//
//	func (t *MyTracker) Retrieve(ctx context.Context) (string, bool) {
//	    val, ok := ctx.Value(trackerKey).(string)
//	    return val, ok
//	}
type Tracker interface {
	// Generate creates a new unique tracker ID to identify a request or task
	// execution.
	Generate() string

	// Add inserts the given tracker ID into the provided context and returns
	// the updated context.
	Add(ctx context.Context, id string) context.Context

	// Retrieve attempts to extract the tracker ID from the given context.
	// Returns the ID and a boolean indicating whether it was found.
	Retrieve(ctx context.Context) (string, bool)
}
