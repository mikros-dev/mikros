package options

import (
	"fmt"
	"reflect"

	"google.golang.org/grpc"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

// GrpcServiceOptions gathers options to initialize a gRPC service.
type GrpcServiceOptions struct {
	ProtoServiceDescription *grpc.ServiceDesc
}

func (g *GrpcServiceOptions) Kind() definition.ServiceType {
	return definition.ServiceType_gRPC
}

// GrpcClient is a structure to set information about a gRPC client that will
// be coupled by a service.
type GrpcClient struct {
	// ServiceName should be the service name.
	ServiceName service.Name

	// NewClientFunction should point to the service API function that can create
	// its gRPC client interface.
	NewClientFunction interface{}
}

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
