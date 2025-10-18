package worker

import (
	"context"
	"errors"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/apis/services/worker"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Server represents the worker service server.
type Server struct {
	svc    worker.API
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Server struct.
func New() *Server {
	return &Server{}
}

// Name gives the implementation service name.
func (s *Server) Name() string {
	return definition.ServiceTypeWorker.String()
}

// Initialize initializes the service internals.
func (s *Server) Initialize(ctx context.Context, _ *plugin.ServiceOptions) error {
	cctx, cancel := context.WithCancel(ctx)

	s.ctx = cctx
	s.cancel = cancel

	return nil
}

// Info returns the service info to be logged.
func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("service.mode", definition.ServiceTypeWorker.String()),
	}
}

// Run starts the worker server.
func (s *Server) Run(_ context.Context, srv interface{}) error {
	svc, ok := srv.(worker.API)
	if !ok {
		return errors.New("server object does not implement the API interface")
	}

	// Holds a reference to the service, so we can stop it later.
	s.svc = svc

	// And put it to run.
	return svc.Start(s.ctx)
}

// Stop stops the worker server.
func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	return s.svc.Stop(ctx)
}
