package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
	"github.com/go-chi/chi/v5"
)

func PathValue(r *http.Request, key string) (string, error) {
	value := chi.URLParam(r, key)
	if value == "" {
		return "", fmt.Errorf("missing path value %q: %w", key, core_errors.ErrInvalidArgument)
	}

	return value, nil
}

func PathInt64(r *http.Request, key string) (int64, error) {
	raw, err := PathValue(r, key)
	if err != nil {
		return 0, err
	}

	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("path value %q is not an integer: %v: %w", key, err, core_errors.ErrInvalidArgument)
	}

	return value, nil
}
