package http_spec

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/fasthttp/router"
	"github.com/go-playground/validator/v10"
	"github.com/lab259/cors"
	"github.com/valyala/fasthttp"

	"github.com/mikros-dev/mikros/apis/behavior"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
)

type Server struct {
	port              service.ServerPort
	trackerHeaderName string
	defs              *Definitions
	server            *fasthttp.Server
	listener          net.Listener
	logger            logger_api.LoggerAPI
	tracing           behavior.Tracer
	tracker           behavior.Tracker
	panicRecovery     behavior.HTTPSpecRecovery
}

func New() *Server {
	return &Server{}
}

func (s *Server) Name() string {
	return definition.ServiceType_HTTPSpec.String()
}

func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("service.address", fmt.Sprintf(":%v", s.port.Int32())),
		logger.String("service.mode", definition.ServiceType_HTTPSpec.String()),
		logger.String("service.http_auth", fmt.Sprintf("%t", !s.defs.DisableAuth)),
	}
}

func (s *Server) Run(_ context.Context, _ interface{}) error {
	return s.server.Serve(s.listener)
}

func (s *Server) Stop(_ context.Context) error {
	return s.server.Shutdown()
}

func (s *Server) Initialize(ctx context.Context, opt *plugin.ServiceOptions) error {
	if err := s.validate(opt); err != nil {
		return err
	}

	// Initialize specific service definitions
	s.defs = newDefinitions(opt.Definitions)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
	}

	if err := s.initializeHttpServerInternals(ctx, opt); err != nil {
		return err
	}

	s.listener = listener
	s.port = opt.Port
	s.logger = opt.Logger
	s.tracing = s.getTracing(opt)
	s.tracker = s.getTracker(opt)
	s.trackerHeaderName = opt.Env.TrackerHeaderName()

	s.panicRecovery = s.getPanicRecovery(opt)

	return nil
}

func (s *Server) validate(opt *plugin.ServiceOptions) error {
	var (
		validate = validator.New()
		fields   = []interface{}{
			opt.Name,
			opt.Logger,
			opt.Port,
			opt.Env.DeploymentEnv(),
			opt.Service,
			opt.Features,
		}
	)

	for _, f := range fields {
		if err := validate.Var(f, "required"); err != nil {
			return err
		}
	}

	return nil
}

// initializeHttpServerInternals is responsible for setting the HTTP server
// initializing its routes, authentication, CORS and everything, letting it
// in a position to be only started (put in execution) later.
func (s *Server) initializeHttpServerInternals(ctx context.Context, opt *plugin.ServiceOptions) error {
	// Disables this router auto fix-path feature to return a proper
	// 404 when some client uses a wrong endpoint.
	httpRouter := router.New()
	httpRouter.RedirectFixedPath = false

	svc, ok := opt.Service.(*options.HttpSpecServiceOptions)
	if !ok {
		return errors.New("unsupported ServiceOptions received on initialization")
	}

	handlers, err := s.createAuthHandlers(ctx, opt)
	if err != nil {
		return err
	}

	if err = svc.ProtoHttpServer.SetupServer(
		opt.Definitions.ServiceName().String(),
		opt.Logger,
		httpRouter,
		opt.ServiceHandler,
		handlers,
	); err != nil {
		return err
	}

	s.registerHttpServer(httpRouter.Handler, opt)
	if s.server == nil {
		return fmt.Errorf("could not initialize HTTP server without registering a handler first")
	}

	return nil
}

func (s *Server) createAuthHandlers(ctx context.Context, opt *plugin.ServiceOptions) (func(ctx context.Context, handlers map[string]interface{}) error, error) {
	var (
		testMode   = opt.Env.DeploymentEnv() == definition.ServiceDeploy_Test
		auth       = !s.defs.DisableAuth
		authPlugin = s.getAuth(opt)
	)

	// If we're running tests, we won't have authenticated endpoints
	if testMode {
		return nil, nil
	}

	if !auth || authPlugin == nil {
		return nil, nil
	}

	opt.Logger.Info(ctx, "using authenticated HTTP endpoints")
	return authPlugin.AuthHandlers()
}

