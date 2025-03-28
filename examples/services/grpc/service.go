package main

import (
	"context"

	"github.com/somatech1/mikros/protobuf-examples/gen/go/services/user"
)

type service struct{}

func (s *service) GetUserByID(ctx context.Context, req *user.GetUserByIDRequest) (*user.GetUserByIDResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) GetUsers(ctx context.Context, req *user.GetUsersRequest) (*user.GetUsersResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) CreateUser(ctx context.Context, req *user.CreateUserRequest) (*user.CreateUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) UpdateUser(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *service) DeleteUser(ctx context.Context, req *user.DeleteUserRequest) (*user.DeleteUserResponse, error) {
	//TODO implement me
	panic("implement me")
}
