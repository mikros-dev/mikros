//revive:disable:var-naming
package http_spec

//revive:enable:var-naming

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

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/apis/integrations"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
)

// Server represents the HTTP (spec) runtime server.
type Server struct {
	port              service.ServerPort
	trackerHeaderName string
	defs              *Definitions
	server            *fasthttp.Server
	listener          net.Listener
	logger            logger_api.API
	tracing           integrations.Tracer
	tracker           integrations.Tracker
	panicRecovery     integrations.HTTPSpecRecovery
}

// New creates a new Server struct.
func New() *Server {
	return &Server{}
}

// Name gives the implementation runtime name.
func (s *Server) Name() string {
	return definition.RuntimeTypeHTTPSpec.String()
}

// Info returns runtime fields to be logged.
func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("http_spec.listening_address", fmt.Sprintf(":%v", s.port.Int32())),
		logger.String("http_spec.auth_enabled", fmt.Sprintf("%t", !s.defs.DisableAuth)),
	}
}

// Run starts the HTTP (spec) server.
func (s *Server) Run(_ context.Context, _ interface{}) error {
	return s.server.Serve(s.listener)
}

// Stop stops the HTTP (spec) server.
func (s *Server) Stop(_ context.Context) error {
	return s.server.Shutdown()
}

// Initialize initializes the HTTP (spec) server internals.
func (s *Server) Initialize(ctx context.Context, opt *plugin.RuntimeOptions) error {
	if err := s.validate(opt); err != nil {
		return err
	}

	// Initialize specific runtime definitions
	s.defs = newDefinitions(opt.Definitions)

	if err := s.initializeHTTPServerInternals(ctx, opt); err != nil {
		return err
	}

	s.port = opt.Port
	s.logger = opt.Logger
	s.trackerHeaderName = opt.Env.TrackerHeaderName()

	tr, err := s.getTracker(opt)
	if err != nil {
		return err
	}
	s.tracker = tr

	t, err := s.getTracing(opt)
	if err != nil {
		return err
	}
	s.tracing = t

	p, err := s.getPanicRecovery(opt)
	if err != nil {
		return err
	}
	s.panicRecovery = p

	// Starts the listener last so we don't need to worry about closing it in
	// other error paths.
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
	}
	s.listener = listener

	return nil
}

func (s *Server) validate(opt *plugin.RuntimeOptions) error {
	var (
		validate = validator.New()
		fields   = []interface{}{
			opt.Name,
			opt.Logger,
			opt.Port,
			opt.Env.DeploymentEnv(),
			opt.ServiceOptions,
			opt.Integrations,
		}
	)

	for _, f := range fields {
		if err := validate.Var(f, "required"); err != nil {
			return err
		}
	}

	return nil
}

// initializeHTTPServerInternals is responsible for setting the HTTP server,
// initializing its routes, authentication, CORS, and everything, letting it
// in a position to be only started (put in execution) later.
func (s *Server) initializeHTTPServerInternals(ctx context.Context, opt *plugin.RuntimeOptions) error {
	// Disables this router auto fix-path feature to return a proper
	// 404 when some client uses a wrong endpoint.
	httpRouter := router.New()
	httpRouter.RedirectFixedPath = false

	svc, ok := opt.ServiceOptions.(*options.HTTPSpecServiceOptions)
	if !ok {
		return errors.New("unsupported RuntimeOptions received on initialization")
	}

	handlers, err := s.createAuthHandlers(ctx, opt)
	if err != nil {
		return err
	}

	if err = svc.ProtoHTTPServer.SetupServer(
		opt.Definitions.ServiceName().String(),
		opt.Logger,
		httpRouter,
		opt.ServiceHandler,
		handlers,
	); err != nil {
		return err
	}

	if err := s.registerHTTPServer(httpRouter.Handler, opt); err != nil {
		return err
	}
	if s.server == nil {
		return fmt.Errorf("could not initialize HTTP server without registering a handler first")
	}

	return nil
}

func (s *Server) createAuthHandlers(
	ctx context.Context,
	opt *plugin.RuntimeOptions,
) (func(ctx context.Context, handlers map[string]interface{}) error, error) {
	// If we're running tests, we won't have authenticated endpoints
	if opt.Env.DeploymentEnv() == definition.DeploymentEnvTest {
		return nil, nil
	}

	// Also, if running with authentication disabled
	if !s.defs.DisableAuth {
		return nil, nil
	}

	authPlugin, err := s.getAuth(opt)
	if err != nil {
		return nil, err
	}

	opt.Logger.Info(ctx, "using authenticated HTTP endpoints")
	return authPlugin.AuthHandlers()
}

