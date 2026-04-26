package definition

import (
	"context"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
)

func versionValidator(_ context.Context, fl validator.FieldLevel) bool {
	return ValidateVersion(fl.Field().String())
}

// ValidateVersion is a helper function to validate the version format used by
// services.
func ValidateVersion(input string) bool {
	return regexp.MustCompile("^v[0-9]{1,2}(|[.][0-9]{1,2})(|[.][0-9]{1,2})$").MatchString(input)
}

// runtimeTypeValidator validates if a valid runtime type was used inside the
// settings file. It also supports the notation 'type:port', where one can
// set a custom server port for the specific runtime type.
func runtimeTypeValidator(ctx context.Context, fl validator.FieldLevel) bool {
	rt := fl.Field().String()
	if rt == "" {
		return true
	}

	supportedTypes, ok := ctx.Value(runtimeTypeCtx{}).([]string)
	if !ok {
		return false
	}

	if strings.Contains(rt, ":") {
		parts := strings.Split(rt, ":")
		if len(parts) > 1 {
			// The server port was defined and we must validate it.
			if !validatePort(parts[1]) {
				return false
			}
		}

		rt = parts[0]
	}

	for _, t := range supportedTypes {
		if rt == t {
			return true
		}
	}

	return false
}

func validatePort(port string) bool {
	_, err := strconv.ParseInt(port, 10, 32)
	return err == nil
}

// scriptTypeUniqueValidator validates if the 'script' runtime type is alone in
// the list.
func scriptTypeUniqueValidator(_ context.Context, fl validator.FieldLevel) bool {
	if list, ok := fl.Field().Interface().([]string); ok {
		index := slices.Index(list, RuntimeTypeScript.String())
		if index != -1 && len(list) > 1 {
			return false
		}
	}

	return true
}

// duplicatedRuntimeValidator validates if the list contains duplicated elements.
func duplicatedRuntimeValidator(_ context.Context, fl validator.FieldLevel) bool {
	if list, ok := fl.Field().Interface().([]string); ok {
		types := make(map[string]bool)
		for _, t := range list {
			_, ok := types[t]
			if !ok {
				types[t] = true
			}
			if ok {
				return false
			}
		}
	}

	return true
}
