package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/go-playground/validator/v10"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	errorsApi "github.com/mikros-dev/mikros/apis/errors"
	loggerApi "github.com/mikros-dev/mikros/apis/logger"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
)

type Server struct {
	port             service.ServerPort
	server           *grpc.Server
	listener         net.Listener
	health           *health.Server
	errors           errorsApi.ErrorFactory
	protoServiceDesc *grpc.ServiceDesc
}

func New() *Server {
	return &Server{}
}

func (s *Server) Name() string {
	return definition.ServiceType_gRPC.String()
}

func (s *Server) Info() []loggerApi.Attribute {
	return []loggerApi.Attribute{
		logger.String("service.address", fmt.Sprintf(":%v", s.port.Int32())),
		logger.String("service.mode", definition.ServiceType_gRPC.String()),
	}
}

func (s *Server) Run(_ context.Context, srv interface{}) error {
	s.server.RegisterService(s.protoServiceDesc, srv)
	reflection.Register(s.server)

	if err := s.server.Serve(s.listener); err != nil {
		return err
	}

	return nil
}

func (s *Server) Initialize(_ context.Context, opt *plugin.ServiceOptions) error {
	if err := s.validate(opt); err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
	}

	svc, ok := opt.Service.(*options.GrpcServiceOptions)
	if !ok {
		return errors.New("unsupported ServiceOptions received on initialization")
	}

	s.errors = opt.Errors
	s.listener = listener
	s.protoServiceDesc = svc.ProtoServiceDescription
	s.port = opt.Port

	// Starts the gRPC server
	s.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(
					grpc_recovery.WithRecoveryHandlerContext(s.recoverFromGrpcPanic),
				),
			),
		),
	)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s.server, healthSrv)
	s.health = healthSrv

	return nil
}

func (s *Server) recoverFromGrpcPanic(ctx context.Context, p interface{}) error {
	return s.errors.Internal(fmt.Errorf("%v", p)).Submit(ctx)
}

func (s *Server) validate(opt *plugin.ServiceOptions) error {
	var (
		validate = validator.New()
		fields   = []interface{}{
			opt.Service,
			opt.Logger,
			opt.Errors,
			opt.Port,
		}
	)

	for _, f := range fields {
		if err := validate.Var(f, "required"); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Stop(_ context.Context) error {
	// Nothing to do here
	return nil
}
