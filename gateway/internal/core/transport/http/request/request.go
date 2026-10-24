package core_http_request

import (
	"encoding/json"
	"fmt"
	"net/http"

	core_errors "github.com/dyingvoid/shorturl/gateway/internal/core/errors"
)

type validatable interface {
	Validate() error
}

func DecodeAndValidate(r *http.Request, dst any) error {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return fmt.Errorf("decode JSON body: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	value, ok := dst.(validatable)
	if !ok {
		return nil
	}

	if err := value.Validate(); err != nil {
		return fmt.Errorf("validate request: %v: %w", err, core_errors.ErrInvalidArgument)
	}

	return nil
}
