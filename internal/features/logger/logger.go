package logger

import (
	"context"

	flogger "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Client struct {
	plugin.Entry
	flogger.LoggerAPI
}

func New() *Client {
	return &Client{}
}

func (c *Client) CanBeInitialized(_ *plugin.CanBeInitializedOptions) bool {
	// Always enabled
	return true
}

func (c *Client) Initialize(_ context.Context, options *plugin.InitializeOptions) error {
	c.LoggerAPI = options.Logger
	return nil
}

func (c *Client) Fields() []flogger.Attribute {
	return []flogger.Attribute{}
}

func (c *Client) ServiceAPI() interface{} {
	return c.LoggerAPI
}
