package definition

import (
	"context"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Client is the definition feature client.
type Client struct {
	plugin.Entry
	defs *definition.Definitions
}

// New creates the definition feature.
func New() *Client {
	return &Client{}
}

// CanBeInitialized checks if the feature can be initialized.
func (c *Client) CanBeInitialized(_ *plugin.CanBeInitializedOptions) bool {
	// Always enabled
	return true
}

// Initialize initializes the feature.
func (c *Client) Initialize(_ context.Context, options *plugin.InitializeOptions) error {
	c.defs = options.Definitions
	return nil
}

// Fields returns feature fields to be logged.
func (c *Client) Fields() []logger_api.Attribute {
	return []logger_api.Attribute{}
}

// ServiceName retrieves the service name from the definitions and returns it.
func (c *Client) ServiceName() string {
	return c.defs.ServiceName().String()
}
