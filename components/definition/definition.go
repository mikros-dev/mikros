package definition

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"

	"github.com/mikros-dev/mikros/components/service"
	"github.com/mikros-dev/mikros/internal/components/tags"
)

// Definitions is a structure representation of a 'service.toml' file. It holds
// all service information that will be used to initialize it as well as all
// features it will have when executing.
//
//revive:disable:line-length-limit
type Definitions struct {
	Name     string                            `toml:"name" validate:"required"`
	Types    []string                          `toml:"types" validate:"required,single_script,no_duplicated_runtime,dive,runtime_type"`
	Version  string                            `toml:"version" validate:"required,version"`
	Language string                            `toml:"language" validate:"required,oneof=go rust"`
	Product  string                            `toml:"product" validate:"required"`
	Envs     []string                          `toml:"envs,omitempty" validate:"dive,ascii,uppercase"`
	Features Features                          `toml:"features,omitempty"`
	Log      Log                               `toml:"log,omitempty"`
	Tests    Tests                             `toml:"tests,omitempty"`
	Service  map[string]interface{}            `toml:"service,omitempty"`
	Clients  map[string]GrpcClient             `toml:"clients,omitempty"`
	Runtimes map[string]map[string]interface{} `toml:"runtimes,omitempty"`

	path                  string
	supportedRuntimeTypes []string
	externalRuntimes      map[string]ExternalRuntimeEntry
}

// Log represents configuration settings for logging in a service.
type Log struct {
	ErrorStackTrace string            `toml:"error_stack_trace,omitempty" validate:"omitempty,oneof=default disabled structured" default:"default"`
	Level           string            `toml:"level,omitempty" validate:"omitempty,oneof=info debug error warn internal"`
	Attributes      map[string]string `toml:"attributes,omitempty"`
}

// GrpcClient defines the configuration settings for a gRPC coupled client.
type GrpcClient struct {
	Port int32  `toml:"port"`
	Host string `toml:"host"`
}

// Features is a structure that defines a list of features that a service may
// use or not when executing. By convention, all features are turned off
// by default and should be explicitly enabled when desired using the 'enabled' key.
type Features struct {
	// externalFeatures holds settings from all external features that have
	// support for them.
	externalFeatures map[string]ExternalFeatureEntry
}

// ExternalFeatureEntry is a behavior that all external features must have to be
// supported by the package Definitions object.
type ExternalFeatureEntry interface {
	// Enabled must return true or false if the feature is enabled or not.
	Enabled() bool

	// Validate should validate if the custom settings are valid or not.
	Validate() error
}

// ExternalRuntimeEntry is a behavior that all external runtime implementations
// must have to be supported by the Definitions object.
type ExternalRuntimeEntry interface {
	// Name must return the runtime name that the definitions will support.
	Name() string

	// Validate should validate if the custom settings are valid or not.
	Validate() error
}

// Tests gathers unit tests related options.
type Tests struct {
	ExecuteLifecycle   bool  `toml:"execute_lifecycle,omitempty"`
	DiscardLogMessages *bool `toml:"discard_log_messages,omitempty"`
}

//revive:enable:line-length-limit

// New creates a new Definitions structure initializing the service
// features with default values.
func New() (*Definitions, error) {
	defs := &Definitions{}
	if err := defaults.Set(defs); err != nil {
		return nil, err
	}

	// Starts with framework's runtimes
	defs.supportedRuntimeTypes = SupportedRuntimeTypes()

	return defs, nil
}

