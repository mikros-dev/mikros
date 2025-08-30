package env

import (
	"encoding"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/mikros-dev/mikros/components/service"
)

const (
	separator = "__"
)

var (
	errorNilTarget       = errors.New("env: nil target")
	errorNonPtrTarget    = errors.New("env: target must be a non-nil pointer to struct")
	errorNonStructTarget = errors.New("env: target must point to a struct")
	errorNoTagName       = errors.New("'env' tag cannot be empty")
	errorDefaultValue    = errors.New("default_value requires a value")
	errorPointerField    = errors.New("env: pointer-typed fields are not supported; use value type or Env[T]")

	envStringType = reflect.TypeOf(Env[string]{})
	envInt32Type  = reflect.TypeOf(Env[int32]{})

	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
	timeDurationType    = reflect.TypeOf(time.Duration(0))
)

type Options struct {
	Separator string
}

// Env is a type that wraps an environment-backed value, exposing both its value
// and the concrete env var name used to populate it.
type Env[T any] struct {
	value   T
	varName string
}

func (e Env[T]) Value() T {
	return e.value
}

func (e Env[T]) String() string {
	return fmt.Sprintf("%v", e.value)
}

func (e Env[T]) VarName() string {
	return e.varName
}

type envTag struct {
	Required     bool
	Name         string
	DefaultValue string
}

// Load populates a struct from environment variables.
//
// Precedence:
//  1. SERVICE<sep>KEY
//  2. KEY
//
// Example: if service is "file", default sep is "__":
//
//	file__DB_HOST → DB_HOST
func Load(serviceName service.Name, target interface{}, options ...Options) error {
	rv, rt, err := validateTarget(target)
	if err != nil {
		return err
	}

	opt := Options{
		Separator: separator,
	}
	if len(options) > 0 {
		opt = options[0]
	}

	for i := 0; i < rv.NumField(); i++ {
		var (
			f  = rt.Field(i)
			fv = rv.Field(i)
		)

		if !fv.CanSet() {
			continue
		}

		tag, err := parseFieldTag(f.Tag)
		if err != nil {
			return err
		}
		if tag == nil {
			// No tag, skip field
			continue
		}

		// Reject tagged pointer types
		if f.Type.Kind() == reflect.Ptr {
			return fmt.Errorf("%w: %q", errorPointerField, f.Name)
		}

		value, key, ok := resolveEnv(serviceName, tag, opt)
		if tag.Required && !ok && tag.DefaultValue == "" {
			return fmt.Errorf("env: required env %q not set", tag.Name)
		}
		// If not found and no default, leave zero value — except Env[T], which
		// we still populate to capture VarName.
		if !ok && tag.DefaultValue == "" {
			if isEnvWrapperType(f.Type) {
				v, err := zeroEnvWrapperValue(f.Type, key)
				if err != nil {
					return err
				}

				assignField(fv, v)
			}

			continue
		}

		v, err := coerceValue(f, value, key)
		if err != nil {
			return err
		}

		assignField(fv, v)
	}

	return nil
}

func validateTarget(target interface{}) (reflect.Value, reflect.Type, error) {
	if target == nil {
		return reflect.Value{}, nil, errorNilTarget
	}

	rv := reflect.ValueOf(target)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return reflect.Value{}, nil, errorNonPtrTarget
	}

	rv = rv.Elem()
	rt := rv.Type()
	if rt.Kind() != reflect.Struct {
		return reflect.Value{}, nil, errorNonStructTarget
	}

	return rv, rt, nil
}

func parseFieldTag(tag reflect.StructTag) (*envTag, error) {
	raw, ok := tag.Lookup("env")
	if !ok {
		// Members without a tag will be skipped
		return nil, nil
	}

	entries := strings.Split(raw, ",")
	if len(entries) == 0 || strings.TrimSpace(entries[0]) == "" {
		return nil, errorNoTagName
	}

	t := &envTag{
		Name: strings.TrimSpace(entries[0]),
	}

	for _, entry := range entries[1:] {
		k, v, ok := strings.Cut(strings.TrimSpace(entry), "=")
		k = strings.TrimSpace(k)

		switch k {
		case "required":
			t.Required = true
		case "":
			continue
		case "default_value":
			if !ok {
				return nil, errorDefaultValue
			}

			t.DefaultValue = trimQuotes(strings.TrimSpace(v))
		}
	}

	return t, nil
}

func trimQuotes(s string) string {
	if len(s) < 2 {
		return s
	}

	if s[0] != '"' || s[len(s)-1] != '"' {
		return s
	}

	return s[1 : len(s)-1]
}

func resolveEnv(serviceName service.Name, tag *envTag, options Options) (string, string, bool) {
	key := serviceName.String() + options.Separator + tag.Name

	if value, ok := os.LookupEnv(key); ok {
		return value, key, true
	}

	if value, ok := os.LookupEnv(tag.Name); ok {
		return value, tag.Name, true
	}

	return tag.DefaultValue, tag.Name, false
}

