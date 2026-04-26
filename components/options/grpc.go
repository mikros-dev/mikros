package options

import (
	"fmt"
	"reflect"

	"google.golang.org/grpc"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

// GrpcServiceOptions gathers options to initialize a gRPC runtime
type GrpcServiceOptions struct {
	ProtoServiceDescription *grpc.ServiceDesc
}

// Kind returns the runtime type as definition.RuntimeTypeGRPC.
func (g *GrpcServiceOptions) Kind() definition.RuntimeType {
	return definition.RuntimeTypeGRPC
}

// GrpcClient is a structure to set information about a gRPC client that will
// be coupled with another service.
type GrpcClient struct {
	// ServiceName should be the service name.
	ServiceName service.Name

	// NewClientFunction should point to the service API function that can create
	// its gRPC client interface.
	NewClientFunction interface{}
}

// Validate checks if the GrpcClient is properly initialized and its
// NewClientFunction is a valid function.
func (g *GrpcClient) Validate() error {
	if g.NewClientFunction == nil {
		return fmt.Errorf("client '%s' does not have its API initialized", g.ServiceName)
	}

	v := reflect.ValueOf(g.NewClientFunction)
	if v.Type().Kind() != reflect.Func {
		return fmt.Errorf("client '%s' does not have a valid API function", g.ServiceName)
	}

	return nil
}
