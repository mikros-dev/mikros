package main

import (
	"context"

	"github.com/somatech1/mikros"
)

type service struct {
	*mikros.Service
}

func (s *service) Run(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (s *service) Cleanup(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}
