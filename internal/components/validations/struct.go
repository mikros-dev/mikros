// Package validations adds an internal framework API to validate structures
// without tag annotations. Despite the fact it also has support for tag
// annotations in order to skip validation for specific structure members.
//
// The main usage for this API is to validate a service main structure,
// usually the place where its main API are implemented (RPCs and subscription
// handlers), in an automatic way to avoid using uninitialized members inside
// the service.
package validations

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/mikros-dev/mikros/internal/components/tags"
)

// EnsureValuesAreInitialized certifies that all members of a struct v have
// some valid value. It requires a struct object to be passed as argument, and
// it considers a pointer member with nil value as uninitialized.
func EnsureValuesAreInitialized(v interface{}) error {
	if v == nil {
		return errors.New("can't validate nil object")
	}

	elem := reflect.ValueOf(v)
	if reflect.TypeOf(v).Kind() == reflect.Ptr {
		elem = elem.Elem()
	}

	// checks if we're dealing with a structure or not
	if !isStruct(v) {
		return errors.New("can't validate non struct objects")
	}

	for i := 0; i < elem.NumField(); i++ {
		typeField := elem.Type().Field(i)
		valueField := elem.Field(i)

		tag := tags.ParseTag(typeField.Tag)
		if tag != nil {
			// Optional members or gRPC clients don't need to be validated.
			if tag.IsOptional || tag.GrpcClientName != "" {
				continue
			}
		}

		isNil := valueField.Kind() == reflect.Ptr && valueField.IsNil()
		if isNil || valueField.IsZero() {
			return fmt.Errorf("could not initiate struct %s, value from field %s is missing",
				elem.Type().Name(), typeField.Name,
			)
		}
	}

	return nil
}

// EnsureStructIsServiceCompatible validates if v corresponds to a structure that
// can be used by the framework as the service handler.
func EnsureStructIsServiceCompatible(v interface{}) error {
	if isNil(v) {
		return errors.New("can't validate nil object")
	}

	if !isStruct(v) {
		return errors.New("argument is not a structure")
	}

	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return errors.New("argument should be a pointer to a structure")
	}

	elem := reflect.ValueOf(v).Elem()
	for i := 0; i < elem.NumField(); i++ {
		typeField := elem.Type().Field(i)

		// The structure must have a framework Service field.
		if strings.Contains(typeField.Type.String(), "mikros.Service") {
			return nil
		}
	}

	return errors.New("could not find Service member inside structure")
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	var (
		t    = reflect.TypeOf(v)
		kind = t.Kind()
	)

	return kind == reflect.Ptr && reflect.ValueOf(v).IsNil()
}

// isStruct checks if an object is a struct object using reflection.
func isStruct(v interface{}) bool {
	var (
		t    = reflect.TypeOf(v)
		kind = t.Kind()
		ptr  = reflect.Invalid
	)

	if kind == reflect.Ptr {
		ptr = reflect.ValueOf(v).Elem().Kind()
	}

	return kind == reflect.Struct || (kind == reflect.Ptr && ptr == reflect.Struct)
}
