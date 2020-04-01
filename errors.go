package godac

import "fmt"

// Error define an error, that have an error code.
type Error struct {
	Text string
	Code int
}

func (e *Error) Error() string {
	return e.Text
}

// NewError create an Error.
func NewError(text string, code int) *Error {
	return &Error{text, code}
}

// Errorf format the given error text, and return a new error.
func Errorf(e *Error, a ...interface{}) error {
	return NewError(fmt.Sprintf(e.Text, a...), e.Code)
}

const (
	errCodeUser     = 400
	errCodeInternal = 500
)

// Define errors, used for customize the error text and code, %s represents the Field.
var (
	ErrValidation = NewError("%s %v", errCodeUser) // %v represents error, %s must be in front of the %v.
)
