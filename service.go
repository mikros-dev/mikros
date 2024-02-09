package mikros

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"

	errorsApi "github.com/somatech1/mikros/apis/errors"
	loggerApi "github.com/somatech1/mikros/apis/logger"
	mcontext "github.com/somatech1/mikros/components/context"
	"github.com/somatech1/mikros/components/definition"
	mgrpc "github.com/somatech1/mikros/components/grpc"
	"github.com/somatech1/mikros/components/logger"
	"github.com/somatech1/mikros/components/options"
	"github.com/somatech1/mikros/components/plugin"
	"github.com/somatech1/mikros/components/service"
	"github.com/somatech1/mikros/components/testing"
	merrors "github.com/somatech1/mikros/internal/components/errors"
	"github.com/somatech1/mikros/internal/components/lifecycle"
	mlogger "github.com/somatech1/mikros/internal/components/logger"
	"github.com/somatech1/mikros/internal/components/tags"
	"github.com/somatech1/mikros/internal/components/tracker"
	"github.com/somatech1/mikros/internal/components/validations"
	httpFeature "github.com/somatech1/mikros/internal/features/http"
	"github.com/somatech1/mikros/internal/services/grpc"
	"github.com/somatech1/mikros/internal/services/http"
	"github.com/somatech1/mikros/internal/services/native"
	"github.com/somatech1/mikros/internal/services/script"
)

// Service is the object which represents a service application.
type Service struct {
	serviceToml     string
	serviceOptions  map[string]options.ServiceOptions
	runtimeFeatures map[string]interface{}
	errors          *merrors.Factory
	logger          *mlogger.Logger
	ctx             *mcontext.ServiceContext
	servers         []plugin.Service
	clients         map[string]*options.GrpcClient
	definitions     *definition.Definitions
	envs            *Env
	features        *plugin.FeatureSet
	services        *plugin.ServiceSet
	tracker         *tracker.Tracker
}

// ServiceName is the way to retrieve a service name from a string.
func ServiceName(name string) service.Name {
	return service.FromString(name)
}

// NewService creates a new Service object for building and putting to run
// a new application.
//
// We don't return an error here to force the application to end in case
// something wrong happens.
func NewService(opt *options.NewServiceOptions) *Service {
	if err := opt.Validate(); err != nil {
		log.Fatal(err)
	}

	svc, err := initService(opt)
	if err != nil {
		log.Fatal(err)
	}

	return svc
}

