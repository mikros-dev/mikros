package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	http_api "github.com/mikros-dev/mikros/apis/services/http"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
)

type Server struct {
	port     service.ServerPort
	listener net.Listener
	server   *http.Server
	defs     *Definitions
}

func New() *Server {
	return &Server{}
}

func (s *Server) Name() string {
	return definition.ServiceType_HTTP.String()
}

func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("service.address", fmt.Sprintf(":%v", s.port.Int32())),
		logger.String("service.mode", definition.ServiceType_HTTP.String()),
		logger.String("service.http_auth", fmt.Sprintf("%t", !s.defs.DisableAuth)),
	}
}

func (s *Server) Initialize(ctx context.Context, opt *plugin.ServiceOptions) error {
	provider, ok := opt.ServiceHandler.(http_api.HttpAPI)
	if !ok {
		return errors.New("invalid service handler, it does not implement http_api.HttpAPI")
	}

	baseHandler, err := provider.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
	}

	svcOptions, ok := opt.Service.(*options.HttpServiceOptions)
	if !ok {
		return errors.New("unsupported ServiceOptions received on initialization")
	}

	var (
		h     = baseHandler
		defs  = newDefinitions(opt.Definitions, svcOptions)
		chain []func(http.Handler) http.Handler
	)

	if !defs.DisableAuth {
		// TODO
		// add auth middleware into chain
	}

	// User supplied middlewares
	chain = append(chain, svcOptions.Middlewares...)

	// Compose the handlers
	for i := len(chain) - 1; i >= 0; i-- {
		h = chain[i](h)
	}

	if bp := strings.TrimSuffix(defs.BasePath, "/"); bp != "" {
		h = http.StripPrefix(bp, h)
	}

	// Initialize the service
	s.defs = defs
	s.port = opt.Port
	s.listener = listener
	s.server = &http.Server{
		Handler:        h,
		ReadTimeout:    defs.ReadTimeout,
		WriteTimeout:   defs.WriteTimeout,
		IdleTimeout:    defs.IdleTimeout,
		MaxHeaderBytes: defs.MaxHeaderBytes,
	}

	return nil
}

func (s *Server) Run(_ context.Context, _ interface{}) error {
	return s.server.Serve(s.listener)
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
