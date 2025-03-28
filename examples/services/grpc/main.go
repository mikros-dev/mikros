package main

import (
	"github.com/somatech1/mikros"
	"github.com/somatech1/mikros/components/options"
	"github.com/somatech1/mikros/protobuf-examples/gen/go/services/user"
)

func main() {
	svc := mikros.NewService(&options.NewServiceOptions{
		Service: map[string]options.ServiceOptions{
			"grpc": &options.GrpcServiceOptions{
				ProtoServiceDescription: &user.UserService_ServiceDesc,
			},
		},
	})

	svc.Start(&service{})
}
