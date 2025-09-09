package testutils

import "errors"

// NewError creates a new error with the given message
func NewError(message string) error {
	return errors.New(message)
}