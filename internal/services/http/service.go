package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"

	"github.com/lab259/cors"
	"github.com/mikros-dev/mikros/apis/behavior"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	http_api "github.com/mikros-dev/mikros/apis/services/http"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
)

type middleware = func(http.Handler) http.Handler

// Server represents the HTTP service server.
type Server struct {
	port     service.ServerPort
	listener net.Listener
	server   *http.Server
	defs     *Definitions
}

// New creates a new Server struct.
func New() *Server {
	return &Server{}
}

// Name gives the implementation service name.
func (s *Server) Name() string {
	return definition.ServiceTypeHTTP.String()
}

// Info returns service fields to be logged.
func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("service.address", fmt.Sprintf(":%v", s.port.Int32())),
		logger.String("service.mode", definition.ServiceTypeHTTP.String()),
		logger.String("service.http_auth", fmt.Sprintf("%t", !s.defs.DisableAuth)),
	}
}

// Initialize initializes the service internals.
func (s *Server) Initialize(ctx context.Context, opt *plugin.ServiceOptions) error {
	provider, ok := opt.ServiceHandler.(http_api.API)
	if !ok {
		return errors.New("invalid service handler, it does not implement http_api.API")
	}

	baseHandler, err := provider.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
	}

	svcOptions, ok := opt.Service.(*options.HTTPServiceOptions)
	if !ok {
		return errors.New("unsupported ServiceOptions received on initialization")
	}
	if svcOptions == nil {
		return errors.New("invalid ServiceOptions received on initialization")
	}

	var (
		h    = baseHandler
		defs = newDefinitions(opt.Definitions, svcOptions)
	)

	if defs.BasePath != "" {
		h = http.StripPrefix(defs.BasePath, h)
	}

	// Add user supplied middlewares after core ones.
	core, err := buildCoreMiddlewares(ctx, opt, defs)
	if err != nil {
		return err
	}
	chain := append(core, svcOptions.Middlewares...)

	// Compose the handlers
	for i := len(chain) - 1; i >= 0; i-- {
		h = chain[i](h)
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

func buildCoreMiddlewares(ctx context.Context, opt *plugin.ServiceOptions, defs *Definitions) ([]middleware, error) {
	var chain []middleware

	if cors := getCors(opt); cors != nil {
		err := validateCORS(cors)
		if err != nil {
			if defs.CORSStrict {
				return nil, fmt.Errorf("invalid cors options: %w", err)
			}

			opt.Logger.Warn(ctx, "invalid cors options: cors is disabled", logger.Error(err))
		}
		if err == nil {
			chain = append(chain, corsMiddleware(cors))
		}
	}

	if !defs.DisableAuth {
		if auth := getAuth(opt); auth != nil {
			chain = append(chain, func(handler http.Handler) http.Handler {
				return http.HandlerFunc(auth.Handler)
			})
		}
	}

	return chain, nil
}

type corsConfig struct {
	allowedOrigins map[string]struct{}
	allowAll       bool
	allowMethods   string
	allowHeaders   string
}

func corsMiddleware(ch behavior.CorsHandler) func(http.Handler) http.Handler {
	cfg := ch.Cors()
	c := buildConfig(cfg)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")
			setAllowOrigin(w, origin, c, cfg)
			setCredentials(w, origin, cfg)

			if !isPreflight(r) {
				next.ServeHTTP(w, r)
				return
			}

			handlePreflight(w, r, c, cfg)
		})
	}
}

func buildConfig(cfg cors.Options) corsConfig {
	origins := make(map[string]struct{}, len(cfg.AllowedOrigins))
	for _, o := range cfg.AllowedOrigins {
		origins[o] = struct{}{}
	}
	_, allowAll := origins["*"]

	return corsConfig{
		allowedOrigins: origins,
		allowAll:       allowAll,
		allowMethods:   strings.Join(cfg.AllowedMethods, ","),
		allowHeaders:   strings.Join(cfg.AllowedHeaders, ","),
	}
}

func setAllowOrigin(w http.ResponseWriter, origin string, c corsConfig, cfg cors.Options) {
	if c.allowAll && !cfg.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		return
	}
	if _, ok := c.allowedOrigins[origin]; ok {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}
}

func setCredentials(w http.ResponseWriter, origin string, cfg cors.Options) {
	if cfg.AllowCredentials && w.Header().Get("Access-Control-Allow-Origin") == origin {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}
}

func handlePreflight(w http.ResponseWriter, r *http.Request, c corsConfig, cfg cors.Options) {
	w.Header().Add("Vary", "Access-Control-Request-Method")
	w.Header().Add("Vary", "Access-Control-Request-Headers")

	if c.allowMethods != "" {
		w.Header().Set("Access-Control-Allow-Methods", c.allowMethods)
	}

	if c.allowHeaders != "" {
		w.Header().Set("Access-Control-Allow-Headers", c.allowHeaders)
	}

	if c.allowHeaders == "" {
		reqHeaders := r.Header.Get("Access-Control-Request-Headers")
		if reqHeaders != "" {
			w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
		}
	}

	if cfg.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
	}

	w.WriteHeader(http.StatusNoContent)
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
}

func validateCORS(cors behavior.CorsHandler) error {
	cfg := cors.Cors()

	if len(cfg.AllowedOrigins) == 0 {
		return errors.New("allowed origins must not be empty")
	}

	if slices.Contains(cfg.AllowedOrigins, "*") && cfg.AllowCredentials {
		return errors.New(`allowed origins contains "*" but allow credentials is true`)
	}

	if len(cfg.AllowedMethods) == 0 {
		return errors.New("allowed methods must not be empty")
	}

	return nil
}

func getAuth(opt *plugin.ServiceOptions) behavior.HTTPAuthenticator {
	c, err := opt.Features.Feature(options.HTTPAuthFeatureName)
	if err != nil {
		return nil
	}

	api, ok := c.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	auth, ok := api.FrameworkAPI().(behavior.HTTPAuthenticator)
	if !ok {
		return nil
	}

	return auth
}

func getCors(opt *plugin.ServiceOptions) behavior.CorsHandler {
	c, err := opt.Features.Feature(options.HTTPCorsFeatureName)
	if err != nil {
		return nil
	}

	api, ok := c.(plugin.FeatureInternalAPI)
	if !ok {
		return nil
	}

	cors, ok := api.FrameworkAPI().(behavior.CorsHandler)
	if !ok {
		return nil
	}

	return cors
}

// Run runs the service.
func (s *Server) Run(_ context.Context, _ interface{}) error {
	if err := s.server.Serve(s.listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}

	return nil
}

// Stop stops the service.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
