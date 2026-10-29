package handler

import (
	"errors"

	domain_errors "github.com/dyingvoid/shorturl/analytics/internal/domain/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func mapDomainError(err error) error {
	if err == nil {
		return nil
	}

	var code codes.Code
	switch {
	case errors.Is(err, domain_errors.ErrInvalidArgument):
		code = codes.InvalidArgument
	case errors.Is(err, domain_errors.ErrNotFound):
		code = codes.NotFound
	case errors.Is(err, domain_errors.ErrPermissionDenied):
		code = codes.PermissionDenied
	case errors.Is(err, domain_errors.ErrUnavailable):
		code = codes.Unavailable
	default:
		code = codes.Internal
	}

	return status.Error(code, err.Error())
}
