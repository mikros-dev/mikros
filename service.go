package mikros

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/mikros-dev/mikros/apis/behavior"
	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	mcontext "github.com/mikros-dev/mikros/components/context"
	"github.com/mikros-dev/mikros/components/definition"
	mgrpc "github.com/mikros-dev/mikros/components/grpc"
	"github.com/mikros-dev/mikros/components/logger"
	"github.com/mikros-dev/mikros/components/options"
	"github.com/mikros-dev/mikros/components/plugin"
	"github.com/mikros-dev/mikros/components/service"
	"github.com/mikros-dev/mikros/components/testing"
	"github.com/mikros-dev/mikros/internal/components/env"
	merrors "github.com/mikros-dev/mikros/internal/components/errors"
	"github.com/mikros-dev/mikros/internal/components/lifecycle"
	mlogger "github.com/mikros-dev/mikros/internal/components/logger"
	"github.com/mikros-dev/mikros/internal/components/tags"
	"github.com/mikros-dev/mikros/internal/components/tracker"
	"github.com/mikros-dev/mikros/internal/components/validations"
	"github.com/mikros-dev/mikros/internal/features"
	"github.com/mikros-dev/mikros/internal/services"
)

// Service is the object that represents a service application.
type Service struct {
	serviceOptions  map[string]options.ServiceOptions
	runtimeFeatures map[string]interface{}
	errors          *merrors.Factory
	logger          *mlogger.Logger
	ctx             *mcontext.ServiceContext
	servers         []plugin.Service
	clients         map[string]*options.GrpcClient
	definitions     *definition.Definitions
	envs            *env.ServiceEnvs
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
	defs, err := definition.Parse()
	if err != nil {
		return nil, err
	}

	// Loads environment variables
	envs, err := env.NewServiceEnvs(defs)
	if err != nil {
		return nil, err
	}

	// Initialize the service logger system.
	serviceLogger, err := initLogger(defs, envs)
	if err != nil {
		return nil, err
	}

	// Context initialization
	ctx, err := mcontext.New(&mcontext.Options{
		Name: defs.ServiceName(),
	})
	if err != nil {
		return nil, err
	}

	return &Service{
		serviceOptions:  opt.Service,
		runtimeFeatures: opt.RunTimeFeatures,
		errors:          initServiceErrors(defs, serviceLogger),
		logger:          serviceLogger,
		ctx:             ctx,
		clients:         opt.GrpcClients,
		definitions:     defs,
		envs:            envs,
		features:        features.Features(),
		services:        services.Services(),
	}, nil
}

func initLogger(defs *definition.Definitions, envs *env.ServiceEnvs) (*mlogger.Logger, error) {
	// By default, we always discard log messages when running unit tests,
	// but this behavior can be changed using service definitions.
	discardMessages := envs.DeploymentEnv() == definition.ServiceDeployTest
	if discardMessages && defs.Tests.DiscardLogMessages != nil {
		discardMessages = *defs.Tests.DiscardLogMessages
	}

	attributes := map[string]string{
		"service.name":    defs.ServiceName().String(),
		"service.type":    defs.ServiceTypesAsString(),
		"service.version": defs.Version,
		"service.env":     envs.DeploymentEnv().String(),
		"service.product": defs.Product,
	}
	for k, v := range defs.Log.Attributes {
		attributes[k] = v
	}

	// Initialize the service logger system.
	serviceLogger := mlogger.New(mlogger.Options{
		DiscardMessages: discardMessages,
		ErrorStackTrace: defs.Log.ErrorStackTrace,
		FixedAttributes: attributes,
	})

	if defs.Log.Level != "" {
		if _, err := serviceLogger.SetLogLevel(defs.Log.Level); err != nil {
			return nil, err
		}
	}

	return serviceLogger, nil
}

