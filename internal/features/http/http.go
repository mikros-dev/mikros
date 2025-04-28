package http

import (
	"context"
	"fmt"

	"github.com/valyala/fasthttp"

	flogger "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/plugin"
)

type Client struct {
	plugin.Entry
}

func New() *Client {
	return &Client{}
}

func (c *Client) CanBeInitialized(options *plugin.CanBeInitializedOptions) bool {
	_, ok := options.Definitions.ServiceTypes()[definition.ServiceType_HTTP]
	return ok
}

func (c *Client) Initialize(_ context.Context, _ *plugin.InitializeOptions) error {
	return nil
}

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

func (c *Client) SetResponseCode(ctx context.Context, code int) {
	if !c.IsEnabled() {
		return
	}

	if c, ok := ctx.(*fasthttp.RequestCtx); ok {
		c.SetUserValue("handler-response-code", code)
	}
}

func (c *Client) Fields() []flogger.Attribute {
	return []flogger.Attribute{}
}
