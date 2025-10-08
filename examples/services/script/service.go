package main

import (
	"context"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

type service struct {
	Logger logger_api.API `mikros:"feature"`
}

func (s *service) Run(ctx context.Context) error {
	s.Logger.Info(ctx, "service Run")
	return nil
}

func (s *service) Cleanup(ctx context.Context) error {
	s.Logger.Info(ctx, "service Cleanup")
	return nil
}
