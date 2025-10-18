package env

import (
	"context"

	env_api "github.com/mikros-dev/mikros/apis/features/env"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Client is the env feature client.
type Client struct {
	plugin.Entry
	envs env_api.API
}

// New creates the env feature.
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
	c.envs = options.Env
	return nil
}

// Fields returns feature fields to be logged.
func (c *Client) Fields() []logger_api.Attribute {
	return []logger_api.Attribute{}
}

// ServiceAPI returns the env API that services can use.
func (c *Client) ServiceAPI() interface{} {
	return c.envs
}
