package errors

import (
	"context"

	ferrors "github.com/mikros-dev/mikros/apis/features/errors"
	flogger "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Client struct {
	plugin.Entry
	errors ferrors.ErrorAPI
}

func New() *Client {
	return &Client{}
}

func (c *Client) CanBeInitialized(_ *plugin.CanBeInitializedOptions) bool {
	// Always enabled
	return true
}

func (c *Client) Initialize(_ context.Context, options *plugin.InitializeOptions) error {
	c.errors = options.Errors
	return nil
}

func (c *Client) Fields() []flogger.Attribute {
	return []flogger.Attribute{}
}

func (c *Client) ServiceAPI() interface{} {
	return c.errors
}
