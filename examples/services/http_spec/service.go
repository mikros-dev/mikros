package main

import (
	"context"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"

	user_bffpb "github.com/mikros-dev/mikros/examples/protobuf-workspace/gen/go/services/user_bff"
)

type service struct {
	Logger logger_api.LoggerAPI `mikros:"feature"`
}

func (s *service) CreateUser(ctx context.Context, req *user_bffpb.CreateUserRequest) (*user_bffpb.CreateUserResponse, error) {
	s.Logger.Info(ctx, "the real handler")
	return &user_bffpb.CreateUserResponse{}, nil
}
