package env

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/mikros-dev/mikros/components/definition"
	"github.com/mikros-dev/mikros/components/service"
)

// Env is a type that holds information about a single environment variable.
// It can give both the environment variable source name and its loaded
// value.
type Env[T any] struct {
	value   T
	varName string
}

type envTag struct {
	SkipField    bool
	Name         string
	DefaultValue string
}

// Load loads environment variables directly into structures by using
// a special struct tag 'env'.
//
// The 'env' tag allows defining the environment variable name that points
// to the struct member by using the tag first argument. It also allows
// setting a default value, with the attribute 'default_value', for variables
// that are not found, and an attribute 'skip' to skip members of being
// loaded.
//
// The function also receives the service name as argument so it can search
// for environment variables specific to the own service, and to be one
// specific to the service, the environment variable must end (have as suffix)
// a dot '.' followed by the service name. Like:
//
//	FOO => global variable, for all services
//	FOO.bar => variable only for service 'bar'
func Load(serviceName service.Name, env interface{}) error {
	var (
		typeOf  = reflect.TypeOf(env)
		valueOf = reflect.ValueOf(env)
	)

	for i := 0; i < typeOf.Elem().NumField(); i++ {
		typeField := typeOf.Elem().Field(i)
		tag, err := parseFieldTag(typeField.Tag)
		if err != nil {
			return fmt.Errorf("'%s':%w", typeField.Name, err)
		}

		if tag.SkipField {
			continue
		}

		fieldValue, err := loadFieldValue(typeField, tag, serviceName)
		if err != nil {
			return err
		}

		ptr := reflect.New(fieldValue.Type())
		ptr.Elem().Set(fieldValue)
		valueOf.Elem().Field(i).Set(ptr.Elem())
	}

	return nil
}

func parseFieldTag(tag reflect.StructTag) (*envTag, error) {
	t, ok := tag.Lookup("env")
	if !ok {
		return nil, errors.New("field does not have an 'env' tag")
	}

	entries := strings.Split(t, ",")
	if len(entries) == 0 {
		return nil, errors.New("'env' tag cannot be empty")
	}

	parsedTag := &envTag{
		Name: entries[0],
	}

	for _, entry := range entries[1:] {
		parts := strings.Split(entry, "=")
		switch parts[0] {
		case "default_value":
			parsedTag.DefaultValue = parts[1]
		case "skip":
			parsedTag.SkipField = true
		}
	}

	return parsedTag, nil
}

func loadFieldValue(field reflect.StructField, tag *envTag, serviceName service.Name) (reflect.Value, error) {
	switch field.Type.Kind() {
	case reflect.Bool:
		return loadBoolFieldValue(tag, serviceName)

	case reflect.Int32:
		return loadInt32FieldValue(field.Name, tag, serviceName)

	case reflect.String:
		return loadStringFieldValue(tag, serviceName)

	case reflect.Struct:
		return loadStructFieldValue(tag, serviceName, field.Type.String())

	default:
	}

	return reflect.Value{}, fmt.Errorf("unsupported type (%s) for field '%s'", field.Type.Kind().String(), field.Name)
}

func loadBoolFieldValue(tag *envTag, serviceName service.Name) (reflect.Value, error) {
	v := getEnv(serviceName, tag.Name, tag.DefaultValue)
	if v != "true" && v != "false" {
		return reflect.Value{}, fmt.Errorf("unsupported value '%s' of bool field", v)
	}

	boolValue, err := strconv.ParseBool(v)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(boolValue), nil
}

func loadInt32FieldValue(fieldName string, tag *envTag, serviceName service.Name) (reflect.Value, error) {
	v := getEnv(serviceName, tag.Name, tag.DefaultValue)

	// Handle special cases
	if fieldName == "DeploymentEnv" {
		env := definition.ServiceDeploy.FromString(definition.ServiceDeploy(0), v)
		return reflect.ValueOf(env), nil
	}

	intValue, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return reflect.Value{}, err
	}

	return reflect.ValueOf(int32(intValue)), nil
}

func loadStringFieldValue(tag *envTag, serviceName service.Name) (reflect.Value, error) {
	v := getEnv(serviceName, tag.Name, tag.DefaultValue)
	return reflect.ValueOf(v), nil
}

func loadStructFieldValue(tag *envTag, serviceName service.Name, fieldType string) (reflect.Value, error) {
	v := getEnv(serviceName, tag.Name, tag.DefaultValue)

	if strings.Contains(fieldType, "[string]") {
		e := Env[string]{
			value:   v,
			varName: tag.Name,
		}

		return reflect.ValueOf(e), nil
	}

	if strings.Contains(fieldType, "[int32]") {
		intValue, err := strconv.ParseInt(v, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}

		e := Env[int32]{
			value:   int32(intValue),
			varName: tag.Name,
		}

		return reflect.ValueOf(e), nil
	}

	return reflect.Value{}, fmt.Errorf("unsupported Env '%v'", fieldType)
}

// getEnv is a helper function to load environment variables using priority key
// rules, by first checking if key.service_name exists before checking if key
// exists.
func getEnv(serviceName service.Name, key, defaultValue string) string {
	v := getEnvOrDefault(fmt.Sprintf("%v.%v", key, serviceName.String()), "")
	if v == "" {
		v = getEnvOrDefault(key, defaultValue)
	}

	return v
}

// getEnvOrDefault returns an environment variable with no fatal error with the
// possibility of using a default value if the variable does not exist.
func getEnvOrDefault(name string, defaultValue string) string {
	value := os.Getenv(name)

	if defaultValue != "" && value == "" {
		return defaultValue
	}

	return value
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
