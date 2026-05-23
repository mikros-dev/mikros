package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/go-playground/validator/v10"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/definition"
	merrors "github.com/mikros-dev/mikros/components/errors"
	mierrors "github.com/mikros-dev/mikros/internal/components/errors"
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
	errors           errors_api.Errors
	logger           logger_api.API
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

	s.logger = opt.Logger
	s.errors = opt.Errors
	s.listener = listener
	s.protoServiceDesc = svc.ProtoServiceDescription
	s.port = opt.Port

	// Starts the gRPC server
	s.server = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.handleGRPCError,
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(s.recoverFromGrpcPanic),
			),
		),
	)

	healthSrv := health.NewServer()
	healthSrv.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s.server, healthSrv)
	s.health = healthSrv

	return nil
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

func (s *Server) recoverFromGrpcPanic(p interface{}) error {
	return s.errors.Internal(fmt.Errorf("%v", p))
}

func (s *Server) handleGRPCError(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	resp, err := handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	// Logs the RPC error for the current application.
	if e, ok := merrors.From(err); ok {
		fields := []logger_api.Attribute{
			logger.String("grpc.method", info.FullMethod),
			logger.String("error.kind", e.Kind().String()),
		}
		if e.Cause() != nil {
			fields = append(fields, logger.String("error.message", e.Cause().Error()))
		}

		s.logger.Error(ctx, e.Message(), append(fields, e.Attributes()...)...)
	}

	// Try to convert the error to a gRPC status.
	st, ok, err := mierrors.ToGRPCStatus(err)
	if ok {
		if err == nil {
			return resp, st.Err()
		}

		s.logger.Error(ctx, "failed to encode gRPC error", logger.Error(err))
		return resp, status.Error(codes.Internal, "internal server error")
	}

	// If is not our error, return it as is.
	if _, ok := status.FromError(err); ok {
		return resp, err
	}

	s.logger.Error(ctx, "unhandled gRPC error",
		logger.String("grpc.method", info.FullMethod),
		logger.Error(err),
	)

	return resp, status.Error(codes.Internal, "internal server error")
}

// Stop stops the gRPC server.
func (s *Server) Stop(_ context.Context) error {
	// Nothing to do here
	return nil
}