func initServiceErrors(defs *definition.Definitions, log logger_api.API) *merrors.Factory {
	return merrors.NewFactory(merrors.FactoryOptions{
		ServiceName: defs.ServiceName().String(),
		Logger:      log,
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
	ctx := context.Background()

	if err := s.bootstrap(ctx, srv); err != nil {
		s.fatalAbort(ctx, err)
	}

	// If we're running tests, we end the method here to avoid putting the
	// service in execution.
	if s.envs.DeploymentEnv() == definition.ServiceDeployTest {
		return
	}

	s.run(ctx, srv)
}

func (s *Service) bootstrap(ctx context.Context, srv interface{}) *merrors.AbortError {
	s.logger.Info(ctx, "starting service")

	if err := s.postProcessDefinitions(srv); err != nil {
		return merrors.NewAbortError("service definitions error", err)
	}

	if err := s.startFeatures(ctx, srv); err != nil {
		return err
	}

	if err := s.startTracker(); err != nil {
		return merrors.NewAbortError("could not initialize the service tracker", err)
	}

	if err := s.setupLoggerExtractor(); err != nil {
		return merrors.NewAbortError("could not set logger extractor", err)
	}

	if err := s.initializeServiceInternals(ctx, srv); err != nil {
		return err
	}

	s.printServiceResources(ctx)
	return nil
}

// postProcessDefinitions is responsible for loading additional definitions for
// the service. Also, here is where we initialize the service structure member
// tagged as "definitions".
func (s *Service) postProcessDefinitions(srv interface{}) error {
	// Load all feature definitions.
	iter := s.features.Iterator()
	for p, next := iter.Next(); next; p, next = iter.Next() {
		if cfg, ok := p.(plugin.FeatureSettings); ok {
			defs, err := cfg.Definitions(s.definitions.Path())
			if err != nil {
				return err
			}

			s.definitions.AddExternalFeatureDefinitions(p.Name(), defs)
		}
	}

	// Load definitions from all service TOML types and let them available.
	for _, svc := range s.services.Services() {
		if d, ok := svc.(plugin.ServiceSettings); ok {
			defs, err := d.Definitions(s.definitions.Path())
			if err != nil {
				return err
			}

			s.definitions.AddExternalServiceDefinitions(svc.Name(), defs)
		}
	}

	// Load custom service definitions
	if err := s.definitions.LoadCustomServiceDefinitions(srv); err != nil {
		return err
	}

	// Ensure that everything is right
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
		Env:             s.envs,
	}

	// Initialize registered features
	if err := s.features.InitializeAll(ctx, initializeOptions); err != nil {
		return err
	}

	// And execute their Start API
	if err := s.features.StartAll(ctx, srv); err != nil {
		return err
	}

	// Load tagged features into the service struct
	return s.loadTaggedFeatures(ctx, srv)
}

func (s *Service) loadTaggedFeatures(ctx context.Context, srv interface{}) error {
	var (
		typeOf  = reflect.TypeOf(srv)
		valueOf = reflect.ValueOf(srv)
	)

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		typeField := typeOf.Elem().Field(i)
		tag := tags.ParseTag(typeField.Tag)
		if tag == nil || !tag.IsFeature {
			continue
		}

		if valueOf.Elem().Field(i).CanSet() {
			f := reflect.New(typeField.Type).Elem()
			if err := s.Feature(ctx, f.Addr().Interface()); err != nil {
				return err
			}
			valueOf.Elem().Field(i).Set(f)
		}
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
		extractor := api.(behavior.LoggerExtractor)
		s.logger.SetContextFieldExtractor(extractor.Extract)
	}

	return nil
}

func (s *Service) initializeServiceInternals(ctx context.Context, srv interface{}) *merrors.AbortError {
	if err := s.initializeRegisteredServices(ctx, srv); err != nil {
		return merrors.NewAbortError("could not initialize internal services", err)
	}

	// Establishes connection with all gRPC clients.
	if err := s.coupleClients(srv); err != nil {
		return merrors.NewAbortError("could not establish connection with clients", err)
	}

	// Call lifecycle.OnStart before validating the service structure to
	// allow its fields to be initialized at this point. Also ensures that
	// everything declared inside the main struct service is initialized to
	// be used inside the callback.
	if err := lifecycle.OnStart(ctx, srv, &lifecycle.Options{
		Env:            s.envs.DeploymentEnv(),
		ExecuteOnTests: s.definitions.Tests.ExecuteLifecycle,
	}); err != nil {
		return merrors.NewAbortError("failed while running lifecycle.OnStart", err)
	}

	if s.envs.DeploymentEnv() != definition.ServiceDeployTest {
		if err := validations.EnsureValuesAreInitialized(srv); err != nil {
			return merrors.NewAbortError("service server object is not properly initialized", err)
		}
	}

	return nil
}

