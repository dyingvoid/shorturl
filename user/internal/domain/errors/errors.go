package domain_errors

import "errors"

var (
	ErrAlreadyExists   = errors.New("already exists")
	ErrCacheMiss       = errors.New("cache miss")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound        = errors.New("not found")
	ErrUnauthenticated = errors.New("unauthenticated")
)
