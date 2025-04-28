package services

import (
	"context"
)

// ServiceAPI is the API from mikros available inside lifecycle handlers.
type ServiceAPI interface {
	// Feature is a mechanism that allows a service having access to a feature
	// by loading target argument with its public API. It searches the feature
	// by the type target is declared. It returns error if the feature is not
	// found.
	//
	// Example:
	// func (s *service) OnStart(ctx context.Context, svc ServiceAPI) error {
	//		var logger flogger.LoggerAPI
	//		if err := svc.Feature(ctx, &logger); err != nil {
	//			return err
	//		}
	//
	//		// Do something with logger.
	//		logger.Info(ctx, "this is just a sample message")
	//		return nil
	// }
	Feature(ctx context.Context, target interface{}) error

	// Abort is a helper method to abort services in the right way, when external
	// initialization is needed.
	Abort(message string, err error)
}