// Validate validates if all data loaded from the service definitions is
// correct.
//
// It also validates external services and external features custom definitions.
func (d *Definitions) Validate() error {
	validate := validator.New()

	if err := validate.RegisterValidationCtx("version", versionValidator); err != nil {
		return err
	}

	if err := validate.RegisterValidationCtx("runtime_type", runtimeTypeValidator); err != nil {
		return err
	}

	if err := validate.RegisterValidationCtx("single_script", scriptTypeUniqueValidator); err != nil {
		return err
	}

	if err := validate.RegisterValidationCtx("no_duplicated_runtime", duplicatedRuntimeValidator); err != nil {
		return err
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, runtimeTypeCtx{}, d.supportedRuntimeTypes)

	if err := validate.StructCtx(ctx, d); err != nil {
		return err
	}

	for _, svc := range d.externalRuntimes {
		if err := svc.Validate(); err != nil {
			return err
		}
	}

	for _, f := range d.Features.externalFeatures {
		if err := f.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// IsRuntimeType checks if the current service definitions is of a specific
// runtime type.
func (d *Definitions) IsRuntimeType(runtimeType RuntimeType) bool {
	for t := range d.RuntimeTypes() {
		if t == runtimeType {
			return true
		}
	}

	return false
}

// ServiceName returns the service.Name loaded from the definitions.
func (d *Definitions) ServiceName() service.Name {
	return service.FromString(d.Name)
}

// RuntimeTypesAsString converts all runtime types in the definitions to a
// single comma-separated string.
func (d *Definitions) RuntimeTypesAsString() string {
	var s []string

	for t := range d.RuntimeTypes() {
		s = append(s, t.String())
	}

	return strings.Join(s, ",")
}

// RuntimeTypes gives back all runtime types found inside the service definitions.
func (d *Definitions) RuntimeTypes() map[RuntimeType]service.ServerPort {
	runtimes := make(map[RuntimeType]service.ServerPort)
	for _, rt := range d.Types {
		t, p := splitTypeAndPort(rt)

		var (
			runtimeType = CreateRuntimeType(t)
			port        = service.ServerPort(p)
		)

		runtimes[runtimeType] = port
	}

	return runtimes
}

func splitTypeAndPort(runtimeType string) (string, int32) {
	parts := strings.Split(runtimeType, ":")
	if len(parts) == 1 {
		return runtimeType, int32(0)
	}

	// Ignores the error since the Validate was already called.
	p, _ := strconv.ParseInt(parts[1], 10, 32)
	return parts[0], int32(p)
}

// AddExternalFeatureDefinitions adds definitions from external features into
// the Definitions object.
func (d *Definitions) AddExternalFeatureDefinitions(name string, defs ExternalFeatureEntry) {
	if d.Features.externalFeatures == nil {
		d.Features.externalFeatures = make(map[string]ExternalFeatureEntry)
	}

	d.Features.externalFeatures[name] = defs
}

// ExternalFeatureDefinitions retrieves definitions from an external feature
// previously added into the Definitions.
func (d *Definitions) ExternalFeatureDefinitions(name string) (ExternalFeatureEntry, error) {
	v, ok := d.Features.externalFeatures[name]
	if !ok {
		return nil, fmt.Errorf("could not find definitions for feature '%v'", name)
	}

	return v, nil
}

// AddExternalRuntimeDefinitions adds definitions from external runtime into
// the Definitions object.
func (d *Definitions) AddExternalRuntimeDefinitions(name string, defs ExternalRuntimeEntry) {
	if d.externalRuntimes == nil {
		d.externalRuntimes = make(map[string]ExternalRuntimeEntry)
	}

	d.externalRuntimes[name] = defs
}

// AddSupportedRuntimeType adds a new runtime type as supported by the service
// definitions.
func (d *Definitions) AddSupportedRuntimeType(name string) {
	isIn := func(n string, h []string) bool {
		for _, e := range h {
			if e == n {
				return true
			}
		}

		return false
	}

	if !isIn(name, d.supportedRuntimeTypes) {
		d.supportedRuntimeTypes = append(d.supportedRuntimeTypes, name)
	}
}

// ExternalRuntimeDefinitions retrieves definitions from an external runtime
// previously added into the Definitions.
func (d *Definitions) ExternalRuntimeDefinitions(name string) (ExternalRuntimeEntry, error) {
	v, ok := d.externalRuntimes[name]
	if !ok {
		return nil, fmt.Errorf("could not find definitions for service '%v'", name)
	}

	return v, nil
}

// LoadRuntime retrieves only definitions from a specific runtime type.
func (d *Definitions) LoadRuntime(runtimeType RuntimeType) (map[string]interface{}, bool) {
	dd, ok := d.Runtimes[runtimeType.String()]
	return dd, ok
}

// LoadCustomServiceDefinitions loads the [service] object directly inside the
// service member tagged with "definitions".
func (d *Definitions) LoadCustomServiceDefinitions(srv interface{}) error {
	var (
		v = reflect.ValueOf(srv).Elem()
		t = v.Type()
	)

	for i := 0; i < t.NumField(); i++ {
		var (
			buf      bytes.Buffer
			field    = t.Field(i)
			fieldTag = tags.ParseTag(field.Tag)
		)

		if fieldTag == nil {
			continue
		}

		if fieldTag.IsDefinitions {
			if err := d.handleServiceDefinitions(&buf, i, v, field); err != nil {
				return err
			}

			// Only one service definition is allowed.
			break
		}
	}

	return nil
}

func (d *Definitions) handleServiceDefinitions(
	buf *bytes.Buffer,
	i int,
	v reflect.Value,
	field reflect.StructField,
) error {
	// Serialize service settings back into TOML for us
	if err := toml.NewEncoder(buf).Encode(d.Service); err != nil {
		return err
	}

	fieldVal := v.Field(i)
	if fieldVal.IsNil() {
		fieldVal.Set(reflect.New(field.Type.Elem()))
	}

	// Decode TOML into the custom service structure
	if _, err := toml.Decode(buf.String(), fieldVal.Interface()); err != nil {
		return err
	}

	// Validates the settings just loaded.
	if validador, ok := fieldVal.Interface().(Validator); ok {
		if err := validador.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Path returns the original path loaded to the current definitions.
func (d *Definitions) Path() string {
	return d.path
}
