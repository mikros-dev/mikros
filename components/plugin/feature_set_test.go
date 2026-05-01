package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	env_api "github.com/mikros-dev/mikros/apis/features/env"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	mcontext "github.com/mikros-dev/mikros/components/context"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

type fakeEnv struct{}

func (fakeEnv) Get(string) string {
	return ""
}

func (fakeEnv) GetInt(string) (int, error) {
	return 0, nil
}

func (fakeEnv) GetBool(string) (bool, error) {
	return false, nil
}

func (fakeEnv) DeploymentEnv() definition.DeploymentEnv {
	return definition.DeploymentEnvLocal
}

func (fakeEnv) TrackerHeaderName() string {
	return ""
}

func (fakeEnv) IsCICD() bool {
	return false
}

func (fakeEnv) CoupledNamespace() string {
	return ""
}

func (fakeEnv) CoupledPort() int32 {
	return 0
}

func (fakeEnv) GrpcPort() int32 {
	return 0
}

func (fakeEnv) HTTPPort() int32 {
	return 0
}

var _ env_api.API = fakeEnv{}

type fakeFeature struct {
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
	lastDeps      map[string]Feature
}

func (f *fakeFeature) CanBeInitialized(_ *CanBeInitializedOptions) bool {
	return f.allow
}

func (f *fakeFeature) Initialize(_ context.Context, options *InitializeOptions) error {
	f.initCalled = true
	if f.initOrder != nil {
		*f.initOrder = append(*f.initOrder, f.id)
	}

	f.lastDeps = options.Dependencies
	return f.initErr
}

func (f *fakeFeature) Fields() []logger_api.Attribute {
	return nil
}

func (f *fakeFeature) Start(_ context.Context, _ interface{}) error {
	f.startCalled = true
	return f.startErr
}

func (f *fakeFeature) Cleanup(_ context.Context) error {
	f.cleanupCalled = true
	return f.cleanupErr
}

func TestFeatureSetRegisterAndFeature(t *testing.T) {
	var (
		set  = NewFeatureSet()
		feat = &fakeFeature{}
	)

	set.Register("test_feature", feat)

	assert.Equal(t, 1, set.Count())
	got, err := set.Feature("test_feature")
	require.NoError(t, err)
	assert.Same(t, feat, got)
	assert.Equal(t, "test_feature", feat.Name())
}

func TestFeatureSetRegisterReplacesFeatureInPlace(t *testing.T) {
	var (
		set   = NewFeatureSet()
		order = []string{}
		dep   = &fakeFeature{
			id:        "dep",
			allow:     true,
			initOrder: &order,
		}
		first = &fakeFeature{
			id:        "first",
			allow:     true,
			initOrder: &order,
		}
		second = &fakeFeature{
			id:        "second",
			allow:     true,
			initOrder: &order,
		}
	)

	set.Register("dep", dep)
	set.Register("test_feature", first, "dep")
	set.Register("test_feature", second, "dep")

	assert.Equal(t, 2, set.Count())

	got, err := set.Feature("test_feature")
	require.NoError(t, err)
	assert.Same(t, second, got)

	err = set.InitializeAll(context.Background(), &InitializeOptions{
		Env:         fakeEnv{},
		Definitions: &definition.Definitions{},
	})
	require.NoError(t, err)

	assert.False(t, first.initCalled)
	assert.True(t, second.initCalled)
	assert.Equal(t, []string{"dep", "second"}, order)
	require.Contains(t, second.lastDeps, "dep")
}

func TestFeatureSetInitializeAll(t *testing.T) {
	var (
		set = NewFeatureSet()
		dep = &fakeFeature{
			allow: true,
		}
		main = &fakeFeature{
			allow: true,
		}
		skip = &fakeFeature{
			allow: false,
		}
	)

	set.Register("dep", dep)
	set.Register("main", main, "dep")
	set.Register("skip", skip)

	svcCtx, err := mcontext.New(&mcontext.Options{Name: service.FromString("svc")})
	require.NoError(t, err)

	err = set.InitializeAll(context.Background(), &InitializeOptions{
		Env:            fakeEnv{},
		Definitions:    &definition.Definitions{},
		ServiceContext: svcCtx,
	})
	require.NoError(t, err)

	assert.True(t, dep.initCalled)
	assert.True(t, main.initCalled)
	assert.False(t, skip.initCalled)
	require.Contains(t, main.lastDeps, "dep")
}

func TestFeatureSetInitializeAllStopsOnError(t *testing.T) {
	var (
		set  = NewFeatureSet()
		fail = &fakeFeature{
			allow:   true,
			initErr: errors.New("init failed"),
		}
		after = &fakeFeature{
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

func TestFeatureSetStartAllAndCleanupAll(t *testing.T) {
	var (
		set = NewFeatureSet()
		a   = &fakeFeature{
			allow: true,
		}
		b = &fakeFeature{
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

func TestFeatureSetIteratorReset(t *testing.T) {
	var (
		set = NewFeatureSet()
		a   = &fakeFeature{}
		b   = &fakeFeature{}
	)

	set.Register("a", a)
	set.Register("b", b)

	it := set.Iterator()
	first, ok := it.Next()
	require.True(t, ok)
	assert.Same(t, a, first)

	_, _ = it.Next()
	_, ok = it.Next()
	assert.False(t, ok)

	it.Reset()
	first, ok = it.Next()
	require.True(t, ok)
	assert.Same(t, a, first)
}

func TestFeatureSetAppend(t *testing.T) {
	var (
		left  = NewFeatureSet()
		right = NewFeatureSet()
		a     = &fakeFeature{}
		b     = &fakeFeature{}
	)

	left.Register("a", a)
	right.Register("b", b)

	left.Append(right)
	assert.Equal(t, 2, left.Count())
}
