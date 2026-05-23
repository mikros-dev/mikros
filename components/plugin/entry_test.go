package plugin

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

type fakeErrorAPI struct {
	last *fakeErrorBuilder
}

type fakeErrorBuilder struct {
	err error
}

func (f *fakeErrorAPI) RPC(err error, _ string) errors_api.Value {
	f.last = &fakeErrorBuilder{
		err: err,
	}

	return f.last
}

func (f *fakeErrorAPI) InvalidArgument(err error) errors_api.Value {
	f.last = &fakeErrorBuilder{
		err: err,
	}

	return f.last
}

func (f *fakeErrorAPI) FailedPrecondition(message string) errors_api.Value {
	f.last = &fakeErrorBuilder{
		err: errors.New(message),
	}

	return f.last
}

func (f *fakeErrorAPI) NotFound() errors_api.Value {
	f.last = &fakeErrorBuilder{
		err: errors.New("not found"),
	}

	return f.last
}

func (f *fakeErrorAPI) Internal(err error) errors_api.Value {
	f.last = &fakeErrorBuilder{
		err: err,
	}

	return f.last
}

func (f *fakeErrorAPI) PermissionDenied() errors_api.Value {
	f.last = &fakeErrorBuilder{
		err: errors.New("permission denied"),
	}

	return f.last
}

func (f *fakeErrorBuilder) WithCode(_ errors_api.Code) errors_api.Value {
	return f
}

func (f *fakeErrorBuilder) WithAttributes(_ ...logger_api.Attribute) errors_api.Value {
	return f
}

func (f *fakeErrorBuilder) Error() string {
	return f.err.Error()
}

func TestEntryUpdateInfoAndHelpers(t *testing.T) {
	var (
		e      Entry
		errAPI = &fakeErrorAPI{}
	)

	e.UpdateInfo(UpdateInfoEntry{
		Enabled: true,
		Name:    "my_feature",
		Errors:  errAPI,
	})

	assert.True(t, e.IsEnabled())
	assert.Equal(t, "my_feature", e.Name())
	assert.Nil(t, e.Logger())
}

func TestEntryError(t *testing.T) {
	var e Entry

	e.UpdateInfo(UpdateInfoEntry{
		Name: "feature",
	})

	assert.EqualError(t, e.Error("boom"), "feature: boom")
	assert.EqualError(t, e.Error(errors.New("bad")), "feature: bad")
	assert.EqualError(t, e.Error(42), "feature: 42")
}

func TestEntryWrapError(t *testing.T) {
	var (
		e      Entry
		errAPI = &fakeErrorAPI{}
	)

	e.UpdateInfo(UpdateInfoEntry{
		Name:   "feature",
		Errors: errAPI,
	})

	err := e.WrapError(context.Background(), errors.New("boom"))
	require.Error(t, err)
	assert.EqualError(t, err, "boom")
	require.NotNil(t, errAPI.last)
}

func TestEntryWrapErrorWithNil(t *testing.T) {
	var (
		e      Entry
		errAPI = &fakeErrorAPI{}
	)

	e.UpdateInfo(UpdateInfoEntry{
		Name:   "feature",
		Errors: errAPI,
	})

	err := e.WrapError(context.Background(), nil)
	require.Error(t, err)
	assert.EqualError(t, err, "unknown internal feature error")
}
