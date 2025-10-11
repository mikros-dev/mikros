package lifecycle

import (
	"context"

	"github.com/mikros-dev/mikros/components/definition"
)

// Options defines the configuration for service lifecycle controls,
// including the environment and test execution settings.
type Options struct {
	Env            definition.ServiceDeploy
	ExecuteOnTests bool
}

// OnStart initializes the service lifecycle, invoking its OnStart method if
// it implements ServiceLifecycleStarter.
func OnStart(ctx context.Context, s interface{}, opt *Options) error {
	if !shouldExecute(opt) {
		return nil
	}

	if l, ok := s.(ServiceLifecycleStarter); ok {
		return l.OnStart(ctx)
	}

	return nil
}

// OnFinish triggers the OnFinish lifecycle method for a service if it implements
// ServiceLifecycleFinisher and execution is allowed.
func OnFinish(ctx context.Context, s interface{}, opt *Options) {
	if !shouldExecute(opt) {
		return
	}

	if l, ok := s.(ServiceLifecycleFinisher); ok {
		l.OnFinish(ctx)
	}
}

func shouldExecute(opt *Options) bool {
	// Do not execute lifecycle events by default in tests to force them to mock
	// features that are being initialized by the service.
	if opt.Env == definition.ServiceDeployTest {
		return opt.ExecuteOnTests
	}

	return true
}
