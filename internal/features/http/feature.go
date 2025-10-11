package http

import (
	"context"
	"fmt"

	"github.com/valyala/fasthttp"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/plugin"
)

// Client is the http feature client.
type Client struct {
	plugin.Entry
}

// New creates the http feature.
func New() *Client {
	return &Client{}
}

// CanBeInitialized checks if the feature can be initialized.
func (c *Client) CanBeInitialized(options *plugin.CanBeInitializedOptions) bool {
	_, ok := options.Definitions.ServiceTypes()[definition.ServiceTypeHTTPSpec]
	return ok
}

// Initialize initializes the feature.
func (c *Client) Initialize(_ context.Context, _ *plugin.InitializeOptions) error {
	return nil
}

// AddResponseHeader adds a custom response header by setting a handler-specific
// attribute in the request context.
func (c *Client) AddResponseHeader(ctx context.Context, key, value string) {
	if !c.IsEnabled() {
		return
	}

	if c, ok := ctx.(*fasthttp.RequestCtx); ok {
		// We only accept a string 'value' here to avoid doing conversion
		// inside the handler.
		c.SetUserValue(fmt.Sprintf("handler-attribute-%s", key), value)
	}
}

// SetResponseCode sets a custom HTTP response code in the request context if
// the feature is enabled.
func (c *Client) SetResponseCode(ctx context.Context, code int) {
	if !c.IsEnabled() {
		return
	}

	if c, ok := ctx.(*fasthttp.RequestCtx); ok {
		c.SetUserValue("handler-response-code", code)
	}
}

// Fields returns feature fields to be logged.
func (c *Client) Fields() []logger_api.Attribute {
	return []logger_api.Attribute{}
}
