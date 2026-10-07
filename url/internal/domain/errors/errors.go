package domain_errors

import "errors"

var (
	ErrUnavailable       = errors.New("unavailable")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrNotFound          = errors.New("not found")
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrUniqueViolation   = errors.New("unique violation")
)
