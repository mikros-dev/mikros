package tags

import (
	"reflect"
	"strings"
)

const (
	tagKeyName = "mikros"
)

// Tag represents metadata about a structure field inside the main service
// structure.
type Tag struct {
	// IsFeature indicates if the field is tied to a feature-related tag.
	IsFeature bool

	// IsOptional denotes if the field should be skipped during parsing.
	IsOptional bool

	// IsDefinitions specifies if the tag is related to definitions.
	IsDefinitions bool

	// GrpcClientName stores the name associated with a gRPC client tag.
	GrpcClientName string
}

// ParseTag parses a struct tag and extracts metadata into a Tag object.
func ParseTag(tag reflect.StructTag) *Tag {
	t, ok := tag.Lookup(tagKeyName)
	if !ok {
		return nil
	}

	parsedTag := &Tag{}
	for _, entry := range strings.Split(t, ",") {
		parts := strings.Split(entry, "=")
		switch parts[0] {
		case "skip":
			parsedTag.IsOptional = true
		case "grpc_client":
			parsedTag.GrpcClientName = parts[1]
		case "feature":
			parsedTag.IsFeature = true
		case "definitions":
			parsedTag.IsDefinitions = true
		}
	}

	return parsedTag
}

// IsClientTag checks if the current tag is a gRPC client tag.
func (tag *Tag) IsClientTag() bool {
	return !tag.IsOptional && !tag.IsFeature && tag.GrpcClientName != ""
}
