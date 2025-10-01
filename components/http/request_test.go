package http

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBind(t *testing.T) {
	t.Run("should bind from multiple locations", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users/123?limit=10", nil)
			v = struct {
				ID    string `json:"id" http:"loc=path"`
				Limit int    `json:"limit" http:"loc=query"`
				Token string `json:"token" http:"loc=header"`
			}{}
		)

		r.SetPathValue("id", "123")
		r.Header.Set("token", "abc123")

		err := Bind(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "123", v.ID)
		assert.Equal(t, 10, v.Limit)
		assert.Equal(t, "abc123", v.Token)
	})

	t.Run("should handle missing values", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users", nil)
			v = struct {
				ID    string `json:"id" http:"loc=path"`
				Limit int    `json:"limit" http:"loc=query"`
			}{}
		)

		err := Bind(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "", v.ID)
		assert.Equal(t, 0, v.Limit)
	})

	t.Run("should return error for non-pointer target", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			v = struct{}{}
		)

		err := Bind(r, v)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target must be a pointer to a struct")
	})

	t.Run("should return error for non-struct target", func(t *testing.T) {
		var (
			r   = httptest.NewRequest(http.MethodGet, "/", nil)
			str = "test"
		)

		err := Bind(r, &str)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "target must be a pointer to a struct")
	})

	t.Run("should skip fields without http tag", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users/123?limit=10", nil)
			v = struct {
				ID       string `json:"id" http:"loc=path"`
				Limit    int    `json:"limit" http:"loc=query"`
				NoTag    string `json:"no_tag"`
				Internal string
			}{}
		)

		r.SetPathValue("id", "123")

		err := Bind(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "123", v.ID)
		assert.Equal(t, 10, v.Limit)
		assert.Equal(t, "", v.NoTag)
		assert.Equal(t, "", v.Internal)
	})

	t.Run("should skip unexported fields", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users/123", nil)
			v = struct {
				ID       string `json:"id" http:"loc=path"`
				internal string `json:"internal" http:"loc=path"`
			}{}
		)

		r.SetPathValue("id", "123")
		r.SetPathValue("internal", "secret")

		err := Bind(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "123", v.ID)
		assert.Equal(t, "", v.internal)
	})
}

func TestBindBody(t *testing.T) {
	t.Run("should bind JSON body", func(t *testing.T) {
		var (
			body = `{"name":"John","age":30}`
			r    = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			v    = struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{}
		)

		err := BindBody(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "John", v.Name)
		assert.Equal(t, 30, v.Age)
	})

	t.Run("should respect MaxBytes limit", func(t *testing.T) {
		var (
			body = `{"name":"John","age":30}`
			r    = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			v    = struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{}
			opts = BindBodyOptions{MaxBytes: 5}
		)

		err := BindBody(r, &v, opts)
		assert.Error(t, err)
	})

	t.Run("should disallow unknown fields when configured", func(t *testing.T) {
		var (
			body = `{"name":"John","age":30,"unknown":"field"}`
			r    = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			v    = struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{}
			opts = BindBodyOptions{DisallowUnknownFields: true}
		)

		err := BindBody(r, &v, opts)
		assert.Error(t, err)
	})

	t.Run("should allow unknown fields by default", func(t *testing.T) {
		var (
			body = `{"name":"John","age":30,"unknown":"field"}`
			r    = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			v    = struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{}
		)

		err := BindBody(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "John", v.Name)
		assert.Equal(t, 30, v.Age)
	})

	t.Run("should reject multiple JSON objects", func(t *testing.T) {
		var (
			body = `{"name":"John"}{"name":"Jane"}`
			r    = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
			v    = struct {
				Name string `json:"name"`
			}{}
		)

		err := BindBody(r, &v)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected only one JSON object")
	})

	t.Run("should handle empty body", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
			v = struct {
				Name string `json:"name"`
			}{}
		)

		err := BindBody(r, &v)
		assert.Error(t, err) // EOF error from JSON decoder
	})
}

