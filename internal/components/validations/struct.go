// Package validations adds an internal framework API to validate structures
// without tag annotations. Despite the fact, it also has support for tag
// annotations to skip validation for specific structure members.
//
// The main usage for this API is to validate a service main structure,
// usually the place where its main API is implemented (RPCs and subscription
// handlers), automatically to avoid using uninitialized members while
// the service is running.
package validations

import (
	"errors"
	"fmt"
	"reflect"

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

		if tag := tags.ParseTag(typeField.Tag); tag != nil {
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
