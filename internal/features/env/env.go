package env

import (
	"context"

	fenv "github.com/mikros-dev/mikros/apis/features/env"
	flogger "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Client struct {
	plugin.Entry
	envs fenv.EnvAPI
}

func New() *Client {
	return &Client{}
}

func (c *Client) CanBeInitialized(_ *plugin.CanBeInitializedOptions) bool {
	// Always enabled
	return true
}

func (c *Client) Initialize(_ context.Context, options *plugin.InitializeOptions) error {
	c.envs = options.Env
	return nil
}

func (c *Client) Fields() []flogger.Attribute {
	return []flogger.Attribute{}
}

func (c *Client) ServiceAPI() interface{} {
	return c.envs
}
