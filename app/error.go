package app

import (
	"errors"
	"fmt"

	"github.com/c13n-io/c13n-backend/lnchat"
	"github.com/c13n-io/c13n-backend/store"
)

// ErrKind defines the kind of an error.
type ErrKind int

// Declaration of error kind enumeration.
const (
	MarshalError ErrKind = iota
	Cancelled
	DeadlineExceeded
	NetworkError
	PermissionError
	NoRouteFound
	InvalidAddress
	InsufficientBalance
	ContactAlreadyExists
	ContactNotFound
	DiscussionAlreadyExists
	DiscussionNotFound
	UnknownError
	InternalError
)

func kindFromErr(err error) ErrKind {
	switch {
	case errors.Is(err, lnchat.ErrCancelled):
		return Cancelled
	case errors.Is(err, lnchat.ErrDeadlineExceeded):
		return DeadlineExceeded
	case errors.Is(err, lnchat.ErrPermissionDenied):
		return PermissionError
	case errors.Is(err, lnchat.ErrNetworkUnavailable):
		return NetworkError
	case errors.Is(err, lnchat.ErrUnknown):
		return UnknownError
	case errors.Is(err, lnchat.ErrInternal):
		return InternalError
	case errors.Is(err, lnchat.ErrNoRouteFound):
		return NoRouteFound
	case errors.Is(err, lnchat.ErrInvalidAddress):
		return InvalidAddress
	case errors.Is(err, lnchat.ErrInsufficientBalance):
		return InsufficientBalance
	case errors.Is(err, store.ErrContactNotFound):
		return ContactNotFound
	case errors.Is(err, store.ErrContactAlreadyExists):
		return ContactAlreadyExists
	case errors.Is(err, store.ErrDiscussionNotFound):
		return DiscussionNotFound
	case errors.Is(err, store.ErrDiscussionAlreadyExists):
		return DiscussionAlreadyExists
	default:
		return InternalError
	}
}

// Error represents an application error and contains error details.
type Error struct {
	Kind    ErrKind
	details string
	Err     error
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: %s", e.details, e.Err)
}

// Unwrap returns the underlying error of an Error.
func (e Error) Unwrap() error {
	return e.Err
}

func newErrorf(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}

	return Error{
		Kind:    kindFromErr(err),
		Err:     err,
		details: fmt.Sprintf(format, args...),
	}
}
