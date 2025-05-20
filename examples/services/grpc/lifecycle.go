package main

import (
	"context"

	"github.com/mikros-dev/mikros/components/logger"
)

func (s *service) OnStart(ctx context.Context) error {
	s.Logger.Info(ctx, "OnStart called", logger.Any("definitions", s.Definitions))
	return nil
}