func (s *Server) getAuth(opt *plugin.RuntimeOptions) (integrations.HTTPSpecAuthenticator, error) {
	i, err := opt.Integrations.Integration(options.HTTPSpecAuthIntegrationName)
	if err != nil {
		return nil, errors.New("http auth is enabled but integration is not available")
	}

	a, ok := i.API().(integrations.HTTPSpecAuthenticator)
	if !ok {
		return nil, errors.New("http auth is enabled but integration does not implement HTTPSpecAuthenticator")
	}

	return a, nil
}

// registerHTTPServer binds the HTTP handler into the runtime. It expects that
// all routes have already been initialized.
func (s *Server) registerHTTPServer(handler fasthttp.RequestHandler, opt *plugin.RuntimeOptions) error {
	handler = s.serverRequestHandler(handler)

	serverCors, err := s.getCors(opt)
	if err != nil {
		return err
	}
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

	return nil
}

func (s *Server) getPanicRecovery(opt *plugin.RuntimeOptions) (integrations.HTTPSpecRecovery, error) {
	if s.defs.DisablePanicRecovery {
		return nil, nil
	}

	i, err := opt.Integrations.Integration(options.PanicRecoveryIntegrationName)
	if err != nil {
		return nil, errors.New("panic recovery is enabled but integration is not available")
	}

	p, ok := i.API().(integrations.HTTPSpecRecovery)
	if !ok {
		return nil, errors.New("panic recovery is enabled but integration does not implement HTTPSpecRecovery")
	}

	return p, nil
}

func (s *Server) getCors(opt *plugin.RuntimeOptions) (integrations.CorsHandler, error) {
	i, err := opt.Integrations.Integration(options.HTTPCorsIntegrationName)
	if err != nil {
		return nil, nil
	}

	c, ok := i.API().(integrations.CorsHandler)
	if !ok {
		return nil, errors.New("http cors integration exists but does not implement CorsHandler")
	}

	return c, nil
}

func (s *Server) serverRequestHandler(h fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		if s.tracker != nil {
			s.injectTrackerID(ctx)
		}

		if ctx.IsGet() && string(ctx.Path()) == "/health" {
			ctx.SetStatusCode(fasthttp.StatusOK)
			return
		}

		data := s.startTracing(ctx)
		if s.panicRecovery != nil {
			defer s.panicRecovery.Recover(ctx)
		}

		// Call the handler
		h(ctx)
		s.stopTracing(ctx, data)
	}
}

func (s *Server) injectTrackerID(ctx *fasthttp.RequestCtx) {
	trackID := s.tracker.Generate()

	// Set the track ID in the current context
	s.tracker.Add(ctx, trackID)

	// Set on the response header the request ID
	ctx.Response.Header.Set(s.trackerHeaderName, trackID)
}

func (s *Server) startTracing(ctx *fasthttp.RequestCtx) interface{} {
	var data interface{}

	if s.tracing != nil {
		d, err := s.tracing.StartMeasurements(ctx, s.Name())
		if err != nil {
			s.logger.Error(ctx, "tracing begin failed", logger.Error(err))
		}
		data = d
	}

	return data
}

func (s *Server) stopTracing(ctx *fasthttp.RequestCtx, data interface{}) {
	if s.tracing != nil {
		if err := s.tracing.ComputeMetrics(ctx, s.Name(), data); err != nil {
			s.logger.Error(ctx, "tracing cease failed", logger.Error(err))
		}
	}
}

func (s *Server) handleHTTPError(ctx *fasthttp.RequestCtx, err error) {
	s.logger.Error(ctx, "http error", logger.Error(err))
}

func (s *Server) getTracing(opt *plugin.RuntimeOptions) (integrations.Tracer, error) {
	i, err := opt.Integrations.Integration(options.TracingIntegrationName)
	if err != nil {
		return nil, nil
	}

	t, ok := i.API().(integrations.Tracer)
	if !ok {
		return nil, errors.New("tracing integration exists but does not implement Tracer")
	}

	return t, nil
}

func (s *Server) getTracker(opt *plugin.RuntimeOptions) (integrations.Tracker, error) {
	i, err := opt.Integrations.Integration(options.TrackerIntegrationName)
	if err != nil {
		return nil, nil
	}

	t, ok := i.API().(integrations.Tracker)
	if !ok {
		return nil, errors.New("tracker integration exists but does not implement Tracker")
	}

	return t, nil
}
