package behavior

import (
	"context"
)

// Tracer defines the contract for plugins that enable tracing and performance
// measurement across all service types supported by mikros.
//
// Implementations of this interface are responsible for collecting runtime
// metrics and diagnostic data throughout the lifecycle of a service request
// or execution unit. This enables detailed observability, latency tracking,
// and performance analysis in a modular and extensible way.
//
// Example:
//
//	func (t *MyTracer) StartMeasurements(ctx context.Context, serviceName string) (interface{}, error) {
//	    start := time.Now()
//	    return start, nil
//	}
//
//	func (t *MyTracer) ComputeMetrics(ctx context.Context, serviceName string, data interface{}) error {
//	    startTime, ok := data.(time.Time)
//	    if !ok {
//	        return errors.New("invalid metric data")
//	    }
//	    duration := time.Since(startTime)
//	    log.Printf("[%s] request took %s", serviceName, duration)
//	    return nil
//	}
type Tracer interface {
	// StartMeasurements initializes tracing or metric collection using the given context.
	// It is called at the beginning of a request or task execution, and may extract and record
	// service-specific metadata. The returned value will be passed unchanged to ComputeMetrics.
	StartMeasurements(ctx context.Context, serviceName string) (interface{}, error)

	// ComputeMetrics finalizes the metric computation using the updated context and the
	// data returned by StartMeasurements. This method is typically called at the end of
	// a request or execution and is responsible for emitting traces, recording durations,
	// or reporting collected metrics.
	ComputeMetrics(ctx context.Context, serviceName string, data interface{}) error
}
