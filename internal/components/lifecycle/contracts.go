package lifecycle

import (
	"context"
)

// ServiceLifecycleStarter is an optional behavior that a service can have to
// receive notifications when the service is initializing.
type ServiceLifecycleStarter interface {
	// OnStart is a method called right before the service enters its (infinite)
	// execution mode, when Service.Start API is called. When called, all features
	// declared inside the main structure service are already initialized and
	// available to use.
	//
	// It is the right place for the service to initialize its external resources
	// or specific members from its main structure.
	OnStart(ctx context.Context) error
}

// ServiceLifecycleFinisher is an optional behavior that a service can have to
// receive notifications when the service is finishing.
type ServiceLifecycleFinisher interface {
	// OnFinish is the method called before the service is finished. Resources
	// should be released here.
	OnFinish(ctx context.Context)
}