// initService parses the service.toml file and creates the Service object
// initializing its main fields.
func initService(opt *options.NewServiceOptions) (*Service, error) {
	path, err := getServiceTomlPath()
	if err != nil {
		return nil, err
	}

	defs, err := definition.Parse(path)
	if err != nil {
		return nil, err
	}

	// Loads environment variables
	envs, err := loadEnvs(defs)
	if err != nil {
		return nil, err
	}

	// Initialize the service logger system.
	serviceLogger := mlogger.New(mlogger.Options{
		LogOnlyFatalLevel:      envs.DeploymentEnv == definition.ServiceDeploy_Test,
		DisableErrorStacktrace: !defs.Log.ErrorStacktrace,
		FixedAttributes: map[string]string{
			"service.name":    defs.ServiceName().String(),
			"service.type":    defs.ServiceTypesAsString(),
			"service.version": defs.Version,
			"service.env":     envs.DeploymentEnv.String(),
			"service.product": defs.Product,
		},
	})

	if defs.Log.Level != "" {
		if _, err := serviceLogger.SetLogLevel(defs.Log.Level); err != nil {
			return nil, err
		}
	}

	// Context initialization
	ctx, err := mcontext.New(&mcontext.Options{
		Name: defs.ServiceName(),
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		logger:          serviceLogger,
		errors:          initServiceErrors(defs, serviceLogger),
		clients:         opt.GrpcClients,
		envs:            envs,
		definitions:     defs,
		runtimeFeatures: opt.RunTimeFeatures,
		serviceOptions:  opt.Service,
		ctx:             ctx,
		serviceToml:     path,
		features:        registerInternalFeatures(),
		services:        registerInternalServices(),
	}, nil
}

func getServiceTomlPath() (string, error) {
	path := flag.String("config", "", "Sets the alternative path for 'service.toml' file.")
	flag.Parse()

	if path != nil && *path != "" {
		return *path, nil
	}

	serviceDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Join(serviceDir, "service.toml"), nil
}

// loadEnvs loads the framework main environment variables through the env
// feature plugin.
func loadEnvs(defs *definition.Definitions) (*Env, error) {
	return newEnv(defs)
}

func registerInternalFeatures() *plugin.FeatureSet {
	features := plugin.NewFeatureSet()
	features.Register(options.HttpFeatureName, httpFeature.New())

	return features
}

func registerInternalServices() *plugin.ServiceSet {
	services := plugin.NewServiceSet()

	services.Register(grpc.New())
	services.Register(http.New())
	services.Register(native.New())
	services.Register(script.New())

	return services
}

func initServiceErrors(defs *definition.Definitions, log loggerApi.Logger) *merrors.Factory {
	var hideDetails bool
	if defs.IsServiceType(definition.ServiceType_HTTP) {
		hideDetails = defs.HTTP.HideErrorDetails
	}

	return merrors.NewFactory(merrors.FactoryOptions{
		HideMessageDetails: hideDetails,
		ServiceName:        defs.ServiceName().String(),
		Logger:             log,
	})
}

// WithExternalServices allows a service to add external service implementations
// into it.
func (s *Service) WithExternalServices(services *plugin.ServiceSet) *Service {
	s.services.Append(services)
	for name := range services.Services() {
		s.definitions.AddSupportedServiceType(name)
	}

	return s
}

// WithExternalFeatures allows a service to add external features into it, so they
// can be used from it.
func (s *Service) WithExternalFeatures(features *plugin.FeatureSet) *Service {
	s.features.Append(features)
	return s
}

// Start puts the service in execution mode and blocks execution. This function
// should be the last one called by the service.
//
// We don't return an error here so that the service does not need to handle it
// inside its code. We abort in case of an error.
func (s *Service) Start(srv interface{}) {
	ctx, err := s.start(srv)
	if err != nil {
		s.abort(ctx, err)
	}

	// If we're running tests, we end the method here to avoid putting the
	// service in execution.
	if s.DeployEnvironment() == definition.ServiceDeploy_Test {
		return
	}

	s.run(ctx, srv)
}

func (s *Service) start(srv interface{}) (context.Context, *merrors.AbortError) {
	ctx := context.Background()
	s.logger.Info(ctx, "starting service")

	if err := s.validateDefinitions(); err != nil {
		return nil, merrors.NewAbortError("service definitions error", err)
	}

	if err := s.startFeatures(ctx, srv); err != nil {
		return nil, err
	}

	if err := s.startTracker(); err != nil {
		return nil, merrors.NewAbortError("could not initialize the service tracker", err)
	}

	if err := s.setupLoggerExtractor(); err != nil {
		return nil, merrors.NewAbortError("could not set logger extractor", err)
	}

	if err := s.initializeServiceInternals(ctx, srv); err != nil {
		return nil, err
	}

	s.printServiceResources(ctx)
	return ctx, nil
}

// validateDefinitions is responsible for validating the 'service.toml' file
// content.
//
// It also adds all features and services (internal and external) settings into
// the service definitions before validating it.
func (s *Service) validateDefinitions() error {
	iter := s.features.Iterator()
	for p, next := iter.Next(); next; p, next = iter.Next() {
		if cfg, ok := p.(plugin.FeatureSettings); ok {
			defs, err := cfg.Definitions(s.serviceToml)
			if err != nil {
				return err
			}

			s.definitions.AddExternalFeatureDefinitions(p.Name(), defs)
		}
	}

	for _, svc := range s.services.Services() {
		if d, ok := svc.(plugin.ServiceSettings); ok {
			defs, err := d.Definitions(s.serviceToml)
			if err != nil {
				return err
			}

			s.definitions.AddExternalServiceDefinitions(svc.Name(), defs)
		}
	}

	return s.definitions.Validate()
}

// startFeatures starts all registered features and everything that are related
// to them.
func (s *Service) startFeatures(ctx context.Context, srv interface{}) *merrors.AbortError {
	s.logger.Info(ctx, "starting dependent services")

	// Initialize features
	if err := s.initializeFeatures(ctx, srv); err != nil {
		return merrors.NewAbortError("could not initialize features", err)
	}

	return nil
}

func (s *Service) initializeFeatures(ctx context.Context, srv interface{}) error {
	initializeOptions := &plugin.InitializeOptions{
		Logger:          s.logger,
		Errors:          s.errors,
		Definitions:     s.definitions,
		Tags:            s.tags(),
		ServiceContext:  s.ctx,
		RunTimeFeatures: s.runtimeFeatures,
		Env:             s.envs.ToMapEnv(),
	}

	// Initialize features
	if err := s.features.InitializeAll(ctx, initializeOptions); err != nil {
		return err
	}

	// And execute their Start API
	if err := s.features.StartAll(ctx, srv); err != nil {
		return err
	}

	return nil
}

func (s *Service) startTracker() error {
	t, err := tracker.New(s.features)
	if err != nil {
		return err
	}

	s.tracker = t
	return nil
}

func (s *Service) setupLoggerExtractor() error {
	e, err := s.features.Feature(options.LoggerExtractorFeatureName)
	if err != nil && !strings.Contains(err.Error(), "could not find feature") {
		return err
	}

	if api, ok := e.(plugin.FeatureInternalAPI); ok {
		extractor := api.(loggerApi.Extractor)
		s.logger.SetContextFieldExtractor(extractor.Extract)
	}

	return nil
}

func (s *Service) initializeServiceInternals(ctx context.Context, srv interface{}) *merrors.AbortError {
	if err := s.initializeServiceHandler(srv); err != nil {
		return merrors.NewAbortError("invalid service server object", err)
	}

	if err := s.initializeRegisteredServices(ctx, srv); err != nil {
		return merrors.NewAbortError("could not initialize internal services", err)
	}

	// Call lifecycle.OnStart before validating the service structure to
	// allow its fields to be initialized at this point.
	if err := lifecycle.OnStart(srv, ctx); err != nil {
		return merrors.NewAbortError("failed while running lifecycle.OnStart", err)
	}

	if s.envs.DeploymentEnv != definition.ServiceDeploy_Test {
		// Establishes connection with all gRPC clients.
		if err := s.coupleClients(srv); err != nil {
			return merrors.NewAbortError("could not establish connection with clients", err)
		}

		if err := validations.EnsureValuesAreInitialized(srv); err != nil {
			return merrors.NewAbortError("service server object is not properly initialized", err)
		}
	}

	return nil
}

// initializeServiceHandler initializes the service structure ensuring that it
// is framework compatible, i.e., it has at least a *mikros.Service member,
// in order to give access to the framework API through it.
func (s *Service) initializeServiceHandler(srv interface{}) error {
	if err := validations.EnsureStructIsServiceCompatible(srv); err != nil {
		return err
	}

	var (
		typeOf  = reflect.TypeOf(srv)
		valueOf = reflect.ValueOf(srv)
	)

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		typeField := typeOf.Elem().Field(i)

		// Initializes the service *Service member, allowing it having access to
		// the framework API.
		if typeField.Type.String() == "*mikros.Service" {
			ptr := reflect.New(reflect.ValueOf(s).Type())
			ptr.Elem().Set(reflect.ValueOf(s))
			valueOf.Elem().Field(i).Set(ptr.Elem())
		}
	}

	return nil
}

