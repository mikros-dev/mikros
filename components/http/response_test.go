package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mikros-dev/mikros/components/logger"
	merrors "github.com/mikros-dev/mikros/internal/components/errors"
)

type code struct {
	Code int
}

func (c *code) ErrorCode() int32 {
	return int32(c.Code)
}

var (
	ctx = context.Background()
)

func TestProblem(t *testing.T) {
	t.Run("with custom Output", func(t *testing.T) {
		var (
			rec    = httptest.NewRecorder()
			err    = errors.New("boom")
			called = false
			opts   = ProblemOptions{
				HTTPStatusCode: http.StatusBadRequest,
				Output: func(ctx context.Context, w http.ResponseWriter, e error, code int) {
					called = true
					assert.Equal(t, err, e)
					assert.Equal(t, http.StatusBadRequest, code)
					w.Header().Set("X-Custom", "1")
					w.WriteHeader(code)
					_, _ = w.Write([]byte("custom"))
				},
			}
		)

		Problem(ctx, rec, err, opts)

		require.True(t, called, "custom Output must be invoked")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "1", rec.Header().Get("X-Custom"))
		assert.Equal(t, "custom", rec.Body.String())
	})

	t.Run("default Output: status, body, content-type", func(t *testing.T) {
		var (
			rec  = httptest.NewRecorder()
			err  = errors.New("something failed")
			opts = ProblemOptions{
				HTTPStatusCode: http.StatusInternalServerError,
			}
		)

		Problem(ctx, rec, err, opts)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "something failed", rec.Body.String())
	})

	t.Run("zero status code passes 0 to WriteHeader", func(t *testing.T) {
		var (
			rec = httptest.NewRecorder()
			err = errors.New("oops")
		)

		Problem(ctx, rec, err, ProblemOptions{})

		assert.Equal(t, 500, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.Equal(t, "oops", rec.Body.String())
	})

	t.Run("mikros errors", func(t *testing.T) {
		factory := merrors.NewFactory(merrors.FactoryOptions{
			ServiceName: "example",
		})

		rec := httptest.NewRecorder()
		e := factory.FailedPrecondition("failed precondition").Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusPreconditionFailed, rec.Code)

		rec = httptest.NewRecorder()
		e = factory.Custom("custom error").Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		rec = httptest.NewRecorder()
		e = factory.Internal(errors.New("internal error")).Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		rec = httptest.NewRecorder()
		e = factory.RPC(errors.New("rpc error"), "example").Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusInternalServerError, rec.Code)

		rec = httptest.NewRecorder()
		e = factory.NotFound().Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusNotFound, rec.Code)

		rec = httptest.NewRecorder()
		e = factory.InvalidArgument(errors.New("invalid argument")).Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusBadRequest, rec.Code)

		rec = httptest.NewRecorder()
		e = factory.PermissionDenied().WithCode(&code{Code: 9951}).WithAttributes(logger.Any("teste", "teste")).Submit(ctx)
		Problem(ctx, rec, e)
		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}

func TestSuccess(t *testing.T) {
	t.Run("with custom Output", func(t *testing.T) {
		var (
			rec    = httptest.NewRecorder()
			data   = map[string]string{"message": "success"}
			called = false
			opts   = SuccessOptions{
				HTTPStatusCode: http.StatusCreated,
				Output: func(ctx context.Context, w http.ResponseWriter, d interface{}, code int) {
					called = true
					assert.Equal(t, data, d)
					assert.Equal(t, http.StatusCreated, code)
					w.Header().Set("X-Custom", "1")
					w.WriteHeader(code)
					_, _ = w.Write([]byte("custom response"))
				},
			}
		)

		Success(ctx, rec, data, opts)

		require.True(t, called, "custom Output must be invoked")
		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "1", rec.Header().Get("X-Custom"))
		assert.Equal(t, "custom response", rec.Body.String())
	})

	t.Run("default Output: status, body, content-type with data", func(t *testing.T) {
		var (
			rec  = httptest.NewRecorder()
			data = map[string]string{"message": "hello world"}
			opts = SuccessOptions{
				HTTPStatusCode: http.StatusOK,
			}
		)

		Success(ctx, rec, data, opts)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"message":"hello world"}`, rec.Body.String())
	})

	t.Run("default Output: nil data returns 204 No Content", func(t *testing.T) {
		var (
			rec = httptest.NewRecorder()
		)

		Success(ctx, rec, nil)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Type"))
	})

	t.Run("zero status code defaults to 200 OK with data", func(t *testing.T) {
		var (
			rec  = httptest.NewRecorder()
			data = map[string]int{"count": 42}
		)

		Success(ctx, rec, data, SuccessOptions{})

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"count":42}`, rec.Body.String())
	})

	t.Run("zero status code defaults to 204 No Content with nil data", func(t *testing.T) {
		var (
			rec = httptest.NewRecorder()
		)

		Success(ctx, rec, nil, SuccessOptions{})

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Empty(t, rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Type"))
	})

	t.Run("custom headers are set", func(t *testing.T) {
		var (
			rec  = httptest.NewRecorder()
			data = map[string]string{"id": "123"}
			opts = SuccessOptions{
				HTTPStatusCode: http.StatusCreated,
				Headers: map[string]string{
					"Location":      "/users/123",
					"X-Request":     "processed",
					"Cache-Control": "no-cache",
				},
			}
		)

		Success(ctx, rec, data, opts)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, "/users/123", rec.Header().Get("Location"))
		assert.Equal(t, "processed", rec.Header().Get("X-Request"))
		assert.Equal(t, "no-cache", rec.Header().Get("Cache-Control"))
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `{"id":"123"}`, rec.Body.String())
	})

	t.Run("custom headers with nil data", func(t *testing.T) {
		var (
			rec  = httptest.NewRecorder()
			opts = SuccessOptions{
				HTTPStatusCode: http.StatusAccepted,
				Headers: map[string]string{
					"X-Processing": "async",
					"Retry-After":  "60",
				},
			}
		)

		Success(ctx, rec, nil, opts)

		assert.Equal(t, http.StatusAccepted, rec.Code)
		assert.Equal(t, "async", rec.Header().Get("X-Processing"))
		assert.Equal(t, "60", rec.Header().Get("Retry-After"))
		assert.Empty(t, rec.Body.String())
		assert.Empty(t, rec.Header().Get("Content-Type"))
	})

	t.Run("complex data structures", func(t *testing.T) {
		type User struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		}

		var (
			rec  = httptest.NewRecorder()
			data = []User{
				{ID: 1, Name: "John"},
				{ID: 2, Name: "Jane"},
			}
		)

		Success(ctx, rec, data)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `[{"id":1,"name":"John"},{"id":2,"name":"Jane"}]`, rec.Body.String())
	})

	t.Run("empty slice returns JSON array", func(t *testing.T) {
		var (
			rec  = httptest.NewRecorder()
			data = []string{}
		)

		Success(ctx, rec, data)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
		assert.JSONEq(t, `[]`, rec.Body.String())
	})
}