func (s *Service) initializeRegisteredServices(ctx context.Context, srv interface{}) error {
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
			Port:           s.getServicePort(servicePort, serviceType.String()),
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
			Env:            s.envs,
		}); err != nil {
			return err
		}

		// Saves only the initialized services
		s.servers = append(s.servers, svc)
	}

	return nil
}

func (s *Service) getServicePort(port service.ServerPort, serviceType string) service.ServerPort {
	// Use default port values in case no port was set in the service.toml
	if port == 0 {
		if serviceType == definition.ServiceTypeGRPC.String() {
			return service.ServerPort(s.envs.GrpcPort())
		}

		if serviceType == definition.ServiceTypeHTTPSpec.String() ||
			serviceType == definition.ServiceTypeHTTP.String() {
			return service.ServerPort(s.envs.HTTPPort())
		}
	}

	return port
}

// coupleClients establishes connections with all client services that a service
// has as dependency.
func (s *Service) coupleClients(srv interface{}) error {
	// If the service does not have dependencies, or we are running tests,
	// don't need to continue.
	if len(s.clients) == 0 || s.envs.DeploymentEnv() == definition.ServiceDeployTest {
		return nil
	}

	var (
		typeOf  = reflect.TypeOf(srv)
		valueOf = reflect.ValueOf(srv)
	)

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		typeField := typeOf.Elem().Field(i)
		tag := tags.ParseTag(typeField.Tag)
		if tag == nil || !tag.IsClientTag() {
			continue
		}

		client, ok := s.clients[tag.GrpcClientName]
		if !ok {
			return fmt.Errorf("could not find gRPC client '%s' inside service options", tag.GrpcClientName)
		}
		if err := client.Validate(); err != nil {
			return err
		}

		cOpts := s.createGrpcCoupledClientOptions(client)
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

	return nil
}

func (s *Service) createGrpcCoupledClientOptions(client *options.GrpcClient) *mgrpc.ClientConnectionOptions {
	serviceTracker, _ := s.tracker.Tracker()

	// For each valid client, establishes their gRPC connection and
	// initializes the service structure properly by pointing its
	// members to these connections.

	opts := &mgrpc.ClientConnectionOptions{
		ServiceName: s.definitions.ServiceName(),
		ClientName:  client.ServiceName,
		Context:     s.ctx,
		Connection: mgrpc.ConnectionOptions{
			Namespace: s.envs.CoupledNamespace(),
			Port:      s.envs.CoupledPort(),
		},
		Tracker: serviceTracker,
	}

	if s.definitions.Clients != nil {
		if opt, ok := s.definitions.Clients[client.ServiceName.String()]; ok {
			opts.AlternativeConnection = &mgrpc.ConnectionOptions{
				Host: opt.Host,
				Port: opt.Port,
			}
		}
	}

	return opts
}

func (s *Service) printServiceResources(ctx context.Context) {
	var (
		fields []logger_api.Attribute
		iter   = s.features.Iterator()
	)

	for f, next := iter.Next(); next; f, next = iter.Next() {
		fields = append(fields, f.Fields()...)
	}

	s.logger.Info(ctx, "service resources", fields...)
}

func (s *Service) run(ctx context.Context, srv interface{}) {
	defer s.stopService(ctx)
	defer lifecycle.OnFinish(ctx, srv, &lifecycle.Options{
		Env:            s.envs.DeploymentEnv(),
		ExecuteOnTests: s.definitions.Tests.ExecuteLifecycle,
	})

	// In case we're a script service, only execute its function and terminate
	// the execution.
	if s.definitions.IsServiceType(definition.ServiceTypeScript) {
		svc := s.servers[0]
		s.logger.Info(ctx, "service is running", svc.Info()...)

		if err := svc.Run(ctx, srv); err != nil {
			s.fatalAbort(ctx, merrors.NewAbortError("could not execute service", err))
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
		s.fatalAbort(ctx, merrors.NewAbortError("could not execute service", err))

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
				append([]logger_api.Attribute{logger.Error(err)}, svc.Info()...)...)
		}
	}

	s.logger.Info(ctx, "service stopped")
}