func isEnvWrapperType(t reflect.Type) bool {
	if t == envStringType {
		return true
	}
	if t == envInt32Type {
		return true
	}

	return false
}

func zeroEnvWrapperValue(t reflect.Type, key string) (reflect.Value, error) {
	if t == envStringType {
		return reflect.ValueOf(Env[string]{
			value:   "",
			varName: key,
		}), nil
	}
	if t == envInt32Type {
		return reflect.ValueOf(Env[int32]{
			value:   0,
			varName: key,
		}), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported Env wrapper type %v", t)
}

func coerceValue(sf reflect.StructField, value string, key string) (reflect.Value, error) {
	t := sf.Type

	// Check for Env[T] types
	if t == envStringType {
		return reflect.ValueOf(Env[string]{
			value:   value,
			varName: key,
		}), nil
	}
	if t == envInt32Type {
		n, err := parseInt(value, 32)
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(Env[int32]{
			value:   int32(n),
			varName: key,
		}), nil
	}

	// time.Duration
	if t == timeDurationType {
		d, err := time.ParseDuration(strings.TrimSpace(value))
		if err != nil {
			return reflect.Value{}, err
		}

		return reflect.ValueOf(d), nil
	}

	// Check if field implements UnmarshalText
	if implementsTextUnmarshaler(t) {
		return unmarshalTextValue(t, value)
	}

	// Scalar types
	switch t.Kind() {
	case reflect.String:
		return reflect.ValueOf(value), nil
	case reflect.Bool:
		return parseBool(value)
	case reflect.Int, reflect.Int32, reflect.Int64:
		return parseIntValue(value, t.Kind())
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		return parseUintValue(value, t.Kind())
	case reflect.Float32, reflect.Float64:
		return parseFloatValue(value, t.Kind())
	default:
	}

	return reflect.Value{}, fmt.Errorf("unsupported type %v", t)
}

func implementsTextUnmarshaler(t reflect.Type) bool {
	if t.Implements(textUnmarshalerType) {
		return true
	}

	pt := reflect.PointerTo(t)
	return pt.Implements(textUnmarshalerType)
}

func unmarshalTextValue(t reflect.Type, raw string) (reflect.Value, error) {
	if !reflect.PointerTo(t).Implements(textUnmarshalerType) {
		return reflect.Value{}, fmt.Errorf("%v does not implement encoding.TextUnmarshaler", t)
	}

	var (
		ptr = reflect.New(t)
		u   = ptr.Interface().(encoding.TextUnmarshaler)
	)

	if err := u.UnmarshalText([]byte(raw)); err != nil {
		return reflect.Value{}, err
	}

	return ptr.Elem(), nil
}

func parseBool(s string) (reflect.Value, error) {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(b), nil
}

func parseIntValue(s string, k reflect.Kind) (reflect.Value, error) {
	bitSize := 0
	if k == reflect.Int32 {
		bitSize = 32
	}
	if k == reflect.Int64 {
		bitSize = 64
	}

	n, err := parseInt(s, bitSize)
	if err != nil {
		return reflect.Value{}, err
	}

	if bitSize == 0 {
		return reflect.ValueOf(int(n)), nil
	}
	if bitSize == 32 {
		return reflect.ValueOf(int32(n)), nil
	}

	return reflect.ValueOf(n), nil
}

func parseInt(s string, bitSize int) (int64, error) {
	n, err := strconv.ParseInt(strings.TrimSpace(s), 10, bitSize)
	if err != nil {
		return 0, err
	}

	return n, nil
}

func parseUintValue(s string, k reflect.Kind) (reflect.Value, error) {
	bitSize := 0
	if k == reflect.Uint32 {
		bitSize = 32
	}
	if k == reflect.Uint64 {
		bitSize = 64
	}

	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, bitSize)
	if err != nil {
		return reflect.Value{}, err
	}

	if bitSize == 0 {
		return reflect.ValueOf(uint(n)), nil
	}
	if bitSize == 32 {
		return reflect.ValueOf(uint32(n)), nil
	}

	return reflect.ValueOf(n), nil
}

func parseFloatValue(s string, k reflect.Kind) (reflect.Value, error) {
	bitSize := 64
	if k == reflect.Float32 {
		bitSize = 32
	}

	f, err := strconv.ParseFloat(strings.TrimSpace(s), bitSize)
	if err != nil {
		return reflect.Value{}, err
	}

	if bitSize == 32 {
		return reflect.ValueOf(float32(f)), nil
	}

	return reflect.ValueOf(f), nil
}

func assignField(dst reflect.Value, src reflect.Value) {
	if !src.IsValid() {
		return
	}

	if src.Type().AssignableTo(dst.Type()) {
		dst.Set(src)
		return
	}

	if src.Type().ConvertibleTo(dst.Type()) {
		dst.Set(src.Convert(dst.Type()))
		return
	}
}
