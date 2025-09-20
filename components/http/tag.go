package http

import (
	"errors"
	"reflect"
	"slices"
	"strings"
)

type bindTag struct {
	Location   string
	TimeFormat string
}

func parseBindTag(tag reflect.StructTag) (*bindTag, error) {
	raw, ok := tag.Lookup("http")
	if !ok {
		// No tag, skip field
		return nil, nil
	}

	var (
		entries = strings.Split(raw, ",")
		t       = &bindTag{}
	)
	if len(entries) == 0 || strings.TrimSpace(entries[0]) == "" {
		return nil, errors.New("http tag cannot be empty")
	}

	for _, entry := range entries {
		k, v, ok := strings.Cut(strings.TrimSpace(entry), "=")
		k = strings.TrimSpace(k)

		switch k {
		case "loc":
			if !ok {
				return nil, errors.New("http: missing member location")
			}
			if !slices.Contains([]string{"query", "header", "path", "body"}, v) {
				return nil, errors.New("http: invalid location")
			}
			t.Location = strings.TrimSpace(v)

		case "time_format":
			if !ok {
				return nil, errors.New("http: missing member time_format")
			}
			t.TimeFormat = strings.TrimSpace(v)
		}
	}

	return t, nil
}