// stopDependentServices stops other services that are running along with the
// main service.
func (s *Service) stopDependentServices(ctx context.Context) error {
	s.logger.Info(ctx, "stopping dependent services")
	return s.features.CleanupAll(ctx)
}

// Logger gives access to the logger API from inside a service context.
//
// Deprecated: This method is deprecated and should not be used anymore. To access
// the log API, one must declare an internal service feature and initialize it
// using struct tags.
func (s *Service) Logger() logger_api.API {
	return s.logger
}

// Errors gives access to the errors API from inside a service context.
//
// Deprecated: This method is deprecated and should not be used anymore. To access
// the error API, one must declare an internal service feature
// and initialize it using struct tags.
func (s *Service) Errors() errors_api.ErrorAPI {
	return s.errors
}

// Abort is a helper method to abort services in the right way when external
// initialization is needed.
func (s *Service) Abort(message string, err error) {
	s.fatalAbort(context.TODO(), merrors.NewAbortError(message, err))
}

// abort is an internal helper method to finish the service execution with an
// error message.
func (s *Service) fatalAbort(ctx context.Context, err *merrors.AbortError) {
	s.logger.Fatal(ctx, err.Message, logger.Error(err.InnerError))
}

// ServiceName gives back the service name.
//
// Deprecated: This method is deprecated and should not be used anymore. To know
// the current service name, one must declare an internal service feature for
// the definitions and initialize it using struct tags.
func (s *Service) ServiceName() string {
	return s.definitions.ServiceName().String()
}

// DeployEnvironment exposes the current service deploymentEnv environment.
//
// Deprecated: This method is deprecated and should not be used anymore. To know
// this information, one must declare an internal service feature for the
// environment variables and initialize it using struct tags.
func (s *Service) DeployEnvironment() definition.ServiceDeploy {
	return s.envs.DeploymentEnv()
}

// tags function gives a map of current service tags to be used with external
// resources.
func (s *Service) tags() map[string]string {
	serviceType := s.definitions.ServiceTypesAsString()
	if strings.Contains(serviceType, ",") {
		// SQS tag does not accept commas, just Unicode letters, digits,
		// whitespace, or one of these symbols: _ . : / = + - @
		serviceType = "hybrid"
	}

	return map[string]string{
		"service.name":    s.definitions.ServiceName().String(),
		"service.type":    serviceType,
		"service.version": s.definitions.Version,
		"service.product": s.definitions.Product,
	}
}

// Feature is the service mechanism to have access to an external feature
// public API.
func (s *Service) Feature(ctx context.Context, target interface{}) error {
	if reflect.TypeOf(target).Kind() != reflect.Ptr {
		return s.errors.Internal(errors.New("requested target API must be a pointer")).
			Submit(ctx)
	}

	it := s.features.Iterator()
	for {
		feature, next := it.Next()
		if !next {
			break
		}

		f := reflect.ValueOf(feature)
		if externalAPI, ok := feature.(plugin.FeatureExternalAPI); ok {
			// If the feature has implemented the plugin.FeatureExternalAPI,
			// we give priority to it, trying to check if its returned
			// interface{} has the desired target interface. This way, we let the
			// feature decide if it is going to implement its public interface
			// itself or if it will return something that implements.
			f = reflect.ValueOf(externalAPI.ServiceAPI())
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

	return s.errors.Internal(errors.New("could not find feature that supports this requested API")).
		Submit(ctx)
}

// Env gives access to the framework environment variables public API.
//
// Deprecated: This method is deprecated and should not be used anymore. To load
// environment variable values, one must declare an internal service feature and
// initialize it using struct tags.
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
//
// Deprecated: This method is deprecated and should not be used anymore. To load
// custom service definitions, use the tag `mikros:"definitions"` with a structure
// member inside the service.
func (s *Service) CustomDefinitions() map[string]interface{} {
	return s.definitions.Service
}
