package lifecycle

import (
	"context"
)

// ServiceLifecycleStarter is an optional behavior that a service can have to
// receive notifications when the service is initializing.
type ServiceLifecycleStarter interface {
	// OnStart is a method called right before the service enters its infinite
	// execution mode, when Service.Start API is called, before database migrations
	// and service structure fields validation.
	//
	// It is also the right place for the service to initialize something that
	// requires accessing the framework.Service API or initialize specific fields
	// in its main structure.
	OnStart(ctx context.Context) error
}

// ServiceLifecycleFinisher is an optional behavior that a service can have to
// receive notifications when the service is finishing.
type ServiceLifecycleFinisher interface {
	// OnFinish is the method called before the service is finished by the framework.
	OnFinish(ctx context.Context)
}