func (s *Service) initializeRegisteredServices(ctx context.Context, srv interface{}) error {
	getServicePort := func(port service.ServerPort, serviceType string) service.ServerPort {
		// Use default port values in case no port was set in the service.toml
		if port == 0 {
			if serviceType == definition.ServiceType_gRPC.String() {
				return service.ServerPort(s.envs.GrpcPort)
			}

			if serviceType == definition.ServiceType_HTTP.String() {
				return service.ServerPort(s.envs.HttpPort)
			}
		}

		return port
	}

	// Creates the service
	for serviceType, servicePort := range s.definitions.ServiceTypes() {
		svc, ok := s.services.Services()[serviceType.String()]
		if !ok {
			return fmt.Errorf("could not find service implementation for '%v", serviceType.String())
		}

		opt, ok := s.serviceOptions[serviceType.String()]
		if !ok {
			return fmt.Errorf("could not find service type '%v' options in initialization", serviceType.String())
		}

		if err := svc.Initialize(ctx, &plugin.ServiceOptions{
			Port:           getServicePort(servicePort, serviceType.String()),
			Type:           serviceType,
			Name:           s.definitions.ServiceName(),
			Product:        s.definitions.Product,
			Logger:         s.logger,
			Errors:         s.errors,
			ServiceContext: s.ctx,
			Tags:           s.tags(),
			Service:        opt,
			Definitions:    s.definitions,
			Features:       s.features,
			ServiceHandler: srv,
			Env:            s.envs.ToMapEnv(),
		}); err != nil {
			return err
		}

		// Saves only the initialized services
		s.servers = append(s.servers, svc)
	}

	return nil
}

