package script

import (
	"context"
)

// ScriptApi corresponds to the API that a script service must implement in
// its main structure.
type ScriptApi interface {
	// Run must be the service function where things happen. It is executed
	// only once and the service terminates.
	//
	// Services should avoid blocking this function since there are other
	// type of services for this purpose.
	Run(ctx context.Context) error

	// Cleanup must clean or finish anything that was initialized or any resource
	// that need to be released.
	Cleanup(ctx context.Context) error
}
