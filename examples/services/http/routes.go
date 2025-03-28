package main

import (
	"context"
	"encoding/json"

	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"

	"github.com/somatech1/mikros/protobuf-examples/gen/go/services/user_bff"
)

type routes struct {
	*service
}

func (r *routes) SetupServer(_ string, _ interface{}, router *router.Router, _ interface{}, _ func(ctx context.Context, handlers map[string]interface{}) error) error {
	router.POST("/user-bff/v1/users/{name}", func(ctx *fasthttp.RequestCtx) {
		// Parse request body
		req := &user_bff.CreateUserRequest{}
		if len(ctx.PostBody()) == 0 {
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		if err := json.Unmarshal(ctx.PostBody(), req); err != nil {
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		name := ctx.UserValue("name").(string)
		req.Name = name

		// Call the handler
		out, err := r.service.CreateUser(ctx, req)
		if err != nil {
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		// Send the response
		b, err := json.Marshal(out)
		if err != nil {
			ctx.Response.SetStatusCode(fasthttp.StatusInternalServerError)
			return
		}

		ctx.Response.SetStatusCode(fasthttp.StatusCreated)
		ctx.Response.SetBody(b)
	})

	return nil
}
