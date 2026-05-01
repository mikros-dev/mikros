package plugin

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

type fakeRuntime struct {
	name string
}

func (f fakeRuntime) Name() string {
	return f.name
}

func (f fakeRuntime) Info() []logger_api.Attribute {
	return nil
}

func (f fakeRuntime) Initialize(context.Context, *RuntimeOptions) error {
	return nil
}

func (f fakeRuntime) Run(context.Context, interface{}) error {
	return nil
}

func (f fakeRuntime) Stop(context.Context) error {
	return nil
}

func TestRuntimeSetRegister(t *testing.T) {
	set := NewRuntimeSet()

	set.Register(fakeRuntime{name: ""})
	assert.Equal(t, 0, len(set.Runtimes()))

	set.Register(fakeRuntime{name: "http"})
	set.Register(fakeRuntime{name: "http"})
	assert.Equal(t, 1, len(set.Runtimes()))
}

func TestRuntimeSetReturnsCopy(t *testing.T) {
	set := NewRuntimeSet()
	set.Register(fakeRuntime{name: "http"})

	services := set.Runtimes()
	delete(services, "http")

	assert.Equal(t, 1, len(set.Runtimes()))
}

func TestRuntimeSetAppend(t *testing.T) {
	var (
		left = NewRuntimeSet()
		right = NewRuntimeSet()
	)

	left.Register(fakeRuntime{name: "http"})
	right.Register(fakeRuntime{name: "grpc"})
	right.Register(fakeRuntime{name: "http"})

	left.Append(right)
	assert.Equal(t, 2, len(left.Runtimes()))
}
