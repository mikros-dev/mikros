package main

import (
	"github.com/mikros-dev/mikros"
	"github.com/mikros-dev/mikros/components/options"
)

func main() {
	svc := mikros.NewService(&options.NewServiceOptions{
		Service: map[string]options.ServiceOptions{
			"worker": &options.WorkerServiceOptions{},
		},
	})

	svc.Start(&service{})
}
