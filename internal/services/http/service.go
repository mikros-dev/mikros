package http

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"slices"
	"strings"

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

type Server struct {
	port     service.ServerPort
	listener net.Listener
	server   *http.Server
	defs     *Definitions
}

func New() *Server {
	return &Server{}
}

func (s *Server) Name() string {
	return definition.ServiceType_HTTP.String()
}

func (s *Server) Info() []logger_api.Attribute {
	return []logger_api.Attribute{
		logger.String("service.address", fmt.Sprintf(":%v", s.port.Int32())),
		logger.String("service.mode", definition.ServiceType_HTTP.String()),
		logger.String("service.http_auth", fmt.Sprintf("%t", !s.defs.DisableAuth)),
	}
}

func (s *Server) Initialize(ctx context.Context, opt *plugin.ServiceOptions) error {
	provider, ok := opt.ServiceHandler.(http_api.HttpAPI)
	if !ok {
		return errors.New("invalid service handler, it does not implement http_api.HttpAPI")
	}

	baseHandler, err := provider.HTTPHandler(ctx)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", opt.Port))
	if err != nil {
		return fmt.Errorf("could not listen to service port: %w", err)
	}

	svcOptions, ok := opt.Service.(*options.HttpServiceOptions)
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

func corsMiddleware(ch behavior.CorsHandler) func(http.Handler) http.Handler {
	cfg := ch.Cors()

	var (
		allowedOrigins = make(map[string]struct{}, len(cfg.AllowedOrigins))
		allowAll       = false
		allowMethods   = strings.Join(cfg.AllowedMethods, ",")
		allowHeaders   = strings.Join(cfg.AllowedHeaders, ",")
	)

	for _, o := range cfg.AllowedOrigins {
		allowedOrigins[o] = struct{}{}
	}

	if _, ok := allowedOrigins["*"]; ok {
		allowAll = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin == "" {
				// Not a CORS request; pass through.
				next.ServeHTTP(w, r)
				return
			}

			w.Header().Add("Vary", "Origin")

			if allowAll && !cfg.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}
			if !allowAll || cfg.AllowCredentials {
				// Reflect only if the specific origin is allowed.
				if _, ok := allowedOrigins[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
				}
			}

			// The credentials' flag is only meaningful with a specific origin (not "*").
			if cfg.AllowCredentials && w.Header().Get("Access-Control-Allow-Origin") == origin {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Is this a preflight request?
			isPreflight := r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != ""
			if !isPreflight {
				next.ServeHTTP(w, r)
				return
			}

			// Preflight: add appropriate Vary headers.
			w.Header().Add("Vary", "Access-Control-Request-Method")
			w.Header().Add("Vary", "Access-Control-Request-Headers")

			if allowMethods != "" {
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
			}

			if allowHeaders != "" {
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			}
			if allowHeaders == "" {
				// If not configured, echo back the request’s ask.
				reqHeaders := r.Header.Get("Access-Control-Request-Headers")
				if reqHeaders != "" {
					w.Header().Set("Access-Control-Allow-Headers", reqHeaders)
				}
			}

			// Set cache preflight result if configured.
			if cfg.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", cfg.MaxAge))
			}

			// If the Origin wasn't allowed, we didn't set Allow-Origin. The
			// browser will block, which is the correct signal for bad config.
			w.WriteHeader(http.StatusNoContent)
		})
	}
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
	c, err := opt.Features.Feature(options.HttpAuthFeatureName)
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
	c, err := opt.Features.Feature(options.HttpCorsFeatureName)
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

func (s *Server) Run(_ context.Context, _ interface{}) error {
	if err := s.server.Serve(s.listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}

		return err
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
