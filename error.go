package godac

// Error variables definition.
var (
	ErrExists = NewError("this record is in use", 400)
)

// Error formats.
var (
	ErrorFormatOnDelete = "can not be deleted, %s"
)

// Error define an error that include error code.
type Error interface {
	error
	Code() int
}

type errorObject struct {
	text string
	code int
}

func (e *errorObject) Error() string {
	return e.text
}

func (e *errorObject) Code() int {
	return e.code
}

// NewError to create an error.
func NewError(text string, code int) Error {
	return &errorObject{text, code}
}
