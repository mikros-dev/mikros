package main

import (
	"github.com/somatech1/mikros"
	"github.com/somatech1/mikros/components/options"
	"github.com/somatech1/mikros/protobuf-examples/gen/go/services/user"
)

func main() {
	s := &service{}
	svc := mikros.NewService(&options.NewServiceOptions{
		Service: map[string]options.ServiceOptions{
			"http": &options.HttpServiceOptions{
				ProtoHttpServer: &routes{s},
			},
		},
		GrpcClients: map[string]*options.GrpcClient{
			"user": {
				ServiceName:       mikros.ServiceName("user"),
				NewClientFunction: user.NewUserServiceClient,
			},
		},
	})

	svc.Start(s)
}
