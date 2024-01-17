package errors

import (
	"fmt"
)

type AbortError struct {
	Message    string
	InnerError error
}

// NewAbortError creates an error that the API Start call must return internally.
func NewAbortError(message string, err error) *AbortError {
	return &AbortError{
		Message:    message,
		InnerError: err,
	}
}

func (s *AbortError) Error() string {
	return fmt.Sprintf("%v:%v", s.Message, s.InnerError.Error())
}
