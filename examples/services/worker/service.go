package main

import (
	"context"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

type service struct {
	Logger logger_api.API `mikros:"feature"`
}

func (s *service) Start(ctx context.Context) error {
	s.Logger.Info(ctx, "service Start")
	return nil
}

func (s *service) Stop(ctx context.Context) error {
	s.Logger.Info(ctx, "service Stop")
	return nil
}
