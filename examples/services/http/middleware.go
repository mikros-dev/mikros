package main

import (
	"net/http"
	"time"

	"github.com/mikros-dev/mikros/components/logger"
)

// loggingMiddleware is a simple middleware example to hook up in each request
func (s *service) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Call the next handler
		next.ServeHTTP(w, r)

		// Log after response is written
		s.Logger.Info(r.Context(), "endpoint activated",
			logger.String("method", r.Method),
			logger.String("path", r.URL.Path),
			logger.String("query", r.URL.RawQuery),
			logger.Any("time", time.Since(start)))
	})
}
