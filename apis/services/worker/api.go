package worker

import (
	"context"
)

// WorkerAPI corresponds to the API that a worker service must implement in
// its main structure.
type WorkerAPI interface {
	// Start must put the service in execution. It can block and wait for some
	// signal to finish, which should be done in the Stop call. If it needs
	// a loop, i.e., execute forever, it must be done at the service level.
	// One way to do it is to use ctx.Done() to check if the loop needs to
	// finish.
	Start(ctx context.Context) error

	// Stop should end the Run execution.
	Stop(ctx context.Context) error
}
