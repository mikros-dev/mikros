package main

import (
	"context"
	"fmt"

	"github.com/somatech1/mikros"

	"github.com/somatech1/mikros/protobuf-examples/gen/go/services/user"
	"github.com/somatech1/mikros/protobuf-examples/gen/go/services/user_bff"
)

type service struct {
	*mikros.Service

	user.UserServiceServer `mikros:"grpc_client=user"`
}

func (s *service) CreateUser(ctx context.Context, req *user_bff.CreateUserRequest) (*user_bff.CreateUserResponse, error) {
	fmt.Println("CreateUser RPC call:", req)

	return &user_bff.CreateUserResponse{
		User: &user.UserWire{
			Name:  req.Name,
			Email: req.Email,
			Age:   req.Age,
		},
	}, nil
}
