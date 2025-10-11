package main

import (
	"github.com/mikros-dev/mikros"
	"github.com/mikros-dev/mikros/components/options"
)

func main() {
	s := &service{}
	svc := mikros.NewService(&options.NewServiceOptions{
		Service: map[string]options.ServiceOptions{
			"http-spec": &options.HTTPSpecServiceOptions{
				ProtoHTTPServer: &routes{s},
			},
		},
	})

	svc.Start(s)
}