func (s *Server) getAuth(opt *plugin.ServiceOptions) behavior.HTTPSpecAuthenticator {
	c, err := opt.Features.Feature(options.HttpSpecAuthFeatureName)
	if err != nil {
		return nil
	}

	api, ok := c.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	return api.FrameworkAPI().(behavior.HTTPSpecAuthenticator)
}

// registerHttpServer binds the HTTP handler into the service. It expects that
// all routes have already been initialized.
func (s *Server) registerHttpServer(handler fasthttp.RequestHandler, opt *plugin.ServiceOptions) {
	handler = s.serverRequestHandler(handler)
	serverCors := s.getCors(opt)

	if serverCors != nil {
		handler = cors.New(serverCors.Cors()).Handler(handler)
	}

	s.server = &fasthttp.Server{
		NoDefaultServerHeader: true,
		Handler:               handler,
		ErrorHandler:          s.handleHTTPError,
		ReadTimeout:           60 * time.Second,
		WriteTimeout:          60 * time.Second,
		ReadBufferSize:        64 * 1024,
		WriteBufferSize:       64 * 1024,
		MaxRequestBodySize:    s.defs.MaxRequestBodySize * 1024 * 1024,
	}
}

func (s *Server) getPanicRecovery(opt *plugin.ServiceOptions) behavior.HTTPSpecRecovery {
	if s.defs.DisablePanicRecovery {
		return nil
	}

	c, err := opt.Features.Feature(options.PanicRecoveryFeatureName)
	if err != nil {
		return nil
	}

	api, ok := c.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	return api.FrameworkAPI().(behavior.HTTPSpecRecovery)
}

func (s *Server) getCors(opt *plugin.ServiceOptions) behavior.CorsHandler {
	c, err := opt.Features.Feature(options.HttpCorsFeatureName)
	if err != nil {
		return nil
	}

	api, ok := c.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	return api.FrameworkAPI().(behavior.CorsHandler)
}

func (s *Server) serverRequestHandler(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if s.tracker != nil {
			trackId := s.tracker.Generate()

			// Set the track ID in the current context
			s.tracker.Add(ctx, trackId)

			// Set on the response header the request ID
			ctx.Response.Header.Set(s.trackerHeaderName, trackId)
		}

		if ctx.IsGet() && string(ctx.Path()) == "/health" {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		var data interface{}
		if s.tracing != nil {
			d, err := s.tracing.StartMeasurements(ctx, s.Name())
			if err != nil {
				s.logger.Error(ctx, "tracing begin failed", logger.Error(err))
			}
			data = d
		}

		if s.panicRecovery != nil {
			defer s.panicRecovery.Recover(ctx)
		}

		h(ctx)

		if s.tracing != nil {
			if err := s.tracing.ComputeMetrics(ctx, s.Name(), data); err != nil {
				s.logger.Error(ctx, "tracing cease failed", logger.Error(err))
			}
		}
	}
}

func (s *Server) handleHTTPError(ctx *fasthttp.RequestCtx, err error) {
	s.logger.Error(ctx, "http error", logger.Error(err))
}

func (s *Server) getTracing(opt *plugin.ServiceOptions) behavior.Tracer {
	t, err := opt.Features.Feature(options.TracingFeatureName)
	if err != nil {
		return nil
	}

	api, ok := t.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	return api.FrameworkAPI().(behavior.Tracer)
}

func (s *Server) getTracker(opt *plugin.ServiceOptions) behavior.Tracker {
	t, err := opt.Features.Feature(options.TrackerFeatureName)
	if err != nil {
		return nil
	}

	api, ok := t.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	return api.FrameworkAPI().(behavior.Tracker)
}