// coupleClients establishes connections with all client services that a service
// has as dependency.
func (s *Service) coupleClients(srv interface{}) error {
	// Assures that the service has dependencies.
	if len(s.clients) == 0 {
		return nil
	}

	var (
		typeOf  = reflect.TypeOf(srv)
		valueOf = reflect.ValueOf(srv)
	)

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		typeField := typeOf.Elem().Field(i)
		if tag := tags.ParseTag(typeField.Tag); tag != nil {
			if tag.IsOptional {
				continue
			}

			client, ok := s.clients[tag.GrpcClientName]
			if !ok {
				return fmt.Errorf("could not find gRPC client '%s' inside service options", tag.GrpcClientName)
			}
			if err := client.Validate(); err != nil {
				return err
			}

			serviceTracker, _ := s.tracker.Tracker()

			// For each valid client, establishes their gRPC connection and
			// initializes the service structure properly by pointing its
			// members to these connections.

			cOpts := &mgrpc.ClientConnectionOptions{
				ServiceName: client.ServiceName,
				Context:     s.ctx,
				Connection: mgrpc.ConnectionOptions{
					Namespace: s.envs.CoupledNamespace,
					Port:      s.envs.CoupledPort,
				},
				Tracker: serviceTracker,
			}

			if s.definitions.Clients != nil {
				if opt, ok := s.definitions.Clients[client.ServiceName.String()]; ok {
					cOpts.AlternativeConnection = &mgrpc.ConnectionOptions{
						Host: opt.Host,
						Port: opt.Port,
					}
				}
			}

			conn, err := mgrpc.ClientConnection(cOpts)
			if err != nil {
				return err
			}

			call := reflect.ValueOf(client.NewClientFunction)
			out := call.Call([]reflect.Value{reflect.ValueOf(conn)})

			ptr := reflect.New(out[0].Type())
			ptr.Elem().Set(out[0].Elem())
			valueOf.Elem().Field(i).Set(ptr.Elem())
		}
	}

	return nil
}

func (s *Service) printServiceResources(ctx context.Context) {
	var (
		fields []loggerApi.Attribute
		iter   = s.features.Iterator()
	)

	for f, next := iter.Next(); next; f, next = iter.Next() {
		fields = append(fields, f.Fields()...)
	}

	s.logger.Info(ctx, "service resources", fields...)
}

func (s *Service) run(ctx context.Context, srv interface{}) {
	defer s.stopService(ctx)
	defer lifecycle.OnFinish(srv, ctx)

	// In case we're a script service, only execute its function and terminate
	// the execution.
	if s.definitions.IsServiceType(definition.ServiceType_Script) {
		svc := s.servers[0]
		s.logger.Info(ctx, "service is running", svc.Info()...)

		if err := svc.Run(ctx, srv); err != nil {
			s.abort(ctx, merrors.NewAbortError("fatal error", err))
		}

		return
	}

	// Otherwise, initialize all service types and put them to run.

	// Create channels for finishing the service and bind the signal that
	// finishes it.
	errChan := make(chan error)
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGTERM, syscall.SIGINT)

	for _, svc := range s.servers {
		go func(service plugin.Service) {
			s.logger.Info(ctx, "service is running", service.Info()...)
			if err := service.Run(ctx, srv); err != nil {
				errChan <- err
			}
		}(svc)
	}

	// Blocks the call
	select {
	case err := <-errChan:
		s.abort(ctx, merrors.NewAbortError("fatal error", err))

	case <-stopChan:
	}
}

