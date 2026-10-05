package domain_errors

import "errors"

var (
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrNotFound = errors.New("not found")
	ErrUnauthenticated = errors.New("unauthenticated")
)