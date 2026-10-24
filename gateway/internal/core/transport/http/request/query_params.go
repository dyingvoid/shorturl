package core_http_request

import (
	"fmt"
	"net/http"
	"strconv"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
)

func QueryString(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

func QueryInt32(r *http.Request, key string) (*int32, error) {
	raw := QueryString(r, key)
	if raw == "" {
		return nil, nil
	}

	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("query parameter %q is not an integer: %v: %w", key, err, core_errors.ErrInvalidArgument)
	}

	parsed := int32(value)
	return &parsed, nil
}
