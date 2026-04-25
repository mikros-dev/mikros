package main

import (
	"context"
)

func (s *service) OnStart(ctx context.Context) error {
	s.Logger.Info(ctx, "lifecycle OnStart")
	return nil
}

func (s *service) OnFinish(ctx context.Context) {
	s.Logger.Info(ctx, "lifecycle OnFinish")
}