package script

import (
	"context"
	"errors"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/apis/runtimes/script"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Server represents the script runtime server.
type Server struct {
	svc    script.API
	ctx    context.Context
	cancel context.CancelFunc
}

// New creates a new Server struct.
func New() *Server {
	return &Server{}
}

// Name gives the implementation runtime name.
func (s *Server) Name() string {
	return definition.RuntimeTypeScript.String()
}

// Initialize initializes the runtime internals.
func (s *Server) Initialize(ctx context.Context, _ *plugin.RuntimeOptions) error {
	cctx, cancel := context.WithCancel(ctx)

	s.ctx = cctx
	s.cancel = cancel

	return nil
}

// Info returns the runtime info to be logged.
func (s *Server) Info() []logger_api.Attribute {
	return nil
}

// Run starts the script server.
func (s *Server) Run(_ context.Context, srv interface{}) error {
	svc, ok := srv.(script.API)
	if !ok {
		return errors.New("server object does not implement the script.API interface")
	}

	// Holds a reference to the runtime, so we can stop it later.
	s.svc = svc

	// And put it to run.
	return svc.Run(s.ctx)
}

// Stop stops the script server.
func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	return s.svc.Cleanup(ctx)
}
