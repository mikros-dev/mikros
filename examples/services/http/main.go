package main

import (
	"net/http"

	"github.com/mikros-dev/mikros"
	"github.com/mikros-dev/mikros/components/options"
)

func main() {
	srv := &service{}
	svc := mikros.NewService(&options.NewServiceOptions{
		Service: map[string]options.ServiceOptions{
			"http": &options.HTTPServiceOptions{
				BasePath: "/example/v1",
				Middlewares: []func(http.Handler) http.Handler{
					srv.loggingMiddleware,
				},
			},
		},
	})

	svc.Start(srv)
}
