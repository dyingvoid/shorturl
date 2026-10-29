package domain_errors

import "errors"

var (
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("permission denied")
	ErrUnavailable      = errors.New("unavailable")
)
