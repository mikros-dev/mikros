package behavior

import (
	"context"
)

// Tracer is an interface that a tracing feature plugin should implement to be
// used by all internal supported services.
type Tracer interface {
	// StartMeasurements must retrieve required information from the current
	// application context and initialize any internal information that will
	// be used as metrics for it.
	StartMeasurements(ctx context.Context, serviceType string) (interface{}, error)

	// ComputeMetrics receives an updated context and data returned by the
	// StartMeasurements method to compute the current application metrics.
	ComputeMetrics(ctx context.Context, serviceType string, data interface{}) error
}
