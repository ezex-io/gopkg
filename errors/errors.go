package errors

type Error struct {
	Code    int               `json:"code"`
	Message string            `json:"message"`
	Meta    map[string]string `json:"meta"`
}

// New creates a new structured Error with a code, message, and optional metadata.
// The metadata must be provided as key-value pairs (e.g., "field", "email").
// If an odd number of metaKeyVal arguments is provided, an "invalid meta" key will be added to Meta.
//
// Example:
//
//	err := errors.New(404, "User not found", "email", "foo@example.com")
func New(code int, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Meta:    make(map[string]string),
	}
}

// AddMeta add metadata to error message.
func (e *Error) AddMeta(keyVal ...string) *Error {
	if len(keyVal)%2 != 0 {
		e.Meta["error"] = "invalid meta key/value args"
	} else {
		for i := 0; i < len(keyVal); i += 2 {
			e.Meta[keyVal[i]] = keyVal[i+1]
		}
	}

	return e
}

func (e *Error) Error() string {
	return e.Message
}