func (s *Service) stopService(ctx context.Context) {
	s.logger.Info(ctx, "stopping service")

	if err := s.stopDependentServices(ctx); err != nil {
		s.logger.Error(ctx, "could not stop other running services", logger.Error(err))
	}

	for _, svc := range s.servers {
		if err := svc.Stop(ctx); err != nil {
			s.logger.Error(ctx, "could not stop service server",
				append([]loggerApi.Attribute{logger.Error(err)}, svc.Info()...)...)
		}
	}

	s.Logger().Info(ctx, "service stopped")
}

// stopDependentServices stops other services that are running along with the
// main service.
func (s *Service) stopDependentServices(ctx context.Context) error {
	s.logger.Info(ctx, "stopping dependent services")

	if err := s.features.CleanupAll(ctx); err != nil {
		return err
	}

	return nil
}

// Logger gives access to the logger API from inside a service context.
func (s *Service) Logger() loggerApi.Logger {
	return s.logger
}

// Errors gives access to the errors API from inside a service context.
func (s *Service) Errors() errorsApi.ErrorFactory {
	return s.errors
}

// Abort is a helper method to abort services in the right way, when external
// initialization is needed.
func (s *Service) Abort(message string, err error) {
	s.abort(context.TODO(), merrors.NewAbortError(message, err))
}

// abort is an internal helper method to finish the service execution with an
// error message.
func (s *Service) abort(ctx context.Context, err *merrors.AbortError) {
	s.logger.Fatal(ctx, err.Message, logger.Error(err.InnerError))
}

// ServiceName gives back the service name.
func (s *Service) ServiceName() string {
	return s.definitions.ServiceName().String()
}

// DeployEnvironment exposes the current service deploymentEnv environment.
func (s *Service) DeployEnvironment() definition.ServiceDeploy {
	return s.envs.DeploymentEnv
}

// tags gives a map of current service tags to be used with external resources.
func (s *Service) tags() map[string]string {
	serviceType := s.definitions.ServiceTypesAsString()
	if strings.Contains(serviceType, ",") {
		// SQS tags does not accept commas, just unicode letters, digits,
		// whitespace, or one of these symbols: _ . : / = + - @
		serviceType = "hybrid"
	}

	return map[string]string{
		"service.name":    s.ServiceName(),
		"service.type":    serviceType,
		"service.version": s.definitions.Version,
		"service.product": s.definitions.Product,
	}
}

// Feature is the service mechanism to have access to an external feature
// public API.
func (s *Service) Feature(ctx context.Context, target interface{}) error {
	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		return s.Errors().Internal(errors.New("requested target API must be a pointer")).
			Submit(ctx)
	}

	it := s.features.Iterator()
	for {
		feature, next := it.Next()
		if !next {
			break
		}

		f := reflect.ValueOf(feature)

		// If we are running unit tests we search for the plugin.FeatureExternalAPI
		// implementation, to load and use feature mocks rather than the real one.
		if s.DeployEnvironment() == definition.ServiceDeploy_Test {
			if externalApi, ok := feature.(plugin.FeatureExternalAPI); ok {
				// If the feature has implemented the plugin.FeatureExternalAPI,
				// we give priority for it, trying to check if its returned
				// interface{} has the desired target interface.
				f = reflect.ValueOf(externalApi.ServiceAPI())
			}
		}

		var (
			featureType = f.Type()
			api         = reflect.TypeOf(target).Elem()
		)

		if im := featureType.Implements(api); im {
			reflect.ValueOf(target).Elem().Set(f)
			return nil
		}
	}

	return s.Errors().Internal(errors.New("could not find feature that supports this requested API")).
		Submit(ctx)
}

// Env gives access to the framework environment variables public API.
func (s *Service) Env(name string) string {
	v, ok := s.envs.DefinedEnv(name)
	if !ok {
		// This should not happen because all envs were already loaded
		// when Service was created.
		s.logger.Fatal(context.TODO(), fmt.Sprintf("environment variable '%s' not found", name))
	}

	return v
}

// SetupTest is an api that should start the testing environment for a unit
// test.
func (s *Service) SetupTest(ctx context.Context, t *testing.Testing) *ServiceTesting {
	return setupServiceTesting(ctx, s, t)
}

// CustomDefinitions gives the service access to the service custom settings
// that it may have put inside the 'service.toml' file.
//
// Note that these settings correspond to everything under the [service]
// object inside the TOML file.
func (s *Service) CustomDefinitions() map[string]interface{} {
	return s.definitions.Service
}
