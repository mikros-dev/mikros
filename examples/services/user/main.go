package main

import (
	"github.com/mikros-dev/mikros"
	"github.com/mikros-dev/mikros/components/options"

	userpb "github.com/mikros-dev/mikros/examples/protobuf-workspace/gen/go/services/user"
)

func main() {
	svc := mikros.NewService(&options.NewServiceOptions{
		Service: map[string]options.ServiceOptions{
			"grpc": &options.GrpcServiceOptions{
				ProtoServiceDescription: &userpb.UserService_ServiceDesc,
			},
		},
	})

	svc.Start(&service{})
}
