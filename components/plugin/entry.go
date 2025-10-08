package plugin

import (
	"context"
	"errors"
	"fmt"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/logger"
)

// Entry is a member that all framework features must have declared inside
// it (as a struct member). It implements the FeatureEntry interface for the
// feature if used.
//
// Also, if a feature uses it, it already receives a logger.API interface
// for it for free and error methods to return a proper error for services.
type Entry struct {
	featureEnabled bool
	featureName    string
	logger         logger_api.API
	errors         errors_api.ErrorAPI
}

// UpdateInfo is an internal method that allows a feature to have its
// information, such as its name, if it's enabled or not, internally.
func (e *Entry) UpdateInfo(info UpdateInfoEntry) {
	if info.Errors != nil {
		e.errors = info.Errors
	}

	if info.Logger != nil {
		e.logger = info.Logger
	}

	if info.Name != "" {
		e.featureName = info.Name
	}

	e.featureEnabled = info.Enabled
}

// IsEnabled is a helper function that every public feature API should call
// at its beginning, to avoid executing it if it is disabled.
func (e *Entry) IsEnabled() bool {
	return e.featureEnabled
}

// Name returns the internal feature name.
func (e *Entry) Name() string {
	return e.featureName
}

// Logger is a helper method that gives the feature access to the logger API.
func (e *Entry) Logger() logger_api.API {
	return e.logger
}

// Error is a helper API to create an error value from a feature using a standard
// for all of them.
func (e *Entry) Error(in interface{}) error {
	var msg string

	switch v := in.(type) {
	case string:
		msg = v
	case error:
		msg = v.Error()
	default:
		msg = fmt.Sprint(v)
	}

	return fmt.Errorf("%s: %s", e.featureName, msg)
}

// WrapError is a helper API to create an error value from another error using
// the error standard of services APIs.
//
// Usually, this method should be used when a public feature API returns an
// error.
func (e *Entry) WrapError(ctx context.Context, err error) error {
	if err == nil {
		err = errors.New("unknown internal feature error")
	}

	return e.errors.Internal(err).
		WithAttributes(logger.String("feature.name", e.featureName)).
		Submit(ctx)
}
