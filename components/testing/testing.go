// Package testing is a framework package to be used inside services unit tests
// providing an API to build specific services unit tests.
package testing

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"
	"go.uber.org/mock/gomock"
)

const (
	JSONContentType           = "application/json"
	FormUrlencodedContentType = "application/x-www-form-urlencoded"
	FormDataContentType       = "multipart/form-data"
)

type RequestOptions struct {
	Path        string
	Headers     map[string]string
	ContentType string
	Body        interface{}
	QueryParams map[string]string
}

func (r *RequestOptions) getContentType() string {
	ct := JSONContentType
	if r.ContentType != "" {
		ct = r.ContentType
	}

	return ct
}

type Response struct {
	statusCode int
	body       []byte
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) Body() []byte {
	return r.body
}

type Testing struct {
	options     *Options
	t           *testing.T
	assert      *assert.Assertions
	ctrl        *gomock.Controller
	httpHandler fasthttp.RequestHandler
}

// Options gathers all available options that can be swapped inside a
// unit test. It is important that the Setup/Teardown API is used
// correctly to avoid receiving unexpected behavior.
type Options struct {
	// Handler is the service main HTTP handler to be used inside a test.
	Handler fasthttp.RequestHandler

	// MockedFeatures is an array of features that will be mocked during the
	// test.
	MockedFeatures []interface{}

	// FeatureOptions is the mechanism that a test can pass specific features
	// options to be used by it.
	FeatureOptions map[string]interface{}
}

// New creates a new Testing object, to help building service unit tests. It can
// be used inside a Test function or test subset providing access to assertions
// and APIs to build specific required types when calling services handlers.
func New(t *testing.T, options ...*Options) *Testing {
	var (
		opt     *Options
		handler fasthttp.RequestHandler
	)

	if len(options) > 0 {
		opt = options[0]

		if opt != nil && opt.Handler != nil {
			handler = opt.Handler
		}
	}

	return &Testing{
		assert:      assert.New(t),
		ctrl:        gomock.NewController(t),
		httpHandler: handler,
		t:           t,
		options:     opt,
	}
}

// T gives access to the internal golang testing.T object.
func (t *Testing) T() *testing.T {
	return t.t
}

// Assert gives access to an assert object, providing an API to compare values
// between variables or objects.
func (t *Testing) Assert() *assert.Assertions {
	return t.assert
}

// MockController gives access to a gomock.Controller object that enables
// creating a service mock to be used inside tests.
func (t *Testing) MockController() *gomock.Controller {
	return t.ctrl
}

func (t *Testing) MockAny() gomock.Matcher {
	return gomock.Any()
}

// SkipCICD skips the test if the CICD_TEST environment variable is set to true.
func (t *Testing) SkipCICD() {
	if os.Getenv("CICD_TEST") != "" {
		t.t.SkipNow()
	}
}

// Options gives access to the internal Testing layer options.
func (t *Testing) Options() *Options {
	return t.options
}

// HttpHandler gives access to the service HTTP request handler.
func (t *Testing) HttpHandler() fasthttp.RequestHandler {
	return t.httpHandler
}

// Get makes a GET request to a service's endpoint.
func (t *Testing) Get(opts *RequestOptions) (*Response, error) {
	return t.makeRequest(opts, http.MethodGet)
}

// Post makes a POST request to a service's endpoint.
func (t *Testing) Post(opts *RequestOptions) (*Response, error) {
	return t.makeRequest(opts, http.MethodPost)
}

// Put makes a PUT request to a service's endpoint.
func (t *Testing) Put(opts *RequestOptions) (*Response, error) {
	return t.makeRequest(opts, http.MethodPut)
}

// Delete makes a DELETE request to a service's endpoint.
func (t *Testing) Delete(opts *RequestOptions) (*Response, error) {
	return t.makeRequest(opts, http.MethodDelete)
}

// Patch makes a PATCH request to a service's endpoint.
func (t *Testing) Patch(opts *RequestOptions) (*Response, error) {
	return t.makeRequest(opts, http.MethodPatch)
}

func (t *Testing) makeRequest(opts *RequestOptions, method string) (*Response, error) {
	req, err := createRequest(opts, method)
	if err != nil {
		return nil, err
	}

	res := fasthttp.AcquireResponse()
	err = makeHttpRequest(t, req, res)
	if err != nil {
		return nil, err
	}

	return &Response{
		statusCode: res.StatusCode(),
		body:       res.Body(),
	}, nil
}

func createRequest(opts *RequestOptions, method string) (*fasthttp.Request, error) {
	var (
		req     = fasthttp.AcquireRequest()
		baseURL = fmt.Sprintf("http://test%s", opts.Path)
		uri     fasthttp.URI
	)

	uri.Update(baseURL)
	for k, v := range opts.QueryParams {
		uri.QueryArgs().Add(k, v)
	}

	req.SetRequestURI(uri.String())
	req.Header.SetMethod(method)
	req.Header.SetContentType(opts.getContentType())

	if opts.Body != nil {
		b := opts.Body
		if opts.getContentType() == JSONContentType || opts.getContentType() == FormUrlencodedContentType {
			body, err := json.Marshal(b)
			if err != nil {
				return nil, err
			}
			req.SetBodyString(string(body))
		}

		if opts.getContentType() == FormDataContentType {
			req.SetBody(b.([]byte))
		}
	}

	for key, value := range opts.Headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func makeHttpRequest(t *Testing, req *fasthttp.Request, res *fasthttp.Response) error {
	if err := httpTestingIsEnabled(t); err != nil {
		return err
	}

	ln := fasthttputil.NewInmemoryListener()
	defer func(ln *fasthttputil.InmemoryListener) {
		_ = ln.Close()
	}(ln)

	go func() {
		if err := fasthttp.Serve(ln, t.HttpHandler()); err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}()

	client := fasthttp.Client{
		Dial: func(addr string) (net.Conn, error) {
			return ln.Dial()
		},
	}

	return client.Do(req, res)
}

func httpTestingIsEnabled(t *Testing) error {
	if t.HttpHandler() == nil {
		return errors.New("http testing is not enabled")
	}

	return nil
}
