package http

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/stoewer/go-strcase"
)

// Bind extracts and binds HTTP request parameters to a struct based on struct
// field tags. It supports binding from multiple sources (path, query, headers)
// based on the `http` struct tag.
//
// Struct fields must have an `http` tag with a location specifier:
//   - `http:"loc=path"` - extract from URL path parameters
//   - `http:"loc=query"` - extract from query string parameters
//   - `http:"loc=header"` - extract from HTTP headers
//
// Field names are resolved from the `json` tag, or fall back to the struct field name.
//
// Example usage:
//
//	type RequestParams struct {
//	    UserID string `json:"user_id" http:"loc=path"`
//	    Filter string `json:"filter" http:"loc=query"`
//	    Token  string `json:"token" http:"loc=header"`
//	}
//
//	var params RequestParams
//	err := Bind(r, &params)
func Bind(r *http.Request, target interface{}) error {
	o := getBindOptions()

	b, err := newBinder(r, target, &o)
	if err != nil {
		return err
	}

	for i := 0; i < b.rt.NumField(); i++ {
		if err := b.bindField(i); err != nil {
			return err
		}
	}

	return nil
}

type binder struct {
	r          *http.Request
	target     interface{}
	rv         reflect.Value
	rt         reflect.Type
	opt        *BindOptions
	bodyParsed interface{}
}

func newBinder(r *http.Request, target interface{}, opt *BindOptions) (*binder, error) {
	rv, rt, err := validateBindTarget(target)
	if err != nil {
		return nil, err
	}

	return &binder{
		r:      r,
		target: target,
		rv:     rv,
		rt:     rt,
		opt:    opt,
	}, nil
}

func validateBindTarget(target interface{}) (reflect.Value, reflect.Type, error) {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return reflect.Value{}, nil, errors.New("target must be a pointer to a struct")
	}

	rv := v.Elem()
	return rv, rv.Type(), nil
}

func (b *binder) bindField(index int) error {
	sf := b.rt.Field(index)
	fv := b.rv.Field(index)

	if !fv.CanSet() {
		return nil
	}

	name, ok := resolveFieldName(sf, b.opt.FallbackSnakeCase)
	if !ok {
		return nil
	}

	tag, err := parseBindTag(sf.Tag)
	if err != nil || tag == nil {
		return err
	}

	if tag.Location == "body" {
		return b.bindFromBody(index, sf, fv)
	}

	return b.bindFromExtractor(name, tag.Location, sf, fv)
}

func (b *binder) bindFromBody(index int, sf reflect.StructField, fv reflect.Value) error {
	if err := b.ensureBodyParsed(); err != nil {
		return err
	}

	bf := reflect.ValueOf(b.bodyParsed).Elem().Field(index)
	if isZeroValue(bf) {
		return nil
	}

	return setFieldValues(fv, sf, []string{
		fmt.Sprintf("%v", bf.Interface()),
	}, b.opt)
}

func (b *binder) ensureBodyParsed() error {
	if b.bodyParsed != nil {
		return nil
	}

	bt := reflect.New(b.rt).Interface()
	if err := BindBody(b.r, bt); err != nil {
		return err
	}
	b.bodyParsed = bt

	return nil
}

func (b *binder) bindFromExtractor(
	name, location string,
	sf reflect.StructField,
	fv reflect.Value,
) error {
	val := extractor(location, name, b.r)
	if val == "" {
		return nil
	}

	return setFieldValues(fv, sf, []string{val}, b.opt)
}

func extractor(location, name string, r *http.Request) string {
	switch strings.ToLower(location) {
	case "path":
		return r.PathValue(name)
	case "query":
		return r.URL.Query().Get(name)
	case "header":
		return r.Header.Get(name)
	default:
		return ""
	}
}

func isZeroValue(v reflect.Value) bool {
	return !v.IsValid() || reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// BindBodyOptions configures the behavior of BindBody.
type BindBodyOptions struct {
	// MaxBytes limits the size of the request body (0 = unlimited)
	MaxBytes int64

	// DisallowUnknownFields reject JSON with fields not present in the target
	// struct.
	DisallowUnknownFields bool
}

// BindBody decodes a JSON request body into a target struct. It supports
// optional limits on body size and strict field validation.
func BindBody(r *http.Request, target interface{}, options ...BindBodyOptions) error {
	var bindOpts BindBodyOptions
	if len(options) > 0 {
		bindOpts = options[0]
	}

	body := io.Reader(r.Body)
	if bindOpts.MaxBytes > 0 {
		body = io.LimitReader(body, bindOpts.MaxBytes)
	}

	dec := json.NewDecoder(body)
	if bindOpts.DisallowUnknownFields {
		dec.DisallowUnknownFields()
	}

	if err := dec.Decode(target); err != nil {
		return err
	}

	// Ensure we're dealing with a single JSON
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		return errors.New("expected only one JSON object in request body")
	}

	return nil
}

