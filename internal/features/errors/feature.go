package errors

import (
	"context"

	errors_api "github.com/mikros-dev/mikros/apis/features/errors"
	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Client is the errors feature client.
type Client struct {
	plugin.Entry
	errors errors_api.ErrorAPI
}

// New creates the errors feature.
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
	c.errors = options.Errors
	return nil
}

// Fields returns feature fields to be logged.
func (c *Client) Fields() []logger_api.Attribute {
	return []logger_api.Attribute{}
}

// ServiceAPI returns the errors API that services can use.
func (c *Client) ServiceAPI() interface{} {
	return c.errors
}
