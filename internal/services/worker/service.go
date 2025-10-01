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

type Server struct {
	svc    worker.WorkerAPI
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Server {
	return &Server{}
}

func (s *Server) Name() string {
	return definition.ServiceType_Worker.String()
}

func (s *Server) Initialize(ctx context.Context, _ *plugin.ServiceOptions) error {
	cctx, cancel := context.WithCancel(ctx)

	s.ctx = cctx
	s.cancel = cancel

	return nil
}

func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("service.mode", definition.ServiceType_Worker.String()),
	}
}

func (s *Server) Run(_ context.Context, srv interface{}) error {
	svc, ok := srv.(worker.WorkerAPI)
	if !ok {
		return errors.New("server object does not implement the WorkerAPI interface")
	}

	// Holds a reference to the service, so we can stop it later.
	s.svc = svc

	// And put it to run.
	return svc.Start(s.ctx)
}

func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	return s.svc.Stop(ctx)
}
