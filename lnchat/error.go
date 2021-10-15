package lnchat

import (
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	// ErrNoRouteFound signifies that no route
	// was found for the provided destination address.
	ErrNoRouteFound = fmt.Errorf("No route found")
	// ErrInvalidAddress signifies that the provided
	// address is invalid.
	ErrInvalidAddress = fmt.Errorf("Invalid address")
	// ErrInsufficientBalance is returned when a payment fails
	// due to insufficient balance.
	ErrInsufficientBalance = fmt.Errorf("Insufficient balance")

	// ErrCancelled is returned when a grpc call returns
	// with code Canceled.
	ErrCancelled = fmt.Errorf("Operation canceled")
	// ErrDeadlineExceeded is returned when a grpc call returns
	// with code DeadlineExceeded.
	ErrDeadlineExceeded = fmt.Errorf("Operation deadline exceeded")

	// ErrUnknown is returned when the error cause cannot be discerned.
	ErrUnknown = fmt.Errorf("Unknown error")
	// ErrInternal is returned when an error was unexpected.
	ErrInternal = fmt.Errorf("Internal error")

	// ErrNetworkUnavailable is returned when a grpc call returns
	// with code Unavailable.
	ErrNetworkUnavailable = fmt.Errorf("Network unavailable")
	// ErrCredentials signifies an error while creating Lightning credentials.
	ErrCredentials = fmt.Errorf("Credential error")
	// ErrPermissionDenied is returned when a grpc call returns
	// with code PermissionDenied.
	ErrPermissionDenied = fmt.Errorf("Permission denied")
)

// Error represents an error of the lnchat package.
type Error struct {
	Err   error
	cause error
}

func (e Error) Error() string {
	return fmt.Sprintf("%v", e.Err)
}

// Unwrap returns the underlying error of an Error.
func (e Error) Unwrap() error {
	return e.Err
}

// Cause returns the cause of an Error.
func (e Error) Cause() error {
	return e.cause
}

func withCause(err error, cause error) error {
	var t Error
	if errors.As(err, &t) {
		t.cause = cause
		return t
	}
	return err
}

func newError(code error) error {
	return Error{Err: code}
}

func newErrorf(code error, format string, args ...interface{}) error {
	return Error{
		Err: fmt.Errorf("%w: %s", code, fmt.Sprintf(format, args...)),
	}
}

func translateCommonRPCErrors(cause error) error {
	if errStatus, ok := status.FromError(cause); ok {
		switch errStatus.Code() {
		case codes.Canceled:
			return withCause(newError(ErrCancelled), cause)
		case codes.DeadlineExceeded:
			return withCause(newError(ErrDeadlineExceeded), cause)
		case codes.Unavailable:
			return withCause(newError(ErrNetworkUnavailable), cause)
		}
	}
	return cause
}

func interceptRPCError(cause error, code error) error {
	if errStatus, ok := status.FromError(cause); ok {
		if errStatus.Code() == codes.Unknown {
			return withCause(newErrorf(code, errStatus.Message()), cause)
		}
	}
	return withCause(newError(ErrInternal), cause)
}
