// Package domain defines core business entities, interfaces and errors
package domain

import (
	"errors"
	"fmt"
)

// ErrBadRequest returns a new error indicating an invalid client request
func ErrBadRequest(msg string) error {
	return &badRequestError{msg: msg}
}

// ErrBadRequestf returns a formatted bad request error
func ErrBadRequestf(format string, args ...any) error {
	return &badRequestError{msg: fmt.Sprintf(format, args...)}
}

// badRequestError is a client-side validation error that should result in HTTP 400
type badRequestError struct {
	msg string
}

// Error returns the error message
func (e *badRequestError) Error() string {
	return e.msg
}

// IsBadRequest reports whether err is a bad request error
func IsBadRequest(err error) bool {
	var target *badRequestError
	return errors.As(err, &target)
}
