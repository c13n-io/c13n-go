package errors

import (
	"fmt"
	"strings"
)

// Declaration of error type enumeration.
const (
	UnknownError = iota
	InternalError
	ExternalError
	NetworkError
	DatabaseError
	ApplicationError
	OtherError
)

// Error represents an error.
type Error struct {
	Cause      error
	Code       int
	Desc       string
	StackTrace []string
}

func (e Error) Error() string {
	stackString := strings.Join(e.StackTrace, "\n")
	if e.Cause == nil {
		return fmt.Sprintf("Error %v: %v\nStackTrace:\n%v",
			e.Code, e.Desc, stackString)
	}

	return fmt.Sprintf("Error %v: %v\nExternal Error:\n%+v\nStackTrace:\n%v",
		e.Code, e.Desc, e.Cause, stackString)
}

// NewError constructs a new error with a specific error code and description.
// Automatically generates a stack trace at the point it was called.
func NewError(code int, desc string) Error {
	l := getStackTrace()
	return Error{nil, code, desc, l}
}

// NewErrorWithCause constructs a new error with a specific error code,description and an
// external error reference.
// Automatically generates a stack trace at the point it was called.
func NewErrorWithCause(code int, desc string, cause error) Error {
	l := getStackTrace()
	return Error{cause, code, desc, l}
}
