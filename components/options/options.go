package options

import (
	"errors"

	"github.com/go-playground/validator/v10"

	"github.com/mikros-dev/mikros/components/definition"
)

// NewServiceOptions gathers all the main options that one can use to create a new
// service.
type NewServiceOptions struct {
	// Service must have all required service options according the types
	// defined in the 'service.toml' file. The same type name should be
	// used as key here.
	Service map[string]ServiceOptions `validate:"required"`

	// RunTimeFeatures must hold everything that will only be available
	// when the service executes. The key here should be the same as the
	// feature where the options will be sent.
	RunTimeFeatures map[string]interface{}

	// GrpcClients should have every gRPC dependency that the service
	// may have.
	GrpcClients map[string]*GrpcClient
}

// ServiceOptions is an interface that all services options structure must
// implement.
type ServiceOptions interface {
	Kind() definition.ServiceType
}

// Validate validates if a NewServiceOptions object contains the required information
// initialized to proceed.
func (o *NewServiceOptions) Validate() error {
	if o == nil {
		return errors.New("cannot validate a nil object")
	}

	// Ensures that we're receiving a proper NewServiceOptions object, with
	// everything that we need initialized.
	validate := validator.New()
	if err := validate.Struct(o); err != nil {
		return err
	}

	// Initialize default values for optional members if everything is right.

	if o.GrpcClients == nil {
		o.GrpcClients = make(map[string]*GrpcClient)
	}

	return nil
}
