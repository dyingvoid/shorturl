package handler

import (
	"errors"

	domain_errors "github.com/dyingvoid/shorturl/user/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapDomainError(err error) error {
	var code codes.Code
	switch {
	case errors.Is(err, domain_errors.ErrAlreadyExists):
		code = codes.AlreadyExists
	case errors.Is(err, domain_errors.ErrInvalidArgument):
		code = codes.InvalidArgument
	case errors.Is(err, domain_errors.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, domain_errors.ErrUnauthenticated):
		code = codes.Unauthenticated
	default:
		code = codes.Internal
	}

	return status.Error(code, err.Error())
}
