package script

import (
	"context"
	"errors"

	loggerApi "github.com/mikros-dev/mikros/apis/logger"
	"github.com/mikros-dev/mikros/apis/services/script"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Server struct {
	svc    script.ServiceAPI
	ctx    context.Context
	cancel context.CancelFunc
}

func New() *Server {
	return &Server{}
}

func (s *Server) Name() string {
	return definition.ServiceType_Script.String()
}

func (s *Server) Initialize(ctx context.Context, _ *plugin.ServiceOptions) error {
	cctx, cancel := context.WithCancel(ctx)

	s.ctx = cctx
	s.cancel = cancel

	return nil
}

func (s *Server) Info() []loggerApi.Attribute {
	return []loggerApi.Attribute{
		logger.String("service.mode", definition.ServiceType_Script.String()),
	}
}

func (s *Server) Run(_ context.Context, srv interface{}) error {
	svc, ok := srv.(script.ServiceAPI)
	if !ok {
		return errors.New("server object does not implement the script.ServiceAPI interface")
	}

	// Holds a reference to the service, so we can stop it later.
	s.svc = svc

	// And put it to run.
	return svc.Run(s.ctx)
}

func (s *Server) Stop(ctx context.Context) error {
	s.cancel()
	return s.svc.Cleanup(ctx)
}
