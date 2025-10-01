package errors

import (
	"context"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Client struct {
	plugin.Entry
	errors errors_api.ErrorAPI
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

func (c *Client) Fields() []logger_api.Attribute {
	return []logger_api.Attribute{}
}

func (c *Client) ServiceAPI() interface{} {
	return c.errors
}