type (
	parameterExtractor func(name string) ([]string, bool)

	// PathGetter defines a function type for extracting path parameters from
	// HTTP requests. Implementations should return the parameter value and a
	// boolean indicating if the parameter exists.
	PathGetter func(r *http.Request, name string) (string, bool)
)

var (
	timeType            = reflect.TypeOf(time.Time{})
	durationType        = reflect.TypeOf(time.Duration(0))
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// BindOptions configures the behavior of parameter binding operations. It
// provides control over naming conventions, data parsing, and type conversion.
type BindOptions struct {
	// PathGetter specifies how to extract path parameters from requests. Only
	// used for BindPath.
	PathGetter PathGetter

	// FallbackSnakeCase controls field name resolution when no explicit tag is
	// present. If true, convert field names to snake_case. If false, it uses
	// lower-case field names.
	FallbackSnakeCase bool

	// SplitSingleCSV enables automatic splitting of comma-separated values in
	// single parameters. When true, a parameter value "a,b,c" will be split
	// into ["a", "b", "c"] for slice fields. Only applies when a single value
	// is received that contains the CSV separator.
	SplitSingleCSV bool

	// CSVSeparator specifies the character used to split CSV values. Only used
	// when SplitSingleCSV is true. Defaults to comma (',').
	CSVSeparator rune

	// DefaultTimeLayout specifies the time format for parsing time.Time fields.
	// Can be overridden per-field using `http:"time_format=..."` struct tags.
	// It defaults to time.RFC3339.
	DefaultTimeLayout string

	// EnableTextUnmarshaler enables support for types implementing
	// encoding.TextUnmarshaler. Default is true.
	EnableTextUnmarshaler bool
}

func getBindOptions(opts ...*BindOptions) BindOptions {
	if len(opts) == 0 || opts[0] == nil {
		return BindOptions{
			FallbackSnakeCase:     false,
			SplitSingleCSV:        true,
			CSVSeparator:          ',',
			DefaultTimeLayout:     time.RFC3339,
			EnableTextUnmarshaler: true,
			PathGetter:            StdPathGetter,
		}
	}

	o := *opts[0]

	// Set some default options if not set by the caller.
	if o.CSVSeparator == 0 {
		o.CSVSeparator = ','
	}
	if o.DefaultTimeLayout == "" {
		o.DefaultTimeLayout = time.RFC3339
	}
	if o.PathGetter == nil {
		o.PathGetter = StdPathGetter
	}

	return o
}

// BindQuery extracts query string parameters and binds them to a struct. It
// supports multiple values for the same parameter name, which will be bound
// to slice fields.
func BindQuery(r *http.Request, target interface{}, opts ...*BindOptions) error {
	var (
		o = getBindOptions(opts...)
		q = r.URL.Query()
	)

	return bindParameters(target, &o, func(name string) ([]string, bool) {
		v, ok := valuesLookup(q, name)
		return v, ok
	})
}

func valuesLookup(q url.Values, name string) ([]string, bool) {
	if v, ok := q[name]; ok && len(v) > 0 {
		return v, true
	}

	if s := q.Get(name); s != "" {
		return []string{s}, true
	}

	return nil, false
}

// BindHeader extracts HTTP headers and binds them to a struct. Header names are
// case-insensitive as per HTTP specification.
func BindHeader(r *http.Request, target interface{}, opts ...*BindOptions) error {
	var (
		o = getBindOptions(opts...)
		h = r.Header
	)

	return bindParameters(target, &o, func(name string) ([]string, bool) {
		if v := h.Values(name); len(v) > 0 {
			return v, true
		}

		return nil, false
	})
}

// BindPath extracts URL path parameters and binds them to a struct.
func BindPath(r *http.Request, target interface{}, opts ...*BindOptions) error {
	o := getBindOptions(opts...)

	return bindParameters(target, &o, func(name string) ([]string, bool) {
		if v, ok := o.PathGetter(r, name); ok {
			return []string{v}, true
		}

		return nil, false
	})
}

func bindParameters(target interface{}, opt *BindOptions, extractor parameterExtractor) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return errors.New("target must be a pointer to a struct")
	}

	var (
		rv = v.Elem()
		rt = rv.Type()
	)

	for i := 0; i < rt.NumField(); i++ {
		var (
			sf = rt.Field(i)
			fv = rv.Field(i)
		)

		if !fv.CanSet() {
			continue // unexported
		}

		name, ok := resolveFieldName(sf, opt.FallbackSnakeCase)
		if !ok {
			continue // e.g. json:"-"
		}

		values, ok := extractor(name)
		if !ok || len(values) == 0 {
			continue
		}

		if err := setFieldValues(fv, sf, values, opt); err != nil {
			return err
		}
	}

	return nil
}

