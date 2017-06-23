// Package e provides error types and handling.
// Patterns are inspired by the following article:
// https://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
package e

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type errFull struct {
	error
}

// ErrFull wraps an error such that it will fulfill IsFull.
func ErrFull(err error) error {
	return &errFull{err}
}

// Full signals that this error indicates something was full.
func (e *errFull) Full() {}

type errNotFound struct {
	error
}

// ErrNotFound wraps an error such that it will fulfill IsNotFound.
func ErrNotFound(err error) error {
	return &errNotFound{err}
}

// NotFound signals that this error indicates something was not found.
func (e *errNotFound) NotFound() {}

type errInvalid struct {
	error
}

// ErrInvalid wraps an error such that it will fulfill IsInvalid.
func ErrInvalid(err error) error {
	return &errInvalid{err}
}

// Invalid signals that this error indicates an input was invalid.
func (e *errInvalid) Invalid() {}

// IsNotFound determines whether an error indicates something was not found.
// It does this by walking down the stack of errors built by pkg/errors and
// returning true for the first error that implements the following interface:
//
// type notfounder interface {
//   NotFound()
// }
func IsNotFound(err error) bool {
	for {
		if _, ok := err.(interface {
			NotFound()
		}); ok {
			return true
		}
		if c, ok := err.(interface {
			Cause() error
		}); ok {
			err = c.Cause()
			continue
		}
		return false
	}
}

// IsFull determines whether an error indicates something was full.
// It does this by walking down the stack of errors built by pkg/errors and
// returning true for the first error that implements the following interface:
//
// type fuller interface {
//   Full()
// }
func IsFull(err error) bool {
	for {
		if _, ok := err.(interface {
			Full()
		}); ok {
			return true
		}
		if c, ok := err.(interface {
			Cause() error
		}); ok {
			err = c.Cause()
			continue
		}
		return false
	}
}

// IsInvalid determines whether an error indicates an input was invalid.
// It does this by walking down the stack of errors built by pkg/errors and
// returning true for the first error that implements the following interface:
//
// type invalider interface {
//   Invalid()
// }
func IsInvalid(err error) bool {
	for {
		if _, ok := err.(interface {
			Invalid()
		}); ok {
			return true
		}
		if c, ok := err.(interface {
			Cause() error
		}); ok {
			err = c.Cause()
			continue
		}
		return false
	}
}

// GRPC annotates an error with the appropriate gRPC status code based on the
// error interfaces it fulfills.
func GRPC(err error) error {
	switch {
	case err == nil:
		return nil
	case IsNotFound(err):
		return status.Error(codes.NotFound, err.Error())
	case IsFull(err):
		return status.Error(codes.ResourceExhausted, err.Error())
	case IsInvalid(err):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Unknown, err.Error())
	}
}
