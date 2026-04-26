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

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
)

// Server represents the gRPC runtime server.
type Server struct {
	port             service.ServerPort
	server           *grpc.Server
	listener         net.Listener
	health           *health.Server
	errors           errors_api.ErrorAPI
	protoServiceDesc *grpc.ServiceDesc
}

// New creates a new Server struct.
func New() *Server {
	return &Server{}
}

// Name gives the implementation runtime name.
func (s *Server) Name() string {
	return definition.RuntimeTypeGRPC.String()
}

// Info returns runtime fields to be logged.
func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("grpc.listening_address", fmt.Sprintf(":%v", s.port.Int32())),
	}
}

// Run starts the gRPC server.
func (s *Server) Run(_ context.Context, srv interface{}) error {
	s.server.RegisterService(s.protoServiceDesc, srv)
	reflection.Register(s.server)
	return s.server.Serve(s.listener)
}

// Initialize initializes the gRPC server internals.
func (s *Server) Initialize(_ context.Context, opt *plugin.RuntimeOptions) error {
	if err := s.validate(opt); err != nil {
		return err
	}

	svc, ok := opt.ServiceOptions.(*options.GrpcServiceOptions)
	if !ok {
		return errors.New("unsupported RuntimeOptions received on initialization")
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
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

func (s *Server) validate(opt *plugin.RuntimeOptions) error {
	var (
		validate = validator.New()
		fields   = []interface{}{
			opt.ServiceOptions,
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

// Stop stops the gRPC server.
func (s *Server) Stop(_ context.Context) error {
	// Nothing to do here
	return nil
}