func TestBindQuery(t *testing.T) {
	t.Run("should bind single query parameter", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?name=John", nil)
			v = struct {
				Name string `json:"name"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "John", v.Name)
	})

	t.Run("should bind multiple query parameters", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?name=John&age=30&active=true", nil)
			v = struct {
				Name   string `json:"name"`
				Age    int    `json:"age"`
				Active bool   `json:"active"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "John", v.Name)
		assert.Equal(t, 30, v.Age)
		assert.Equal(t, true, v.Active)
	})

	t.Run("should bind slice parameters", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?tags=red&tags=blue&tags=green", nil)
			v = struct {
				Tags []string `json:"tags"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		assert.Equal(t, []string{"red", "blue", "green"}, v.Tags)
	})

	t.Run("should split CSV values", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?tags=red,blue,green", nil)
			v = struct {
				Tags []string `json:"tags"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		assert.Equal(t, []string{"red", "blue", "green"}, v.Tags)
	})

	t.Run("should handle numeric types", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/data?int=42&float=3.14&uint=123", nil)
			v = struct {
				Int   int     `json:"int"`
				Float float64 `json:"float"`
				Uint  uint    `json:"uint"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		assert.Equal(t, 42, v.Int)
		assert.Equal(t, 3.14, v.Float)
		assert.Equal(t, uint(123), v.Uint)
	})

	t.Run("should handle pointer fields", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?name=John", nil)
			v = struct {
				Name *string `json:"name"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		require.NotNil(t, v.Name)
		assert.Equal(t, "John", *v.Name)
	})

	t.Run("should use fallback field names", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?first_name=John", nil)
			v = struct {
				FirstName string
			}{}
			opts = &BindOptions{FallbackSnakeCase: true}
		)

		err := BindQuery(r, &v, opts)
		require.NoError(t, err)
		assert.Equal(t, "John", v.FirstName)
	})

	t.Run("should return error for invalid integer", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?age=abc", nil)
			v = struct {
				Age int `json:"age"`
			}{}
		)

		err := BindQuery(r, &v)
		assert.Error(t, err)
	})

	t.Run("should return error for invalid float", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?score=invalid", nil)
			v = struct {
				Score float64 `json:"score"`
			}{}
		)

		err := BindQuery(r, &v)
		assert.Error(t, err)
	})

	t.Run("should return error for invalid bool", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?active=maybe", nil)
			v = struct {
				Active bool `json:"active"`
			}{}
		)

		err := BindQuery(r, &v)
		assert.Error(t, err)
	})

	t.Run("should return error for unsupported field type", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users?data=test", nil)
			v = struct {
				Data complex64 `json:"data"`
			}{}
		)

		err := BindQuery(r, &v)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported field type")
	})

	t.Run("should bind time.Time fields", func(t *testing.T) {
		var (
			timeStr = "2023-01-01T12:00:00Z"
			r       = httptest.NewRequest(http.MethodGet, "/events?created="+timeStr, nil)
			v       = struct {
				Created time.Time `json:"created"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		expected, _ := time.Parse(time.RFC3339, timeStr)
		assert.Equal(t, expected, v.Created)
	})

	t.Run("should bind time.Duration fields", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/config?timeout=30s", nil)
			v = struct {
				Timeout time.Duration `json:"timeout"`
			}{}
		)

		err := BindQuery(r, &v)
		require.NoError(t, err)
		assert.Equal(t, 30*time.Second, v.Timeout)
	})

	t.Run("should return error for invalid time format", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/events?created=invalid-time", nil)
			v = struct {
				Created time.Time `json:"created"`
			}{}
		)

		err := BindQuery(r, &v)
		assert.Error(t, err)
	})

	t.Run("should return error for invalid duration format", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/config?timeout=invalid", nil)
			v = struct {
				Timeout time.Duration `json:"timeout"`
			}{}
		)

		err := BindQuery(r, &v)
		assert.Error(t, err)
	})
}

func TestBindHeader(t *testing.T) {
	t.Run("should bind single header", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			v = struct {
				Auth string `json:"authorization"`
			}{}
		)

		r.Header.Set("Authorization", "Bearer token123")

		err := BindHeader(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "Bearer token123", v.Auth)
	})

	t.Run("should bind multiple headers", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			v = struct {
				Auth   string `json:"authorization"`
				Type   string `json:"content-type"`
				Length int    `json:"content-length"`
			}{}
		)

		r.Header.Set("Authorization", "Bearer token123")
		r.Header.Set("Content-Type", "application/json")
		r.Header.Set("Content-Length", "100")

		err := BindHeader(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "Bearer token123", v.Auth)
		assert.Equal(t, "application/json", v.Type)
		assert.Equal(t, 100, v.Length)
	})

	t.Run("should handle case insensitive headers", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			v = struct {
				Auth string `json:"authorization"`
			}{}
		)

		r.Header.Set("AUTHORIZATION", "Bearer token123")

		err := BindHeader(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "Bearer token123", v.Auth)
	})

	t.Run("should bind multiple header values", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/", nil)
			v = struct {
				Accept []string `json:"accept"`
			}{}
		)

		r.Header.Add("Accept", "application/json")
		r.Header.Add("Accept", "text/html")

		err := BindHeader(r, &v)
		require.NoError(t, err)
		assert.Equal(t, []string{"application/json", "text/html"}, v.Accept)
	})
}

func TestBindPath(t *testing.T) {
	t.Run("should bind path parameters", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users/123/posts/456", nil)
			v = struct {
				UserID string `json:"user_id"`
				PostID string `json:"post_id"`
			}{}
		)

		r.SetPathValue("user_id", "123")
		r.SetPathValue("post_id", "456")

		err := BindPath(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "123", v.UserID)
		assert.Equal(t, "456", v.PostID)
	})

	t.Run("should handle numeric path parameters", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users/123", nil)
			v = struct {
				ID int `json:"id"`
			}{}
		)

		r.SetPathValue("id", "123")

		err := BindPath(r, &v)
		require.NoError(t, err)
		assert.Equal(t, 123, v.ID)
	})

	t.Run("should handle missing path parameters", func(t *testing.T) {
		var (
			r = httptest.NewRequest(http.MethodGet, "/users", nil)
			v = struct {
				ID string `json:"id"`
			}{}
		)

		err := BindPath(r, &v)
		require.NoError(t, err)
		assert.Equal(t, "", v.ID)
	})
}
