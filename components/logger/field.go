package logger

// Field is a helper object that implements the loggerApi.Attribute interface
// allowing services to add more information into their log messages.
type Field struct {
	key   string
	value interface{}
}

// String wraps a string into a formatted log string field.
func String(key, value string) Field {
	return Field{
		key:   key,
		value: value,
	}
}

// Int32 wraps an int32 value into a formatted log string field.
func Int32(key string, value int32) Field {
	return Field{
		key:   key,
		value: value,
	}
}

// Any wraps a value into a formatted log string field.
func Any(key string, value interface{}) Field {
	return Field{
		key:   key,
		value: value,
	}
}

// Error wraps an error into a formatted log string field.
func Error(err error) Field {
	return Field{
		key:   "error.message",
		value: err.Error(),
	}
}

// Key returns the key of the Field instance as a string.
func (f Field) Key() string {
	return f.key
}

// Value returns the underlying value associated with the Field instance.
func (f Field) Value() interface{} {
	return f.value
}
