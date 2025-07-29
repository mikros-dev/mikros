package service

import (
	"github.com/stoewer/go-strcase"
)

type Name string

// String is a helper function to avoid casting the Name around.
func (n Name) String() string {
	return string(n)
}

// ToDatabase returns the service name in the format that should be used by
// the database name of a service.
func (n Name) ToDatabase() string {
	return strcase.SnakeCase(string(n))
}

// FromString is the official framework way to retrieve/know a service name assuring
// that it is in the right/supported format.
func FromString(serviceName string) Name {
	return Name(strcase.KebabCase(serviceName))
}
