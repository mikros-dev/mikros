package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcontext "github.com/mikros-dev/mikros/components/context"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

type fakeIntegration struct {
	Entry
	id            string
	allow         bool
	initCalled    bool
	initOrder     *[]string
	startCalled   bool
	cleanupCalled bool
	initErr       error
	startErr      error
	cleanupErr    error
}

func (f *fakeIntegration) API() interface{} {
	return nil
}

func (f *fakeIntegration) CanBeInitialized(_ *CanBeInitializedOptions) bool {
	return f.allow
}

func (f *fakeIntegration) Initialize(_ context.Context, options *InitializeOptions) error {
	f.initCalled = true
	if f.initOrder != nil {
		*f.initOrder = append(*f.initOrder, f.id)
	}

	return f.initErr
}

func (f *fakeIntegration) Start(_ context.Context, _ interface{}) error {
	f.startCalled = true
	return f.startErr
}

func (f *fakeIntegration) Cleanup(_ context.Context) error {
	f.cleanupCalled = true
	return f.cleanupErr
}

func TestIntegrationSetRegisterAndIntegration(t *testing.T) {
	var (
		set         = NewIntegrationSet()
		integration = &fakeIntegration{}
	)

	set.Register("test_integration", integration)

	assert.Equal(t, 1, set.Count())
	got, err := set.Integration("test_integration")
	require.NoError(t, err)
	assert.Same(t, integration, got)
	assert.Equal(t, "test_integration", integration.Name())
}

func TestIntegrationSetRegisterReplacesIntegrationInPlace(t *testing.T) {
	var (
		set   = NewIntegrationSet()
		order = []string{}
		first = &fakeIntegration{
			id:        "first",
			allow:     true,
			initOrder: &order,
		}
		second = &fakeIntegration{
			id:        "second",
			allow:     true,
			initOrder: &order,
		}
		before = &fakeIntegration{
			id:        "before",
			allow:     true,
			initOrder: &order,
		}
	)

	set.Register("before", before)
	set.Register("test_integration", first)
	set.Register("test_integration", second)

	assert.Equal(t, 2, set.Count())

	got, err := set.Integration("test_integration")
	require.NoError(t, err)
	assert.Same(t, second, got)

	err = set.InitializeAll(context.Background(), &InitializeOptions{
		Env:         fakeEnv{},
		Definitions: &definition.Definitions{},
	})
	require.NoError(t, err)

	assert.False(t, first.initCalled)
	assert.True(t, second.initCalled)
	assert.Equal(t, []string{"before", "second"}, order)
}

func TestIntegrationSetInitializeAll(t *testing.T) {
	var (
		set  = NewIntegrationSet()
		main = &fakeIntegration{
			allow: true,
		}
		skip = &fakeIntegration{
			allow: false,
		}
	)

	set.Register("main", main)
	set.Register("skip", skip)

	svcCtx, err := mcontext.New(&mcontext.Options{Name: service.FromString("svc")})
	require.NoError(t, err)

	err = set.InitializeAll(context.Background(), &InitializeOptions{
		Env:            fakeEnv{},
		Definitions:    &definition.Definitions{},
		ServiceContext: svcCtx,
	})
	require.NoError(t, err)

	assert.True(t, main.initCalled)
	assert.False(t, skip.initCalled)
}

func TestIntegrationSetInitializeAllStopsOnError(t *testing.T) {
	var (
		set  = NewIntegrationSet()
		fail = &fakeIntegration{
			allow:   true,
			initErr: errors.New("init failed"),
		}
		after = &fakeIntegration{
			allow: true,
		}
	)

	set.Register("fail", fail)
	set.Register("after", after)

	err := set.InitializeAll(context.Background(), &InitializeOptions{
		Env:         fakeEnv{},
		Definitions: &definition.Definitions{},
	})
	require.Error(t, err)
	assert.EqualError(t, err, "init failed")
	assert.False(t, after.initCalled)
}

func TestIntegrationSetStartAllAndCleanupAll(t *testing.T) {
	var (
		set = NewIntegrationSet()
		a   = &fakeIntegration{
			allow: true,
		}
		b = &fakeIntegration{
			allow: true,
		}
	)

	set.Register("a", a)
	set.Register("b", b)

	require.NoError(t, set.StartAll(context.Background(), struct{}{}))
	require.NoError(t, set.CleanupAll(context.Background()))

	assert.True(t, a.startCalled)
	assert.True(t, b.startCalled)
	assert.True(t, a.cleanupCalled)
	assert.True(t, b.cleanupCalled)
}

func TestIntegrationSetAppend(t *testing.T) {
	var (
		left  = NewIntegrationSet()
		right = NewIntegrationSet()
		a     = &fakeIntegration{}
		b     = &fakeIntegration{}
	)

	left.Register("a", a)
	right.Register("b", b)

	left.Append(right)
	assert.Equal(t, 2, left.Count())
}