func resolveFieldName(sf reflect.StructField, useSnakeCase bool) (string, bool) {
	tag := sf.Tag.Get("json")
	if tag == "-" {
		// Manually skip
		return "", false
	}
	if tag != "" {
		parts := strings.Split(tag, ",")
		return parts[0], true
	}

	if useSnakeCase {
		return strcase.SnakeCase(sf.Name), true
	}

	return strings.ToLower(sf.Name), true
}

func setFieldValues(field reflect.Value, sf reflect.StructField, values []string, opt *BindOptions) error {
	// pointers
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			// Allocate a new one
			field.Set(reflect.New(field.Type().Elem()))
		}

		return setFieldValues(field.Elem(), sf, values, opt)
	}

	// slices
	if field.Kind() == reflect.Slice {
		if opt.SplitSingleCSV && len(values) == 1 && strings.ContainsRune(values[0], opt.CSVSeparator) {
			values = stringsSplitAndTrimRune(values[0], opt.CSVSeparator)
		}

		var (
			elem = field.Type().Elem()
			out  = reflect.MakeSlice(field.Type(), 0, len(values))
		)

		for _, s := range values {
			ev := reflect.New(elem).Elem()
			if err := setScalarValue(ev, sf, s, opt); err != nil {
				return err
			}
			out = reflect.Append(out, ev)
		}

		field.Set(out)
		return nil
	}

	// scalar
	if len(values) > 0 {
		return setScalarValue(field, sf, values[0], opt)
	}

	return nil
}

func stringsSplitAndTrimRune(s string, sep rune) []string {
	var (
		parts  = strings.Split(s, string(sep))
		result = make([]string, 0, len(parts))
	)

	for _, part := range parts {
		if s := strings.TrimSpace(part); s != "" {
			result = append(result, s)
		}
	}

	return result
}

func setScalarValue(field reflect.Value, sf reflect.StructField, value string, opt *BindOptions) error {
	// encoding.TextUnmarshaler
	if opt.EnableTextUnmarshaler && field.CanAddr() && field.Addr().Type().Implements(textUnmarshalerType) {
		return field.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(value))
	}

	// time.Duration
	if field.Type() == durationType {
		return setScalarDurationField(field, value)
	}

	// time.Time
	if field.Type() == timeType {
		return setScalarTimeField(field, sf, value, opt)
	}

	return setScalarField(field, value)
}

func setScalarDurationField(field reflect.Value, value string) error {
	d, err := time.ParseDuration(value)
	if err != nil {
		return err
	}

	field.SetInt(int64(d))
	return nil
}

func setScalarTimeField(field reflect.Value, sf reflect.StructField, value string, opt *BindOptions) error {
	tag, err := parseBindTag(sf.Tag)
	if err != nil {
		return err
	}
	layout := opt.DefaultTimeLayout
	if tag != nil && tag.TimeFormat != "" {
		layout = tag.TimeFormat
	}

	t, err := time.Parse(layout, value)
	if err != nil {
		return err
	}

	field.Set(reflect.ValueOf(t))
	return nil
}

func setScalarField(field reflect.Value, value string) error {
	switch field.Kind() {
	case reflect.String:
		field.SetString(value)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetInt(i)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(value, 10, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetUint(u)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(value, field.Type().Bits())
		if err != nil {
			return err
		}
		field.SetFloat(f)

	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		field.SetBool(b)

	default:
		return errors.New("unsupported field type: " + field.Type().String())
	}

	return nil
}
