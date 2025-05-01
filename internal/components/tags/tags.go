package tags

import (
	"reflect"
	"strings"
)

const (
	tagKeyName = "mikros"
)

type Tag struct {
	IsFeature      bool
	IsOptional     bool
	IsDefinitions  bool
	GrpcClientName string
}

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
