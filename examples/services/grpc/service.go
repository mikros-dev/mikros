package main

import (
	"context"
	"errors"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	userpb "github.com/mikros-dev/mikros/examples/protobuf-workspace/gen/go/services/user"
)

type service struct {
	Definitions *Definitions         `mikros:"definitions"`
	Logger      logger_api.LoggerAPI `mikros:"feature"`
}

type Definitions struct {
	Foo string
	Bar int
}

func (d *Definitions) Validate() error {
	if d.Foo == "" {
		return errors.New("field foo is required")
	}

	return nil
}

func (s *service) GetUserByID(ctx context.Context, req *userpb.GetUserByIDRequest) (*userpb.GetUserByIDResponse, error) {
	return &userpb.GetUserByIDResponse{}, nil
}

func (s *service) GetUsers(ctx context.Context, req *userpb.GetUsersRequest) (*userpb.GetUsersResponse, error) {
	return &userpb.GetUsersResponse{}, nil
}

func (s *service) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.CreateUserResponse, error) {
	return &userpb.CreateUserResponse{}, nil
}

func (s *service) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UpdateUserResponse, error) {
	return &userpb.UpdateUserResponse{}, nil
}

func (s *service) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*userpb.DeleteUserResponse, error) {
	return &userpb.DeleteUserResponse{}, nil
}
