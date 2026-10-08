package domain_errors

import "errors"

var (
	ErrUnavailable       = errors.New("unavailable")
	ErrInvalidArgument   = errors.New("invalid argument")
	ErrNotFound          = errors.New("not found")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrResourceExhausted = errors.New("resource exhausted")
	ErrUniqueViolation   = errors.New("unique violation")
)
