package main

import (
	"context"
	"net/http"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
)

type service struct {
	Logger logger_api.LoggerAPI `mikros:"feature"`
}

func (s *service) HTTPHandler(_ context.Context) (http.Handler, error) {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /items", s.listItems)
	mux.HandleFunc("POST /items", s.createItem)
	mux.HandleFunc("GET /items/{id}", s.getItem)
	mux.HandleFunc("DELETE /items/{id}", s.deleteItem)

	return mux, nil
}
