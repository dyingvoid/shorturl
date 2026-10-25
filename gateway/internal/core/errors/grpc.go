package core_errors

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mappedError struct {
	message  string
	sentinel error
}

func (e mappedError) Error() string {
	return e.message
}

func (e mappedError) Unwrap() error {
	return e.sentinel
}

func FromGRPC(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	var sentinel error
	switch st.Code() {
	case codes.InvalidArgument:
		sentinel = ErrInvalidArgument
	case codes.NotFound:
		sentinel = ErrNotFound
	case codes.AlreadyExists, codes.Aborted:
		sentinel = ErrConflict
	case codes.Unauthenticated:
		sentinel = ErrUnauthenticated
	case codes.PermissionDenied:
		sentinel = ErrPermissionDenied
	case codes.ResourceExhausted:
		sentinel = ErrResourceExhausted
	case codes.DeadlineExceeded:
		sentinel = ErrDeadlineExceeded
	case codes.Unavailable:
		sentinel = ErrUnavailable
	default:
		return err
	}

	return mappedError{message: st.Message(), sentinel: sentinel}
}
