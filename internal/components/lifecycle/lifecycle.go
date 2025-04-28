package lifecycle

import (
	"context"

	"github.com/mikros-dev/mikros/components/definition"
)

type LifecycleOptions struct {
	Env            definition.ServiceDeploy
	ExecuteOnTests bool
}

func OnStart(s interface{}, ctx context.Context, opt *LifecycleOptions) error {
	if !shouldExecute(opt) {
		return nil
	}

	if l, ok := s.(ServiceLifecycleStarter); ok {
		return l.OnStart(ctx)
	}

	return nil
}

func OnFinish(s interface{}, ctx context.Context, opt *LifecycleOptions) {
	if !shouldExecute(opt) {
		return
	}

	if l, ok := s.(ServiceLifecycleFinisher); ok {
		l.OnFinish(ctx)
	}
}

func shouldExecute(opt *LifecycleOptions) bool {
	// Do not execute lifecycle events by default in tests to force them mock
	// features that are being initialized by the service.
	if opt.Env == definition.ServiceDeploy_Test {
		return opt.ExecuteOnTests
	}

	return true
}
