package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	logger_api "github.com/mikros-dev/mikros/apis/features/logger"
	"github.com/mikros-dev/mikros/components/logger"
	merrors "github.com/mikros-dev/mikros/internal/components/errors"
)

// ProblemOptions configures how error responses are handled and output.
type ProblemOptions struct {
	// HTTPStatusCode specifies the HTTP status code to return. If zero, the
	// status code will be determined automatically based on the error type.
	HTTPStatusCode int

	// Logger is used for logging errors that occur during response writing. If
	// nil, errors will be logged using the standard log package.
	Logger logger_api.LoggerAPI

	// Headers contains additional HTTP headers to include in the response.
	Headers map[string]string

	// Output is a custom function for handling error output. If provided, this
	// function will be called instead of the default error handling.
	Output func(ctx context.Context, w http.ResponseWriter, err error, code int)
}

// Problem outputs an HTTP error response for a handler. It automatically
// determines the appropriate HTTP status code based on the error type, sets the
// content type to JSON, and writes the error message as the response body.
func Problem(ctx context.Context, w http.ResponseWriter, err error, options ...ProblemOptions) {
	var problemOpts ProblemOptions
	if len(options) > 0 {
		problemOpts = options[0]
	}
	if problemOpts.HTTPStatusCode == 0 {
		problemOpts.HTTPStatusCode = errorToStatusCode(err)
	}

	// User custom output for the error.
	if problemOpts.Output != nil {
		problemOpts.Output(ctx, w, err, problemOpts.HTTPStatusCode)
		return
	}

	problem(ctx, w, err, problemOpts)
}

func errorToStatusCode(err error) int {
	var e *merrors.Error
	if !errors.As(err, &e) {
		return http.StatusInternalServerError
	}

	switch e.Kind {
	case merrors.KindNotFound:
		return http.StatusNotFound
	case merrors.KindPermission:
		return http.StatusForbidden
	case merrors.KindPrecondition:
		return http.StatusPreconditionFailed
	case merrors.KindValidation:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func problem(ctx context.Context, w http.ResponseWriter, err error, options ProblemOptions) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range options.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(options.HTTPStatusCode)

	if _, err := w.Write([]byte(err.Error())); err != nil {
		if options.Logger != nil {
			options.Logger.Error(ctx, "failed to write response", logger.Error(err))
			return
		}

		log.Printf("failed to write response: status=%d write error: %v\n", options.HTTPStatusCode, err)
		return
	}
}

// SuccessOptions configures how success responses are handled and output.
type SuccessOptions struct {
	// HTTPStatusCode specifies the HTTP status code to return. If zero, defaults
	// to 200 OK for responses with data, or 204 No Content for nil data.
	HTTPStatusCode int

	// Logger is used for logging errors that occur during response writing. If
	// nil, errors will be logged using the standard log package.
	Logger logger_api.LoggerAPI

	// Headers contains additional HTTP headers to include in the response.
	Headers map[string]string

	// Output is a custom function for handling success output. If provided, this
	// function will be called instead of the default success handling.
	Output func(ctx context.Context, w http.ResponseWriter, data interface{}, code int)
}

// Success outputs an HTTP success response for a handler. It automatically
// handles JSON encoding of the response data, sets appropriate content types
// and headers, and manages different scenarios for nil vs. non-nil data.
//
// When data is nil, it returns a 204 No Content response with no body.
// When data is provided, it JSON-encodes the data and returns it with a
// 200 OK status.
func Success(ctx context.Context, w http.ResponseWriter, data interface{}, options ...SuccessOptions) {
	var successOpts SuccessOptions
	if len(options) > 0 {
		successOpts = options[0]
	}

	// User custom output for success
	if successOpts.Output != nil {
		successOpts.Output(ctx, w, data, successOpts.HTTPStatusCode)
		return
	}

	success(ctx, w, data, successOpts)
}

func success(ctx context.Context, w http.ResponseWriter, data interface{}, options SuccessOptions) {
	if data == nil {
		if options.HTTPStatusCode == 0 {
			options.HTTPStatusCode = http.StatusNoContent
		}

		// Set headers and status code
		for k, v := range options.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(options.HTTPStatusCode)

		return
	}

	if options.HTTPStatusCode == 0 {
		options.HTTPStatusCode = http.StatusOK
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(data); err != nil {
		if options.Logger != nil {
			options.Logger.Error(ctx, "failed to encode response", logger.Error(err))
			return
		}

		log.Printf("failed to encode response: %v\n", err)
		return
	}

	// Set headers and status code
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	for k, v := range options.Headers {
		w.Header().Set(k, v)
	}
	w.WriteHeader(options.HTTPStatusCode)

	// Set body
	if _, err := w.Write(buf.Bytes()); err != nil {
		if options.Logger != nil {
			options.Logger.Error(ctx, "failed to write response", logger.Error(err))
			return
		}

		log.Printf("failed to write response: status=%d write error: %v\n", options.HTTPStatusCode, err)
		return
	}
}
